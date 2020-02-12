package intercom

import (
	"net/http"
	"net/url"
	"strings"
	"time"

	"gitlab.com/cake/goctx"
	"gitlab.com/cake/gopkg"
)

type IntercomClient struct {
	httpClient *http.Client
}

func NewIntercomClient(client *http.Client) (ic *IntercomClient) {
	if client != nil {
		ic = &IntercomClient{client}
	}
	return
}

func (ic *IntercomClient) SetHTTPClient(client *http.Client) {
	ic.httpClient = client
}

func (ic *IntercomClient) SetHTTPClientTimeout(to time.Duration) {
	ic.httpClient.Timeout = to
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
