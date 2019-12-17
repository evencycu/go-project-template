package intercom

import (
	"github.com/sirupsen/logrus"
)

const (
	TraceTagGinErrorCode = "gin.error.code"
)

const (
	KeyBody = "rb"
)

const (
	CodeHTTPDo          = 1090101
	CodeParseJSON       = 1090102
	CodeNewRequest      = 1090103
	CodeBadHTTPResponse = 1090104
	CodeReadAll         = 1090105
)

var (
	ErrorTraceLevel = logrus.WarnLevel
)

func SetErrorTraceLevel(lvl logrus.Level) {
	ErrorTraceLevel = lvl
}
