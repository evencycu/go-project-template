package intercom

import (
	"bytes"
	"crypto/tls"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	conntrack "github.com/eaglerayp/go-conntrack"
	"gitlab.com/cake/goctx"
	"gitlab.com/cake/gopkg"
	"gitlab.com/cake/m800log"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
)

var (
	intercomClient *IntercomClient
	tr             *http.Transport
)

func init() {
	defaultTimeout := 30 * time.Second
	tr = &http.Transport{
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

	httpClient := &http.Client{Transport: otelhttp.NewTransport(tr), Timeout: defaultTimeout}
	intercomClient = NewIntercomClient(httpClient)
}

const (
	// HeaderAuthorization is header key for oauth authorization
	HeaderAuthorization      = "Authorization"
	HeaderContentType        = "Content-Type"
	HeaderContentDisposition = "Content-Disposition"
	HeaderJSON               = "application/json"
	HeaderForm               = "application/x-www-form-urlencoded"
)

// SetHTTPClient set the package default http client
func SetHTTPClient(client *http.Client) {
	intercomClient.SetHTTPClient(client)
}

// SetHTTPClientTimeout set the timeout of default http client
func SetHTTPClientTimeout(to time.Duration) {
	intercomClient.SetHTTPClientTimeout(to)
}

// GetHTTPClient returns the default httpClient, lazy init the httpClient
func GetHTTPClient() *http.Client {
	return intercomClient.GetHTTPClient()
}

// HTTPNewRequest
func HTTPNewRequest(ctx goctx.Context, method, url string, body io.Reader) (*http.Request, gopkg.CodeError) {
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		ctx.Set(goctx.LogKeyWrapErrorCode, CodeNewRequest)
		return nil, gopkg.NewWrappedCodeError(0, "", gopkg.NewCodeError(CodeNewRequest, err.Error()))
	}
	return req, nil
}

// M800Do is used for internal service HTTP request
func M800Do(ctx goctx.Context, req *http.Request) (result *JsonResponse, err gopkg.CodeError) {
	return intercomClient.M800Do(ctx, req)
}

// M800Do is used for internal service HTTP request
func M800DoGivenBody(ctx goctx.Context, req *http.Request, body []byte) (result *JsonResponse, err gopkg.CodeError) {
	return intercomClient.M800DoGivenBody(ctx, req, body)
}

// HTTPPostForm
func HTTPPostForm(ctx goctx.Context, url string, data url.Values) (resp *http.Response, err gopkg.CodeError) {
	return intercomClient.HTTPPostForm(ctx, url, data)
}

// HTTPDo is used for external service request, will print debug log of request
func HTTPDo(ctx goctx.Context, req *http.Request) (resp *http.Response, err gopkg.CodeError) {
	return intercomClient.HTTPDo(ctx, req)
}

// HTTPDoGivenBody
func HTTPDoGivenBody(ctx goctx.Context, req *http.Request, body []byte) (resp *http.Response, err gopkg.CodeError) {
	return intercomClient.HTTPDoGivenBody(ctx, req, body)
}

func httpDoGivenBody(ctx goctx.Context, client *http.Client, req *http.Request, body []byte, skip int) (resp *http.Response, err gopkg.CodeError) {
	req.Body = ioutil.NopCloser(bytes.NewReader(body))
	resp, err = httpDo(ctx, client, req, 1+skip)
	return resp, err
}

func httpDo(ctx goctx.Context, client *http.Client, req *http.Request, skip int) (resp *http.Response, err gopkg.CodeError) {
	// inject header to ctx
	ctx.InjectHTTPHeader(req.Header)

	// internal caller
	req.Header.Add(goctx.HTTPHeaderInternalCaller, appName)

	propagators := otel.GetTextMapPropagator()
	propagators.Inject(ctx, propagation.HeaderCarrier(req.Header))

	// ccg handling
	if ccgEnabled {
		originURL := req.URL
		needProxy, _, proxyURL := needProxyToCCG(ctx, localNamespace, originURL.String())
		if needProxy {
			req.Header.Add(ccgForwardURL, originURL.String())
			req.URL = proxyURL
		}
	}

	// http client do
	var errDo error
	resp, errDo = client.Do(req)
	if errDo != nil {
		ctx.Set(goctx.LogKeyWrapErrorCode, CodeHTTPDo)
		LogDumpRequest(ctx, ErrorTraceLevel, req)
		return nil, gopkg.NewWrappedCodeError(0, "", gopkg.NewCodeError(CodeHTTPDo, errDo.Error()))
	}
	if resp.StatusCode >= 300 {
		LogDumpRequest(ctx, ErrorTraceLevel, req)
	}
	return resp, nil
}

func m800DoPostProcessing(ctx goctx.Context, httpResp *http.Response) (result *JsonResponse, err gopkg.CodeError) {
	// should not get this nil http resp
	if httpResp == nil {
		m800log.Error(ctx, "[M800Do] got nil http response with no error")
		return nil, gopkg.NewWrappedCodeError(0, "", gopkg.NewCodeError(CodeBadHTTPResponse, "nil response"))
	}

	// m800do should not always print response
	respPrinted := false
	body, err := ReadFromReadCloser(httpResp.Body)
	if err != nil {
		ctx.Set(goctx.LogKeyErrorCode, err.ErrorCode())
		ctx.Set(goctx.LogKeyWrapErrorCode, err.ErrorCode())
		_ = logDumpResponsePrinted(ctx, ErrorTraceLevel, httpResp, respPrinted)
		return nil, err
	}

	result = &JsonResponse{}
	result.HTTPStatus = httpResp.StatusCode
	err = ParseJSON(ctx, body, result)
	if err != nil {
		ctx.Set(goctx.LogKeyErrorCode, err.ErrorCode())
		ctx.Set(goctx.LogKeyWrapErrorCode, err.ErrorCode())
		_ = logDumpResponseGivenBodyPrinted(ctx, ErrorTraceLevel, httpResp, body, respPrinted)
		return result, err
	}

	if result.Code != 0 {
		if result.Message == "" {
			result.Message = MsgEmpty
		}
		ctx.Set(goctx.LogKeyErrorCode, result.Code)
		ctx.Set(goctx.LogKeyWrapErrorCode, result.Code)
		_ = logDumpResponseGivenBodyPrinted(ctx, ErrorTraceLevel, httpResp, body, respPrinted)
		return result, gopkg.NewWrappedCodeError(0, "", gopkg.NewCodeError(result.Code, result.Message))
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

type IntercomClient struct {
	httpClient *http.Client
	userAgent  string
}

func NewIntercomClient(client *http.Client) (ic *IntercomClient) {
	if client != nil {
		ic = &IntercomClient{httpClient: client, userAgent: ""}
	}
	return
}

func (ic *IntercomClient) SetHTTPClient(client *http.Client) {
	ic.httpClient = client
}

func (ic *IntercomClient) SetHTTPClientTimeout(to time.Duration) {
	ic.httpClient.Timeout = to
	if ic.httpClient.Transport != nil {
		ic.httpClient.Transport.(*http.Transport).TLSHandshakeTimeout = to
	}
}

func (ic *IntercomClient) GetHTTPClient() *http.Client {
	return ic.httpClient
}

func (ic *IntercomClient) M800Do(ctx goctx.Context, req *http.Request) (result *JsonResponse, err gopkg.CodeError) {
	client := ic.httpClient

	// internal upstream metrics
	start := time.Now()
	defer func() {
		updateInternalMetrics(req.URL.Host, start, result, err)
	}()

	httpResp, err := httpDo(ctx, client, req, 1)
	if err != nil {
		return nil, err
	}
	return m800DoPostProcessing(ctx, httpResp)
}

func (ic *IntercomClient) M800DoGivenBody(ctx goctx.Context, req *http.Request, body []byte) (result *JsonResponse, err gopkg.CodeError) {
	client := ic.httpClient
	// internal upstream metrics
	start := time.Now()
	defer func() {
		updateInternalMetrics(req.URL.Host, start, result, err)
	}()

	httpResp, err := httpDoGivenBody(ctx, client, req, body, 1)
	if err != nil {
		return nil, err
	}
	return m800DoPostProcessing(ctx, httpResp)
}

func (ic *IntercomClient) HTTPDo(ctx goctx.Context, req *http.Request) (resp *http.Response, err gopkg.CodeError) {
	client := ic.httpClient
	userAgent := ic.userAgent

	if req.Header.Get("User-Agent") == "" && userAgent != "" {
		req.Header.Set("User-Agent", userAgent)
	}

	// external upstream metrics
	start := time.Now()
	defer func() {
		updateExternalMetrics(req.URL.Host, start, resp, err)
	}()

	return httpDo(ctx, client, req, 1)
}

func (ic *IntercomClient) HTTPDoGivenBody(ctx goctx.Context, req *http.Request, body []byte) (resp *http.Response, err gopkg.CodeError) {
	client := ic.httpClient

	// external upstream metrics
	start := time.Now()
	defer func() {
		updateExternalMetrics(req.URL.Host, start, resp, err)
	}()

	return httpDoGivenBody(ctx, client, req, body, 1)
}

func (ic *IntercomClient) HTTPPostForm(ctx goctx.Context, url string, data url.Values) (resp *http.Response, err gopkg.CodeError) {
	req, err := HTTPNewRequest(ctx, http.MethodPost, url, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set(HeaderJSON, HeaderForm)
	client := ic.httpClient

	// external upstream metrics
	start := time.Now()
	defer func() {
		updateExternalMetrics(req.URL.Host, start, resp, err)
	}()

	return httpDo(ctx, client, req, 1)
}

func SetUserAgent(ua string) {
	intercomClient.SetUserAgent(ua)
}

func (ic *IntercomClient) SetUserAgent(ua string) {
	ic.userAgent = ua
}

func GetUserAgent() string {
	return intercomClient.GetUserAgent()
}

func (ic *IntercomClient) GetUserAgent() string {
	return ic.userAgent
}
