package common

import (
	"crypto/tls"
	"net/http"
	"net/url"
	"runtime"
	"strings"

	"github.com/spf13/viper"
	"gitlab.com/general-backend/goctx"
	"gitlab.com/general-backend/gotrace"
	"gitlab.com/general-backend/m800log"
)

var (
	httpClient *http.Client
)

const (
	// HeaderAuthorization is header key for oauth authorization
	HeaderAuthorization = "Authorization"
	HeaderContentType   = "Content-Type"
	HeaderJSON          = "application/json"
	HeaderJsession      = "jsessionid"
)

// GetHTTPClient returns the default httpClient, lazy init the httpClient
func GetHTTPClient() *http.Client {
	if httpClient != nil {
		return httpClient
	}
	httpClientTimeout := viper.GetDuration("http.client_timeout")

	tr := &http.Transport{
		MaxIdleConns:       10,
		IdleConnTimeout:    httpClientTimeout,
		DisableCompression: true,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}
	httpClient = &http.Client{Transport: tr, Timeout: httpClientTimeout}
	return httpClient
}

func HTTPDo(ctx goctx.Context, req *http.Request) (resp *http.Response, err error) {
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

	// FIXME: performance issue here if use runtime...?
	sp := gotrace.CreateSpanByContext(callerName, ctx, gotrace.ReferenceChildOf, tags)
	defer sp.Finish()
	errInject := gotrace.InjectSpan(sp, req.Header)
	if errInject != nil {
		m800log.Info(ctx, "create inject span error:", errInject)
	}

	resp, err = GetHTTPClient().Do(req)
	if err != nil {
		LogDumpRequest(ctx, req, "http client do error")
		return
	}
	if resp.StatusCode >= http.StatusBadRequest {
		LogDumpRequest(ctx, req, "HTTPDo response>=400")
	}
	return
}

func HTTPPostForm(ctx goctx.Context, url string, data url.Values) (resp *http.Response, err error) {
	req, err := http.NewRequest("POST", url, strings.NewReader(data.Encode()))
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	return HTTPDo(ctx, req)
}
