package intercom

import (
	"net/url"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"gitlab.com/cake/gopkg"
)

const (
	TraceTagGinErrorCode = "gin.error.code"
	LogEntryHandlerName  = "entryHandler"
)

const (
	KeyBody = "rb"
)

const (
	CodeHTTPDo           = 1090101
	CodeParseJSON        = 1090102
	CodeNewRequest       = 1090103
	CodeBadHTTPResponse  = 1090104
	CodeReadAll          = 1090105
	CodeNilRequest       = 1090106
	CodeEmptyServiceHome = 1090107
	CodePanic            = 1090108
	CodeMaliciousHeader  = 1090109

	MsgEmpty              = "intercom response no message"
	MsgEmptyServiceHome   = "service home is empty"
	MsgErrMaliciousHeader = "malicious header"
)

var (
	appName        = gopkg.GetAppName()
	localNamespace = ""
	ccgEnabled     = false
)

var (
	singleFlightRequestDuration = 100 * time.Millisecond // limited parallelism , accept 1 request every 100ms
)

var (
	ErrorTraceLevel = logrus.WarnLevel
)

func SetErrorTraceLevel(lvl logrus.Level) {
	ErrorTraceLevel = lvl
}

// called after viper inited
func Init() {
	ccgEnabled = viper.GetBool("ccg.enabled")

	localNamespace = viper.GetString("app.namespace")
	if localNamespace == "" {
		panic("[intercom] empty local namespace")
	}

	var err error
	ccgHTTPProxyV1FullURL, err = url.Parse(ccgHTTPProxyV1FullURLStr)
	if err != nil {
		panic(err)
	}
	if ccgHTTPProxyV1FullURL.String() == "" {
		panic("[intercom] empty ccg http proxy v1 full url")
	}
}
