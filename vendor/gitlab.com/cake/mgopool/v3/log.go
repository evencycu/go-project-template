package mgopool

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	"gitlab.com/cake/goctx"
	"gitlab.com/cake/m800log"
)

const (
	DB_TYPE_FIELD    = "DBType"
	DB_HOST_FIELD    = "Dh"
	DB_COMMAND_FIELD = "DBCommand"
	DB_TYPE          = "mongo"
)

// systemLogHeaders used for error log format
var systemCtx = goctx.Background()

func init() {
	systemCtx.Set(goctx.LogKeyApp, DB_TYPE)
}

// TODO: add processing time metric
func accessLog(ctx goctx.Context, host []string, method, sql string, start time.Time) {
	m800log.AccessFields(ctx, start, logrus.Fields{
		DB_HOST_FIELD:    host,
		DB_COMMAND_FIELD: sql,
		goctx.LogKeyCase: method,
		DB_TYPE_FIELD:    DB_TYPE,
	})
}

func errLog(ctx goctx.Context, host []string, v ...interface{}) {
	m800log.GetGeneralEntry(ctx).
		WithField(DB_HOST_FIELD, host).
		WithField(DB_TYPE_FIELD, DB_TYPE).
		Error(v...)
}

func errLogf(ctx goctx.Context, host []string, format string, v ...interface{}) {
	m800log.GetGeneralEntry(ctx).
		WithField(DB_HOST_FIELD, host).
		WithField(DB_TYPE_FIELD, DB_TYPE).
		Errorf(format, v...)
}

func infoLog(ctx goctx.Context, host []string, msg string) {
	m800log.GetGeneralEntry(ctx).
		WithField(DB_HOST_FIELD, host).
		WithField(DB_TYPE_FIELD, DB_TYPE).
		Info(msg)
}

func getAccessLevel() logrus.Level {
	return m800log.GetAccessLevel()
}

func formatMsg(msg interface{}, level logrus.Level) string {
	stdLogger := m800log.GetLogger()
	if stdLogger.Level >= level {
		value, _ := json.Marshal(msg)
		return string(value)
	}
	return fmt.Sprintf("%+v", msg)
}
