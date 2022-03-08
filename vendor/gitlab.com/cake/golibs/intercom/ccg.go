package intercom

import (
	"fmt"
	"net"
	"net/url"
	"strings"

	"gitlab.com/cake/goctx"
	"gitlab.com/cake/m800log"
)

const (
	ccgAppName       = "cross-cluster-gateway"
	ccgForwardURL    = "X-M800-CCG-FORWARD-URL"
	ccgForwardRegion = "X-M800-CCG-FORWARD-REGION"
)

const (
	ccgHTTPProxyScheme = "http"
	ccgHTTPProxyPort   = "8999"
	ccgHTTPProxyV1Path = "/internal/v1/do-proxy"
)

const (
	ccgGRPCProxyPort = "8899"
)

var (
	ccgHTTPProxyV1FullURLStr = fmt.Sprintf("%s://%s:%s%s", ccgHTTPProxyScheme, ccgAppName, ccgHTTPProxyPort, ccgHTTPProxyV1Path)
	ccgHTTPProxyHost         = fmt.Sprintf("%s:%s", ccgAppName, ccgHTTPProxyPort)
	ccgGRPCProxyHost         = fmt.Sprintf("dns:///%s:%s", ccgAppName, ccgGRPCProxyPort)
)

var (
	ccgHTTPProxyV1FullURL *url.URL
)

// expect url to full path, could be one of them
// http://svc.ns:8999
// http://svc.ns:8999/
// http://svc.ns:8999/internal/v1/do-something?query=value&query=value2&another=aaa
// https://play.golang.org/p/Mg7RxPPrqTK
func needProxyToCCG(ctx goctx.Context, localNamespace, inputURL string) (needProxy bool, proxyURLStr string, proxyURL *url.URL) {
	funcName := "needProxyToCCG"

	u, err := url.Parse(inputURL)
	if err != nil {
		m800log.Debugf(ctx, "[%s] failed parse url: %s, err: %+v", funcName, inputURL, err)
		return
	}
	host, _, err := net.SplitHostPort(u.Host)
	if err != nil {
		m800log.Debugf(ctx, "[%s] failed split host port: %s, err: %+v", funcName, u.Host, err)
		return
	}
	splitHost := strings.Split(host, ".")
	if len(splitHost) != 2 {
		m800log.Debugf(ctx, "[%s] not valid host, splited host: %+v", funcName, splitHost)
		return
	}

	targetNS := splitHost[1]
	if targetNS == localNamespace {
		m800log.Debugf(ctx, "[%s] no need to proxy request from local namespace: %s", funcName, localNamespace)
		return
	}

	needProxy = true
	proxyURLStr = ccgHTTPProxyV1FullURLStr
	proxyURL = ccgHTTPProxyV1FullURL
	return
}

// getters

func GetForwardURL() string {
	return ccgForwardURL
}

func GetForwardRegion() string {
	return ccgForwardRegion
}

func GetHTTPProxyPort() string {
	return ccgHTTPProxyPort
}

func GetHTTPProxyV1Path() string {
	return ccgHTTPProxyV1Path
}

func GetGRPCProxyHost() string {
	return ccgGRPCProxyHost
}

func GetGRPCProxyPort() string {
	return ccgGRPCProxyPort
}
