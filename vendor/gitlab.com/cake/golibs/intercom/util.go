package intercom

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"runtime"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"gitlab.com/cake/goctx"
	"gitlab.com/cake/gopkg"
	"gitlab.com/cake/m800log"
	"go.opentelemetry.io/otel/baggage"
	"go.opentelemetry.io/otel/trace"
)

type DecodeOption struct {
	UseNumber             bool
	DisallowUnknownFields bool
}

const defaultDumpLimitLength = 16000

// ParseJSONReq read body
func ParseJSONReq(ctx goctx.Context, req *http.Request, v interface{}) gopkg.CodeError {
	if req == nil || req.Body == nil {
		return gopkg.NewCodeError(CodeNilRequest, "nil request")
	}
	d := json.NewDecoder(req.Body)
	defer req.Body.Close()
	d.UseNumber()
	if errJSON := d.Decode(v); errJSON != nil {
		m800log.Debugf(ctx, "[ParseJSONReq] err:%v", errJSON.Error())
		return gopkg.NewCodeError(CodeParseJSON, errJSON.Error())
	}
	return nil
}

// ParseJSONGin
func ParseJSONGin(ctx goctx.Context, c *gin.Context, v interface{}) gopkg.CodeError {
	return ParseJSONGinWithDecodeOption(ctx, c, v,
		DecodeOption{
			UseNumber:             true,
			DisallowUnknownFields: false,
		},
	)
}

func ParseJSONGinWithDecodeOption(ctx goctx.Context, c *gin.Context, v interface{}, option DecodeOption) gopkg.CodeError {
	rawI, ok := c.Get(KeyBody)
	if !ok {
		return ParseJSONReq(ctx, c.Request, v)
	}
	raw, ok := rawI.([]byte)
	if !ok {
		errMsg := fmt.Sprintf("http body type error:%t", rawI)
		m800log.Error(ctx, errMsg)
		return gopkg.NewCodeError(CodeParseJSON, errMsg)
	}
	d := json.NewDecoder(bytes.NewBuffer(raw))
	if option.UseNumber {
		d.UseNumber()
	}
	if option.DisallowUnknownFields {
		d.DisallowUnknownFields()
	}

	if errJSON := d.Decode(v); errJSON != nil {
		m800log.Debugf(ctx, "[ParseJSONGin] err:%v, req body: %s", errJSON.Error(), string(raw))
		return gopkg.NewCodeError(CodeParseJSON, errJSON.Error())
	}
	return nil
}

func ReadFromReadCloser(readCloser io.ReadCloser) ([]byte, gopkg.CodeError) {
	if readCloser == nil {
		return nil, gopkg.NewCodeError(CodeReadAll, "nil readCloser")
	}
	defer readCloser.Close()
	raw, err := ioutil.ReadAll(readCloser)
	if err != nil {
		return nil, gopkg.NewCodeError(CodeReadAll, err.Error())
	}

	return raw, nil
}

// ParseJSONReadCloser
func ParseJSONReadCloser(ctx goctx.Context, readCloser io.ReadCloser, v interface{}) gopkg.CodeError {
	if readCloser == nil {
		return gopkg.NewCodeError(CodeReadAll, "nil readCloser")
	}
	d := json.NewDecoder(readCloser)
	d.UseNumber()
	errJSON := d.Decode(v)
	if errJSON != nil {
		m800log.Debugf(ctx, "[ParseJSONReadCloser] err: %v", errJSON.Error())
		return gopkg.NewCodeError(CodeParseJSON, errJSON.Error())
	}
	return nil
}

// ParseJSON
func ParseJSON(ctx goctx.Context, data []byte, v interface{}) gopkg.CodeError {
	d := json.NewDecoder(bytes.NewBuffer(data))
	d.UseNumber()
	err := d.Decode(v)
	if err != nil {
		m800log.Debugf(ctx, "[ParseJSON] err:%v, input: %s", err.Error(), string(data))
		return gopkg.NewCodeError(CodeParseJSON, err.Error())
	}
	return nil
}

// GetStringFromIO returns the string from given io.ReadCloser
func GetStringFromIO(readCloser io.ReadCloser) string {
	defer readCloser.Close()
	bytes, _ := ioutil.ReadAll(readCloser)
	return string(bytes)
}

func dumpRequest(ctx goctx.Context, level logrus.Level, req *http.Request) {
	if req == nil {
		return
	}
	token := req.Header.Get(HeaderAuthorization)
	req.Header.Del(HeaderAuthorization)
	defer req.Header.Set(HeaderAuthorization, token)
	requestDump, _ := httputil.DumpRequest(req, true)

	log := "DumpRequest: " + string(requestDump)
	if len(log) >= defaultDumpLimitLength {
		log = log[:defaultDumpLimitLength]
	}

	m800log.Logf(ctx, level, log)
}

func dumpRequestGivenBody(ctx goctx.Context, level logrus.Level, req *http.Request, body []byte) {
	if req == nil {
		return
	}
	token := req.Header.Get(HeaderAuthorization)
	req.Header.Del(HeaderAuthorization)
	defer req.Header.Set(HeaderAuthorization, token)
	requestDump, _ := httputil.DumpRequest(req, false)

	log := "DumpRequest: " + string(requestDump) + " Body: " + string(body)
	if len(log) >= defaultDumpLimitLength {
		log = log[:defaultDumpLimitLength]
	}

	m800log.Logf(ctx, level, log)
}

// LogDumpRequest check level first and log by http resp body, because we don't want to waste resource on DumpRequest
func LogDumpRequest(ctx goctx.Context, level logrus.Level, req *http.Request) {
	if m800log.GetLogger().Level >= level {
		dumpRequest(ctx, level, req)
	}
}

// LogDumpRequestGivenBody with given body
func LogDumpRequestGivenBody(ctx goctx.Context, level logrus.Level, req *http.Request, body []byte) {
	if m800log.GetLogger().Level >= level {
		dumpRequestGivenBody(ctx, level, req, body)
	}
}

// LogDumpResponse by http resp body
func LogDumpResponse(ctx goctx.Context, level logrus.Level, resp *http.Response) {
	_ = logDumpResponsePrinted(ctx, level, resp, false)
}

// LogDumpResponseGivenBody with given body
func LogDumpResponseGivenBody(ctx goctx.Context, level logrus.Level, resp *http.Response, body []byte) {
	_ = logDumpResponseGivenBodyPrinted(ctx, level, resp, body, false)
}

// logDumpResponsePrinted by http resp body
func logDumpResponsePrinted(ctx goctx.Context, level logrus.Level, resp *http.Response, printed bool) bool {
	if printed {
		return true
	}
	if resp == nil {
		return true
	}
	if m800log.GetLogger().Level >= level {
		respDump, _ := httputil.DumpResponse(resp, true)

		log := "DumpResponse: " + string(respDump)
		if len(log) >= defaultDumpLimitLength {
			log = log[:defaultDumpLimitLength]
		}

		m800log.Logf(ctx, level, log)
		return true
	}
	return false
}

// logDumpResponseGivenBodyPrinted
func logDumpResponseGivenBodyPrinted(ctx goctx.Context, level logrus.Level, resp *http.Response, body []byte, printed bool) bool {
	if printed {
		return true
	}
	if resp == nil {
		return true
	}
	if m800log.GetLogger().Level >= level {
		respDump, _ := httputil.DumpResponse(resp, false)

		log := "DumpResponse: " + string(respDump) + " Body: " + string(body)
		if len(log) >= defaultDumpLimitLength {
			log = log[:defaultDumpLimitLength]
		}

		m800log.Logf(ctx, level, log)
		return true
	}
	return false
}

// GetContextFromGin is a util generated the goctx from gin.Context
func GetContextFromGin(c *gin.Context) goctx.Context {
	if ctxI, gok := c.Get(goctx.ContextKey); gok {
		ctx, rok := ctxI.(goctx.Context)
		if rok {
			return ctx
		}
	}

	ctx := goctx.GetContextFromHeader(c.Request.Header)
	// Inject trace_context and baggage in goctx.Context (implementation is MapContext.Context)
	// because goctx cannot provide the interface for carrier
	ctx.SetSpan(trace.SpanFromContext(c.Request.Context()))
	ctx.SetBaggage(baggage.FromContext(c.Request.Context()))

	c.Set(goctx.ContextKey, ctx)

	return ctx
}

func GetInternalFirstCaller(c goctx.Context) string {
	callerStr, _ := c.GetString(goctx.LogKeyInternalCaller)
	if callerStr == "" {
		return ""
	}

	return strings.Split(callerStr, ",")[0]
}

func ContainsCaller(c goctx.Context, caller string) bool {
	callerStr, _ := c.GetString(goctx.LogKeyInternalCaller)

	strs := strings.Split(callerStr, ",")
	for _, str := range strs {
		if str == caller {
			return true
		}
	}
	return false
}

func GetInternalLastCaller(c goctx.Context) string {
	callerStr, _ := c.GetString(goctx.LogKeyInternalCaller)
	if callerStr == "" {
		return ""
	}
	strs := strings.Split(callerStr, ",")
	n := len(strs)
	return strs[n-1]
}

func GetCallerName(callerName string) string {
	return getCallerName(callerName, 1)
}

func getCallerName(callerName string, skip int) string {
	fpcs := make([]uintptr, 1)
	runtime.Callers(2+skip, fpcs)
	caller := runtime.FuncForPC(fpcs[0] - 1)
	if caller != nil {
		callerName = caller.Name()
	}
	return callerName
}

func PrintGinRouteInfo(rs []gin.RouteInfo) {
	type RouteInfo struct {
		Method string `json:"method"`
		Path   string `json:"path"`
	}
	var ris []RouteInfo
	for _, r := range rs {
		ri := RouteInfo{
			r.Method,
			r.Path,
		}
		ris = append(ris, ri)
	}
	b, _ := json.Marshal(ris)
	fmt.Printf("%s\n", b)
}
