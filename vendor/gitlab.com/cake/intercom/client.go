package intercom

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	conntrack "github.com/eaglerayp/go-conntrack"

	"github.com/opentracing/opentracing-go/ext"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
	"gitlab.com/cake/goctx"
	"gitlab.com/cake/gopkg"
	"gitlab.com/cake/gotrace/v2"
	"gitlab.com/cake/m800log"
)

var (
	httpClient *http.Client
)

func init() {
	defaultTimeout := 30 * time.Second
	tr := &http.Transport{
		DialContext: conntrack.NewDialContextFunc(
			conntrack.DialWithName("intercom"),
		),
		MaxIdleConns:        1000,
		MaxIdleConnsPerHost: 100,
		IdleConnTimeout:     defaultTimeout,
		DisableCompression:  true,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}

	httpClient = &http.Client{Transport: tr, Timeout: defaultTimeout}
}

const (
	// HeaderAuthorization is header key for oauth authorization
	HeaderAuthorization = "Authorization"
	HeaderContentType   = "Content-Type"
	HeaderJSON          = "application/json"
	HeaderForm          = "application/x-www-form-urlencoded"
)

// SetHTTPClient set the package default http client
func SetHTTPClient(client *http.Client) {
	httpClient = client
}

// SetHTTPClientTimeout set the timeout of default http client
func SetHTTPClientTimeout(to time.Duration) {
	httpClient.Timeout = to
}

// GetHTTPClient returns the default httpClient, lazy init the httpClient
func GetHTTPClient() *http.Client {
	return httpClient
}

// HTTPNewRequest
func HTTPNewRequest(ctx goctx.Context, method, url string, body io.Reader) (*http.Request, gopkg.CodeError) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		m800log.Errorf(ctx, fmt.Sprintf("new http request error: %+v", err))
		return nil, gopkg.NewCodeError(CodeNewRequest, err.Error())
	}
	return req, nil
}

// M800Do is used for internal service HTTP request
func M800Do(ctx goctx.Context, req *http.Request) (result *JsonResponse, err gopkg.CodeError) {
	client := GetHTTPClient()
	httpResp, err := httpDo(ctx, client, req, 1)
	if err != nil {
		return nil, err
	}
	return m800DoPostProcessing(ctx, httpResp)
}

// M800Do is used for internal service HTTP request
func M800DoGivenBody(ctx goctx.Context, req *http.Request, body []byte) (result *JsonResponse, err gopkg.CodeError) {
	client := GetHTTPClient()
	httpResp, err := httpDoGivenBody(ctx, client, req, body, 1)
	if err != nil {
		return nil, err
	}
	return m800DoPostProcessing(ctx, httpResp)
}

// HTTPPostForm
func HTTPPostForm(ctx goctx.Context, url string, data url.Values) (resp *http.Response, err gopkg.CodeError) {
	req, err := HTTPNewRequest(ctx, http.MethodPost, url, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set(HeaderContentType, HeaderForm)
	client := GetHTTPClient()
	return httpDo(ctx, client, req, 1)
}

// HTTPDo is used for external service request, will print debug log of request
func HTTPDo(ctx goctx.Context, req *http.Request) (resp *http.Response, err gopkg.CodeError) {
	client := GetHTTPClient()
	return httpDo(ctx, client, req, 1)
}

// HTTPDoGivenBody
func HTTPDoGivenBody(ctx goctx.Context, req *http.Request, body []byte) (resp *http.Response, err gopkg.CodeError) {
	client := GetHTTPClient()
	return httpDoGivenBody(ctx, client, req, body, 1)
}

func httpDoGivenBody(ctx goctx.Context, client *http.Client, req *http.Request, body []byte, skip int) (resp *http.Response, err gopkg.CodeError) {
	req.Body = ioutil.NopCloser(bytes.NewReader(body))
	resp, err = httpDo(ctx, client, req, 1+skip)
	return resp, err
}

func httpDo(ctx goctx.Context, client *http.Client, req *http.Request, skip int) (resp *http.Response, err gopkg.CodeError) {
	tags := &gotrace.TagsMap{
		Method: req.Method,
		URL:    req.URL,
		Header: req.Header,
	}

	callerName := getCallerName("HTTPSend", 1+skip)
	ctx.InjectHTTPHeader(req.Header)
	// FIXME: performance issue here if use runtime...?
	sp := gotrace.CreateChildOfSpan(ctx, callerName)
	defer sp.Finish()

	gotrace.AttachHttpTags(sp, tags)
	errInject := gotrace.InjectSpan(sp, req.Header)
	if errInject != nil {
		m800log.Info(ctx, "create inject span error:", errInject)
	}
	var errDo error
	start := time.Now()
	resp, errDo = client.Do(req)
	if errDo != nil {
		sp.SetTag("client.do.error", errDo)
		ext.SamplingPriority.Set(sp, uint16(1))
		LogDumpRequest(ctx, ErrorTraceLevel, req)
		m800log.Errorf(ctx, "[HTTP Do] error: %s", errDo)
		return nil, gopkg.NewCodeError(CodeHTTPDo, errDo.Error())
	}
	respCode := getResponseMetricCode(resp)
	upstreamCounter.With(prometheus.Labels{
		labelCode: respCode,
		labelHost: req.URL.Host,
	}).Inc()
	upstreamDuration.With(prometheus.Labels{
		labelCode: respCode,
		labelHost: req.URL.Host,
	}).Observe(time.Since(start).Seconds())
	return resp, nil
}

func m800DoPostProcessing(ctx goctx.Context, httpResp *http.Response) (result *JsonResponse, err gopkg.CodeError) {
	// should not get this nil http resp
	if httpResp == nil {
		m800log.Error(ctx, "[M800Do] got nil http response with no error")
		return nil, gopkg.NewCodeError(CodeBadHTTPResponse, "nil response")
	}

	respPrinted := logDumpResponsePrinted(ctx, logrus.DebugLevel, httpResp, false)
	body, err := ReadFromReadCloser(httpResp.Body)
	if err != nil {
		ctx.Set(goctx.LogKeyErrorCode, err.ErrorCode())
		_ = logDumpResponsePrinted(ctx, ErrorTraceLevel, httpResp, respPrinted)
		return nil, err
	}

	result = &JsonResponse{}
	result.HTTPStatus = httpResp.StatusCode
	err = ParseJSON(ctx, body, result)
	if err != nil {
		ctx.Set(goctx.LogKeyErrorCode, err.ErrorCode())
		_ = logDumpResponseGivenBodyPrinted(ctx, ErrorTraceLevel, httpResp, body, respPrinted)
		return result, err
	}

	if result.Code != 0 {
		ctx.Set(goctx.LogKeyErrorCode, result.Code)
		_ = logDumpResponseGivenBodyPrinted(ctx, ErrorTraceLevel, httpResp, body, respPrinted)
		return result, gopkg.NewCodeError(result.Code, result.Message)
	}
	return result, nil
}

func getResponseMetricCode(resp *http.Response) (status string) {
	if resp == nil {
		return "error"
	}
	c := resp.StatusCode
	switch {
	case c >= 500:
		status = "5xx"
	case c >= 400: // Client error.
		status = "4xx"
	case c >= 300: // Redirection.
		status = "3xx"
	case c >= 200: // Success.
		status = "2xx"
	default: // Informational.
		status = resp.Status
	}
	return status
}
