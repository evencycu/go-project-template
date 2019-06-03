package intercom

import "github.com/sirupsen/logrus"

const (
	TraceTagGinError = "gin.error"
)

const (
	CodeHTTPDo          = 1090101
	CodeParseJSON       = 1090102
	CodeNewRequest      = 1090103
	CodeBadHTTPResponse = 1090104
)

var (
	ErrorTraceLevel = logrus.WarnLevel
)
