package m800log

import (
	"io/ioutil"
	"math"
	"os"
	"runtime"
	"time"

	"github.com/sirupsen/logrus"
	"gitlab.com/general-backend/goctx"
)

const (
	// Discard is the config case to ioutil.Discard
	Discard = "Discard"
	// Stdout is the config case to ioutil.stdout
	Stdout = "Stdout"
	// GeneralType is the general log type (error, info, debug) const
	GeneralType = "General"
	// AccessType is the access log type const
	AccessType = "Access"

	stackField = "eStack"

	AccessValue = "A"
)

func init() {
	// Dummmy for Stupid Usage
	l := logrus.New()
	l.Out = os.Stdout
	l.Level = logrus.DebugLevel
	SetLogger(l)
}

var stdLogger *logrus.Logger
var accessLevel = logrus.InfoLevel
var stackEnabled = false

// SetLogger sets the standard logrus logger
func SetLogger(l *logrus.Logger) {
	stdLogger = l
}

// GetDiscardLogger returns the new logrus logger
func GetDiscardLogger() *logrus.Logger {
	// Dummmy for Stupid Usage
	l := logrus.New()
	l.Out = ioutil.Discard
	l.Level = logrus.PanicLevel
	return l
}

// GetLogger returns the standard logrus logger
func GetLogger() *logrus.Logger {
	return stdLogger
}

var logFd *os.File

func setOutput(logger *logrus.Logger, fileName string) (err error) {
	switch fileName {
	case Discard:
		logger.Out = ioutil.Discard
		return
	case Stdout:
		logger.Out = os.Stdout
		return
	default:
		logFd, err = os.OpenFile(fileName, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			return err
		}

		logger.Out = logFd
	}

	return
}

func setLevel(logger *logrus.Logger, lvl string) (err error) {
	level, err := logrus.ParseLevel(lvl)
	if err == nil {
		logger.SetLevel(level)
	}
	if level <= logrus.DebugLevel {
		logger.SetReportCaller(true)
	}
	return
}

func SetAccessLevel(lvl string) (err error) {
	level, err := logrus.ParseLevel(lvl)
	if err == nil {
		accessLevel = level
	}
	return
}

// Initialize inits the standard logger with log settings
func Initialize(output, level string) (err error) {
	logger := logrus.New()

	err = setOutput(logger, output)
	if err != nil {
		return
	}
	err = setLevel(logger, level)
	if err != nil {
		return
	}

	SetLogger(logger)
	return
}

// SetStackTrace is used to start log the stack info
func SetStackTrace(enable bool) {
	stackEnabled = enable
}

// SetM800JSONFormatter set the M800JSONformatter for the standard logger
func SetM800JSONFormatter(timestampFormat, app, version, env, ns string) {
	stdLogger.Formatter = newM800JSONFormatter(timestampFormat, app, version, env, ns)
}

// Error logs the error level general log
func Error(ctx goctx.Context, v ...interface{}) {
	GetGeneralEntry(ctx).Error(v...)
}

// Debug logs the debug level general log
func Debug(ctx goctx.Context, v ...interface{}) {
	// fast return if level not enough
	if stdLogger.Level >= logrus.DebugLevel {
		GetGeneralEntry(ctx).Debug(v...)
	}
}

// Info logs the info level general log
func Info(ctx goctx.Context, v ...interface{}) {
	// fast return if level not enough
	if stdLogger.Level >= logrus.InfoLevel {
		GetGeneralEntry(ctx).Info(v...)
	}
}

// Info logs the info level general log
func Log(ctx goctx.Context, level logrus.Level, v ...interface{}) {
	// fast return if level not enough
	if stdLogger.Level >= level {
		GetGeneralEntry(ctx).Log(level, v...)
	}
}

// Access logs the access log with preset access level.
func Access(ctx goctx.Context, start time.Time) {
	GetAccessEntry(ctx, start).Log(accessLevel, AccessValue)
}

// AccessFields logs the access log with preset access level and given fields.
func AccessFields(ctx goctx.Context, start time.Time, fields logrus.Fields) {
	GetAccessEntry(ctx, start).WithFields(fields).Log(accessLevel, AccessValue)
}

// SpentTimeInMilliSecond computes the access time in milliseconds
func SpentTimeInMilliSecond(t time.Time) int64 {
	return int64(math.Ceil(time.Since(t).Seconds() * 1e3))
}

// GetAccessEntry return access log entry with preset fields
func GetAccessEntry(ctx goctx.Context, start time.Time) *logrus.Entry {
	return stdLogger.WithFields(ctx.LogFields()).
		WithField(goctx.LogKeyAccessTime, SpentTimeInMilliSecond(start)).
		WithField(goctx.LogKeyLogType, AccessType)
}

// GetGeneralEntry return general log entry with preset fields
func GetGeneralEntry(ctx goctx.Context) *logrus.Entry {
	entry := stdLogger.WithFields(ctx.LogFields()).
		WithField(goctx.LogKeyLogType, GeneralType)
	if stackEnabled {
		entry = entry.WithField(stackField, stackTrace())
	}
	return entry
}

func stackTrace() string {
	buf := make([]byte, 1024)
	var size int
	for {
		size = runtime.Stack(buf, false)
		// The size of the buffer may be not enough to hold the stacktrace,
		// so double the buffer size
		if size >= len(buf) {
			buf = make([]byte, len(buf)<<1)
			continue
		}
		break
	}

	return string(buf[:size])
}
