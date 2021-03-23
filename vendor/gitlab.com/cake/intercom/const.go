package intercom

import (
	"github.com/sirupsen/logrus"
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

type Config struct {
	AppName   string
	Namespace string
}

var (
	ErrorTraceLevel = logrus.WarnLevel
	AppName         = ""
	Namespace       = ""
)

func SetErrorTraceLevel(lvl logrus.Level) {
	ErrorTraceLevel = lvl
}

func Init(config Config) {
	if config.AppName != "" {
		AppName = config.AppName
	}

	if config.Namespace != "" {
		Namespace = config.Namespace
	}
}
