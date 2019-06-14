package intercom

import (
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"runtime"
	"strings"
	"time"

	"gitlab.com/general-backend/goctx"
	"gitlab.com/general-backend/gopkg"
	"gitlab.com/general-backend/gotrace"
	"gitlab.com/general-backend/m800log"
)

var (
	httpClient *http.Client
)

func init() {
	defaultTimeout := 30 * time.Second
	tr := &http.Transport{
		MaxIdleConns:       10,
		IdleConnTimeout:    defaultTimeout,
		DisableCompression: true,
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
)

// SetHTTPClient set the package default http client
func SetHTTPClient(client *http.Client) {
	httpClient = client
}

// GetHTTPClient returns the default httpClient, lazy init the httpClient
func GetHTTPClient() *http.Client {
	return httpClient
}

// M800Do
func M800Do(ctx goctx.Context, req *http.Request) (result *JsonResponse, err gopkg.CodeError) {
	httpResp, err := HTTPDo(ctx, req)
	if err != nil {
		ctx.Set(goctx.LogKeyErrorCode, err.ErrorCode())
		return
	}
	result = &JsonResponse{}
	body, errRead := readFromReadCloser(httpResp.Body)
	if errRead != nil {
		ctx.Set(goctx.LogKeyErrorCode, errRead.ErrorCode())
		LogDumpResponse(ctx, ErrorTraceLevel, httpResp)
		return
	}
	err = ParseJSON(ctx, body, result)
	if err != nil {
		ctx.Set(goctx.LogKeyErrorCode, err.ErrorCode())
		LogDumpResponseAndBody(ctx, ErrorTraceLevel, httpResp, body)
		return
	}

	if result.Code != 0 {
		ctx.Set(goctx.LogKeyErrorCode, result.Code)
		LogDumpRequest(ctx, ErrorTraceLevel, req)
		LogDumpResponseAndBody(ctx, ErrorTraceLevel, httpResp, body)
		err = gopkg.NewCodeError(result.Code, result.Message)
		return
	}
	return
}

// HTTPDo
func HTTPDo(ctx goctx.Context, req *http.Request) (resp *http.Response, err gopkg.CodeError) {
	tags := &gotrace.TagsMap{
		Method: req.Method,
		URL:    req.URL,
		Header: req.Header,
	}

	callerName := "HTTPSend"
	fpcs := make([]uintptr, 1)
	runtime.Callers(2, fpcs)
	caller := runtime.FuncForPC(fpcs[0] - 1)
	if caller != nil {
		callerName = caller.Name()
	}
	ctx.SetHTTPHeaders(req.Header)
	// FIXME: performance issue here if use runtime...?
	sp := gotrace.CreateSpanByContext(callerName, ctx, gotrace.ReferenceChildOf, tags)
	defer sp.Finish()
	errInject := gotrace.InjectSpan(sp, req.Header)
	if errInject != nil {
		m800log.Info(ctx, "create inject span error:", errInject)
	}
	var errDo error
	resp, errDo = GetHTTPClient().Do(req)
	if errDo != nil {
		sp.SetTag("client.do.error", err)
		LogDumpRequest(ctx, ErrorTraceLevel, req)
		err = gopkg.NewCodeError(CodeHTTPDo, errDo.Error())
		return
	}
	if resp.StatusCode >= http.StatusBadRequest {
		sp.SetTag("client.do.status", resp.StatusCode)
		LogDumpRequest(ctx, ErrorTraceLevel, req)
		LogDumpResponse(ctx, ErrorTraceLevel, resp)
		err = gopkg.NewCodeError(CodeBadHTTPResponse, fmt.Sprintf("return http code: %d", resp.StatusCode))
	}
	return
}

// HTTPPostForm
func HTTPPostForm(ctx goctx.Context, url string, data url.Values) (resp *http.Response, err gopkg.CodeError) {
	req, err := HTTPNewRequest(ctx, "POST", url, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	return HTTPDo(ctx, req)
}

func HTTPNewRequest(ctx goctx.Context, method, url string, body io.Reader) (*http.Request, gopkg.CodeError) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		m800log.Error(ctx, "new http request error:", err)
		return nil, gopkg.NewCodeError(CodeNewRequest, err.Error())
	}
	return req, nil
}
