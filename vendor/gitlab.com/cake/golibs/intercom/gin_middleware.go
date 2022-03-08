package intercom

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
	"gitlab.com/cake/goctx"
	m800schema "gitlab.com/cake/golibs/intercom/schemas/v1.0.0"
	metrics "gitlab.com/cake/golibs/metric"
	"gitlab.com/cake/gopkg"
	"gitlab.com/cake/m800log"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	oteltrace "go.opentelemetry.io/otel/trace"
	"golang.org/x/sync/singleflight"
)

var (
	dunno     = []byte("???")
	centerDot = []byte("·")
	dot       = []byte(".")
	slash     = []byte("/")
)

var (
	slowReqDuration = 5 * time.Second
)

var (
	proxyMap = sync.Map{}
)

var (
	reverseProxySingleFlightGroup singleflight.Group
)

var (
	currentRequestCount uint64 = 0
	metricUnit          int    = 1
)

const (
	proxyScheme            = "http"
	proxyHeaderForwardHost = "X-Forward-Host"
	proxyHeaderOriginHost  = "X-Origin-Host"
)

const (
	LogHideWildcardName = "*"
)

var (
	accessMiddlewareTracer = otel.Tracer("gitlab.com/cake/golibs/intercom.AccessMiddleware")
	crossMiddlewareTracer  = otel.Tracer("gitlab.com/cake/golibs/intercom.CrossRegionNamespaceMiddleware")
)

type LogHideOption struct {
	HandlerName   string
	RequestHider  func(b []byte) []byte
	ResponseHider func(httpStatus int, b []byte) []byte
}

type bodyLogWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w bodyLogWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

func SetSlowReqThreshold(t time.Duration) {
	slowReqDuration = t
}

// M800Recovery does the recover for gin handler with M800 response
func M800Recovery(panicCode int) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			handlerName := "M800Recovery"
			if err := recover(); err != nil {
				// Check for a broken connection, as it is not really a condition that warrants a panic stack trace.
				var brokenPipe bool
				var errSyscall *os.SyscallError
				if ne, ok := err.(error); ok && errors.As(ne, &errSyscall) &&
					(strings.Contains(strings.ToLower(errSyscall.Error()), "broken pipe") ||
						strings.Contains(strings.ToLower(errSyscall.Error()), "connection reset by peer")) {
					brokenPipe = true
				}

				stack := stack(3)
				panicStr := fmt.Sprintf("[%s] recovered at: %s, panic err:\n%s,\nstack:\n%s",
					handlerName, timeFormat(time.Now()), err, stack)

				ctx := GetContextFromGin(c)
				ctx.Set(goctx.LogKeyErrorCode, panicCode)

				// If the connection is dead, we can't write a status to it.
				if brokenPipe {
					label := prometheus.Labels{
						labelDownstream: downstreamName(c),
					}
					brokenPipeCounts.With(label).Inc()
					c.Error(err.(error)) // nolint: errcheck
					c.Abort()
					return
				}

				m800log.Error(ctx, panicStr)
				GinErrorCodeMsg(c, panicCode, fmt.Sprintf("%s", err))
			}
		}()
		c.Next()
	}
}

func downstreamName(c *gin.Context) string {
	ctx := GetContextFromGin(c)
	if caller, ok := ctx.GetString(goctx.HTTPHeaderInternalCaller); ok {
		return caller
	}
	if strings.Contains(c.GetHeader("User-Agent"), "Go-http-client") {
		return "golang-client"
	}
	return "client"
}

// stack returns a nicely formatted stack frame, skipping skip frames.
func stack(skip int) []byte {
	buf := new(bytes.Buffer) // the returned data
	// As we loop, we open files and read them. These variables record the currently
	// loaded file.
	var lines [][]byte
	var lastFile string
	for i := skip; ; i++ { // Skip the expected number of frames
		pc, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}
		// Print this much at least.  If we can't find the source, it won't show.
		fmt.Fprintf(buf, "%s:%d (0x%x)\n", file, line, pc)
		if file != lastFile {
			data, err := ioutil.ReadFile(file)
			if err != nil {
				continue
			}
			lines = bytes.Split(data, []byte{'\n'})
			lastFile = file
		}
		fmt.Fprintf(buf, "\t%s: %s\n", function(pc), source(lines, line))
	}
	return buf.Bytes()
}

// source returns a space-trimmed slice of the n'th line.
func source(lines [][]byte, n int) []byte {
	n-- // in stack trace, lines are 1-indexed but our array is 0-indexed
	if n < 0 || n >= len(lines) {
		return dunno
	}
	return bytes.TrimSpace(lines[n])
}

// function returns, if possible, the name of the function containing the PC.
func function(pc uintptr) []byte {
	fn := runtime.FuncForPC(pc)
	if fn == nil {
		return dunno
	}
	name := []byte(fn.Name())
	// The name includes the path name to the package, which is unnecessary
	// since the file name is already included.  Plus, it has center dots.
	// That is, we see
	//	runtime/debug.*T·ptrmethod
	// and want
	//	*T.ptrmethod
	// Also the package path might contains dot (e.g. code.google.com/...),
	// so first eliminate the path prefix
	if lastSlash := bytes.LastIndex(name, slash); lastSlash >= 0 {
		name = name[lastSlash+1:]
	}
	if period := bytes.Index(name, dot); period >= 0 {
		name = name[period+1:]
	}
	name = bytes.Replace(name, centerDot, dot, -1)
	return name
}

func timeFormat(t time.Time) string {
	var timeString = t.Format(time.RFC3339Nano)
	return timeString
}

// AccessMiddleware
func AccessMiddleware(timeout time.Duration, localNamespace string, opts ...*LogHideOption) gin.HandlerFunc {
	hiderReqMap := make(map[string]func(b []byte) []byte)
	hiderRespMap := make(map[string]func(httpStatus int, b []byte) []byte)
	for _, opt := range opts {
		if opt.RequestHider != nil {
			hiderReqMap[opt.HandlerName] = opt.RequestHider
		}
		if opt.ResponseHider != nil {
			hiderRespMap[opt.HandlerName] = opt.ResponseHider
		}
	}
	return func(c *gin.Context) {
		blw := &bodyLogWriter{body: &bytes.Buffer{}, ResponseWriter: c.Writer}
		c.Writer = blw
		ctx := GetContextFromGin(c)
		ctx.Set(goctx.LogKeyHTTPMethod, c.Request.Method)
		ctx.Set(goctx.LogKeyURI, c.Request.URL.RequestURI())
		// init if no cid
		_, ok := ctx.GetString(goctx.LogKeyCID)
		if !ok {
			cid, _ := ctx.GetCID()
			c.Request.Header.Set(goctx.HTTPHeaderCID, cid)
		}

		start := time.Now().UTC()
		defer m800log.Access(ctx, start)

		cancel := ctx.SetTimeout(timeout)
		defer cancel()
		handlerName := c.HandlerName()
		ctx.Set(LogEntryHandlerName, handlerName)

		_, sp := accessMiddlewareTracer.Start(c.Request.Context(), handlerName, oteltrace.WithAttributes(m800schema.M800NamespaceKey.String(localNamespace)))
		defer sp.End()

		var httpBody []byte
		if c.Request.Body != nil {
			var err gopkg.CodeError
			httpBody, err = ReadFromReadCloser(c.Request.Body)
			if err != nil {
				GinError(c, err)
				return
			}
			// support common http req usage
			c.Request.Body = ioutil.NopCloser(bytes.NewReader(httpBody))
			c.Set(KeyBody, httpBody)
		}

		c.Next()
		select {
		case <-ctx.Done():
			m800log.Debug(ctx, "ctx done case")
		default:
			// common case
		}
		elapsed := time.Since(start)
		if elapsed > slowReqDuration {
			ctx.Set("slow", elapsed)
		}
		if elapsed > timeout {
			m800log.Errorf(ctx, "api timeout, timeout setting: %s, elapsed: %s", timeout, elapsed)
		} else if elapsed > slowReqDuration {
			m800log.Warnf(ctx, "api slow, slow setting: %s, elapsed: %s", slowReqDuration, elapsed)
		}

		if traceErrCode := c.GetInt(goctx.LogKeyErrorCode); traceErrCode != 0 {
			sp.SetStatus(codes.Error, c.GetString(goctx.LogKeyErrorMessage))
			ctx.Set(goctx.LogKeyErrorCode, traceErrCode)
			dumpLogHandle(ctx, c, handlerName, httpBody, blw, elapsed, ErrorTraceLevel, hiderReqMap, hiderRespMap)
			return
		}
		if m800log.GetLogger().Level >= logrus.DebugLevel {
			dumpLogHandle(ctx, c, handlerName, httpBody, blw, elapsed, logrus.DebugLevel, hiderReqMap, hiderRespMap)
			return
		}
	}
}

func dumpLogHandle(ctx goctx.Context, c *gin.Context, handlerName string, httpBody []byte, blw *bodyLogWriter, elapsed time.Duration, ErrorTraceLevel logrus.Level, hiderReqMap map[string]func(b []byte) []byte, hiderRespMap map[string]func(httpStatus int, b []byte) []byte) {
	strs := strings.Split(handlerName, ".")
	logHandlerName := strs[len(strs)-1]
	logReqBody, logRespBody := httpBody, blw.body.Bytes()
	if reqHider := hiderReqMap[logHandlerName]; reqHider != nil {
		logReqBody = reqHider(logReqBody)
	} else if reqHider := hiderReqMap[LogHideWildcardName]; reqHider != nil {
		logReqBody = reqHider(logReqBody)
	}
	if respHider := hiderRespMap[logHandlerName]; respHider != nil {
		logRespBody = respHider(c.Writer.Status(), logRespBody)
	} else if respHider := hiderRespMap[LogHideWildcardName]; respHider != nil {
		logRespBody = respHider(c.Writer.Status(), logRespBody)
	}
	dumpRequestGivenBody(ctx, ErrorTraceLevel, c.Request, logReqBody)
	m800log.Logf(ctx, ErrorTraceLevel, "API Response %d: duration: %s body: %s", c.Writer.Status(), elapsed, logRespBody)
}

func newProxy(ctx goctx.Context, forwardedHost string, timeout time.Duration, proxyErrorCode int) (result *httputil.ReverseProxy) {
	forgetSingleFlightGroupKey := func(key string) {
		time.Sleep(singleFlightRequestDuration)
		reverseProxySingleFlightGroup.Forget(key)
	}

	resultI, _, _ := reverseProxySingleFlightGroup.Do(forwardedHost, func() (output interface{}, err error) {
		go forgetSingleFlightGroupKey(forwardedHost)

		output, ok := proxyMap.Load(forwardedHost)
		if ok {
			return
		}

		director := func(req *http.Request) {
			req.Header.Add(proxyHeaderForwardHost, forwardedHost)
			req.Header.Add(proxyHeaderOriginHost, req.Host)
			req.URL.Scheme = proxyScheme
			req.URL.Host = forwardedHost
		}
		proxy := &httputil.ReverseProxy{
			Director: director,
			Transport: &http.Transport{
				ResponseHeaderTimeout: timeout,
			},
			ErrorHandler: func(resp http.ResponseWriter, req *http.Request, err error) {
				if err != nil {
					response := Response{}
					response.Code = proxyErrorCode
					response.Message = "cross region forward error: " + err.Error()
					byteResp, err := json.Marshal(response)
					if err != nil {
						byteResp = []byte(fmt.Sprintf(`{"code":%d,"message":"cross region forward error"}`, proxyErrorCode))
					}
					resp.Header().Set(HeaderContentType, HeaderJSON)
					resp.WriteHeader(http.StatusBadGateway)
					_, errWrite := resp.Write(byteResp)
					if errWrite != nil {
						m800log.Info(ctx, "proxy response write error: ", errWrite)
					}
				}
			},
		}

		proxyMap.Store(forwardedHost, proxy)
		output = proxy
		return
	})

	result = resultI.(*httputil.ReverseProxy)
	return
}

func CrossRegionNamespaceMiddleware(service, servicePort, localNamespace string, nsFunc func(c *gin.Context) (string, gopkg.CodeError), timeout time.Duration, proxyErrorCode int) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Edge server would carry the service home region info
		// https://issuetracking.maaii.com:9443/display/LCC5/Edge+Server+Header+Rules

		ns, err := nsFunc(c)
		if err != nil {
			errMsg := fmt.Sprintf("cross region lookup ns failed, err: %+v", err)
			GinOKError(c, gopkg.NewCodeError(proxyErrorCode, errMsg))
			c.Abort()
			return
		}
		if ns == "" {
			c.Next()
			return
		}
		if ns == localNamespace {
			c.Next()
			return
		}

		// ccg handling
		var forwardedHost string
		if ccgEnabled {
			forwardedURL := fmt.Sprintf("%s://%s.%s:%s%s", proxyScheme, service, ns, servicePort, c.Request.URL.String())
			c.Request.Header.Add(ccgForwardURL, forwardedURL)
			c.Request.URL = ccgHTTPProxyV1FullURL
			forwardedHost = ccgHTTPProxyHost
		} else {
			forwardedHost = fmt.Sprintf("%s.%s:%s", service, ns, servicePort)
		}

		// craft ctx
		ctx := GetContextFromGin(c)
		ctx.Set(goctx.LogKeyFromNamespace, localNamespace)
		cid, _ := ctx.GetCID()

		_, sp := crossMiddlewareTracer.Start(c.Request.Context(),
			"cross region request start",
			oteltrace.WithAttributes(m800schema.M800NamespaceKey.String(localNamespace)),
			oteltrace.WithAttributes(m800schema.M800CrossFromNamespaceKey.String(localNamespace)),
			oteltrace.WithAttributes(m800schema.M800CrossToNamespaceKey.String(ns)),
		)
		defer sp.End()

		propagators := otel.GetTextMapPropagator()
		propagators.Inject(c.Request.Context(), propagation.HeaderCarrier(c.Request.Header))

		m800log.Debugf(ctx, "[cross region middleware] do the cross forward to :%s, cid: %s, path: %s", forwardedHost, cid, c.Request.URL.Path)

		// prepare logger
		logWriter := m800log.GetLogger().WriterLevel(logrus.InfoLevel)
		defer logWriter.Close()

		// prepare reverse proxy
		proxy := newProxy(ctx, forwardedHost, timeout, proxyErrorCode)
		proxy.ErrorLog = log.New(logWriter, "", 0)

		// prepare metrics
		crossRegionStart := time.Now()
		reqSz := float64(metrics.ComputeApproximateRequestSize(c.Request))
		resSz := float64(0)
		nsLabel := prometheus.Labels{
			labelFromNamespace: localNamespace,
			labelToNamespace:   ns,
		}

		defer func() {
			// update metrics
			totalCrossRegionRequestCounts.With(nsLabel).Inc()

			currentCount := atomic.AddUint64(&currentRequestCount, -uint64(metricUnit))
			currentCrossRegionRequestCounts.With(nsLabel).Set(float64(currentCount))

			crossRegionElapsed := time.Since(crossRegionStart)
			crossRegionRequestDuration.With(nsLabel).Observe(float64(crossRegionElapsed / time.Second))

			crossRegionRequestSize.With(nsLabel).Observe(reqSz)

			crossRegionResponseSize.With(nsLabel).Observe(resSz)

			if err := recover(); err != nil {
				// It could because downsteam or upstream disconnets. See https://github.com/gin-gonic/gin/issues/1714
				if ne, ok := err.(error); ok && errors.Is(ne, http.ErrAbortHandler) {
					ctx.Set(goctx.LogKeyErrorCode, CodePanic)

					// update metrics
					label := prometheus.Labels{
						labelDownstream:        downstreamName(c),
						labelUpstream:          service,
						labelUpstreamNamespace: ns,
					}
					proxyBrokenPipeCounts.With(label).Inc()

					totalCrossRegionFailedRequestCounts.With(nsLabel).Inc()

					m800log.Debugf(ctx, "[cross region middleware] cross region abort panic: %+v", ne)
					return
				}
				panic(err)
			}
		}()

		// update metrics
		currentCount := atomic.AddUint64(&currentRequestCount, uint64(metricUnit))
		currentCrossRegionRequestCounts.With(nsLabel).Set(float64(currentCount))

		// start cross region
		ctx.InjectHTTPHeader(c.Request.Header)
		proxy.ServeHTTP(c.Writer, c.Request)

		// update metrics
		resSz = float64(c.Writer.Size())

		c.Abort()
	}
}

// CrossRegionMiddleware
func CrossRegionMiddleware(service, servicePort, localNamespace string, timeout time.Duration, proxyErrorCode int) gin.HandlerFunc {
	nsFunc := func(c *gin.Context) (string, gopkg.CodeError) {
		return c.GetHeader(goctx.HTTPHeaderServiceHome), nil
	}

	return CrossRegionNamespaceMiddleware(service, servicePort, localNamespace, nsFunc, timeout, proxyErrorCode)
}

func UserHomeCrossRegionMiddleware(service, servicePort, localNamespace string, timeout time.Duration, proxyErrorCode int) gin.HandlerFunc {
	nsFunc := func(c *gin.Context) (string, gopkg.CodeError) {
		return c.GetHeader(goctx.HTTPHeaderUserHome), nil
	}

	return CrossRegionNamespaceMiddleware(service, servicePort, localNamespace, nsFunc, timeout, proxyErrorCode)
}

// BanAnonymousMiddleware
func BanAnonymousMiddleware(errCode gopkg.CodeError) gin.HandlerFunc {
	return func(c *gin.Context) {
		if _, ok := c.Request.Header[http.CanonicalHeaderKey(goctx.HTTPHeaderUserAnms)]; !ok {
			c.Next()
			return
		}

		isAnms, err := strconv.ParseBool(c.GetHeader(goctx.HTTPHeaderUserAnms))
		if isAnms {
			GinError(c, errCode)
			c.Abort()
			return
		}

		if err != nil {
			GinErrorCodeMsg(c, CodeMaliciousHeader, MsgErrMaliciousHeader)
			c.Abort()
			return
		}

		c.Next()
	}
}

// utils func for hiding info in LogHideOption
func FindNextJsonStringValue(str, key string) (value string, endIndex int) {
	keyIndex := strings.Index(str, key)
	if keyIndex < 0 {
		return "", -1
	}

	if key[0] != '"' {
		key = `"` + key
	}
	if key[len(key)-1] != '"' {
		key = key + `"`
	}
	kl := len(key)

	keyNextIndex := keyIndex + kl
	nextQ1Index := strings.Index(str[keyNextIndex:], `"`)
	strNextQ1Index := keyNextIndex + nextQ1Index + 1
	nextQ2Offset := 0
	for nextQ2Offset < len(str)-strNextQ1Index {
		nextIndex := strings.Index(str[strNextQ1Index+nextQ2Offset:], `"`)
		if nextIndex < 0 {
			break
		}
		if str[strNextQ1Index+nextIndex+nextQ2Offset-1] != '\\' {
			nextQ2Offset += nextIndex
			break
		}
		nextQ2Offset += nextIndex + 1
	}
	endIndex = strNextQ1Index + nextQ2Offset
	return str[strNextQ1Index:endIndex], endIndex
}

func FindAllJsonStringValue(str, key string) (values []string) {
	for {
		value, index := FindNextJsonStringValue(str, key)
		if index < 0 {
			return
		}
		values = append(values, value)
		str = str[index:]
	}
}

// memory performance issue
func ReplaceUserTextJson(input, key string) string {
	values := FindAllJsonStringValue(input, key)
	for i := range values {
		input = strings.ReplaceAll(input, values[i], fmt.Sprintf("hidetext_len:%d", len(values[i])))
	}
	return input
}

func ReplaceUserTextHeader(input, key string) string {
	lowerStr := strings.ToLower(input)
	keyIndex := strings.Index(lowerStr, key)
	if keyIndex < 0 {
		return input
	}

	// header with ":"
	keyEndIndex := keyIndex + len(key) + 1
	newlineIndex := strings.Index(input[keyEndIndex:], "\n")
	value := input[keyEndIndex : keyEndIndex+newlineIndex]
	input = input[:keyEndIndex] + fmt.Sprintf("hidetext_len:%d", len(value)) + input[keyEndIndex+newlineIndex:]

	return input
}

// example reference to im-common/middleware.go

func HideReqB64(input []byte) []byte {
	buf := make([]byte, base64.StdEncoding.EncodedLen(len(input)))
	base64.StdEncoding.Encode(buf, input)
	return buf
}

func HideResponseB64(status int, input []byte) []byte {
	if status != http.StatusOK {
		return input
	}
	buf := make([]byte, base64.StdEncoding.EncodedLen(len(input)))
	base64.StdEncoding.Encode(buf, input)

	return buf
}
