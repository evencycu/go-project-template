package m800log

import (
	"io/ioutil"
	"math"
	"os"
	"runtime"
	"time"

	"github.com/sirupsen/logrus"
	"gitlab.com/cake/goctx"
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
	// Dummy for Stupid Usage
	l := logrus.New()
	l.Out = os.Stdout
	l.Level = logrus.DebugLevel
	SetLogger(l)
}

var stdLogger *logrus.Logger
var accessLevel = logrus.InfoLevel
var stackEnabled = false
var blockKeys []string

// SetLogger sets the standard logrus logger
func SetLogger(l *logrus.Logger) {
	stdLogger = l
}

// GetDiscardLogger returns the new logrus logger
func GetDiscardLogger() *logrus.Logger {
	// Dummy for Stupid Usage
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
	// TODO: add back after logrus support skip
	// if level <= logrus.DebugLevel {
	// 	logger.SetReportCaller(true)
	// }
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

// Initialize inits the standard logger with log settings and the blockKeys slice for filtering goctx
func InitializeWithKeysToBlock(output, level string, keysToBlock []string) error {
	blockKeys = append(blockKeys, keysToBlock...)
	return Initialize(output, level)
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

// Trace logs the trace level general log
func Trace(ctx goctx.Context, v ...interface{}) {
	// fast return if level not enough
	if stdLogger.Level >= logrus.TraceLevel {
		GetGeneralEntry(ctx).Trace(v...)
	}
}

// Info logs the info level general log
func Info(ctx goctx.Context, v ...interface{}) {
	// fast return if level not enough
	if stdLogger.Level >= logrus.InfoLevel {
		GetGeneralEntry(ctx).Info(v...)
	}
}

// Warn logs the warn level general log
func Warn(ctx goctx.Context, v ...interface{}) {
	// fast return if level not enough
	if stdLogger.Level >= logrus.WarnLevel {
		GetGeneralEntry(ctx).Warn(v...)
	}
}

// Fatal logs the fatal level general log
func Fatal(ctx goctx.Context, v ...interface{}) {
	// fast return if level not enough
	if stdLogger.Level >= logrus.FatalLevel {
		GetGeneralEntry(ctx).Fatal(v...)
	}
}

// Panic logs the panic level general log
func Panic(ctx goctx.Context, v ...interface{}) {
	// fast return if level not enough
	if stdLogger.Level >= logrus.PanicLevel {
		GetGeneralEntry(ctx).Panic(v...)
	}
}

// Log logs by the given level
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

// Errorf formated the error level log with format
func Errorf(ctx goctx.Context, format string, v ...interface{}) {
	GetGeneralEntry(ctx).Errorf(format, v...)
}

// Tracef formated the trace level log with format
func Tracef(ctx goctx.Context, format string, v ...interface{}) {
	// fast return if level not enough
	if stdLogger.Level >= logrus.TraceLevel {
		GetGeneralEntry(ctx).Tracef(format, v...)
	}
}

// Debugf formated the debug level log with format
func Debugf(ctx goctx.Context, format string, v ...interface{}) {
	// fast return if level not enough
	if stdLogger.Level >= logrus.DebugLevel {
		GetGeneralEntry(ctx).Debugf(format, v...)
	}
}

// Infof formated the info level log with format
func Infof(ctx goctx.Context, format string, v ...interface{}) {
	// fast return if level not enough
	if stdLogger.Level >= logrus.InfoLevel {
		GetGeneralEntry(ctx).Infof(format, v...)
	}
}

// Warnf formated the warn level log with format
func Warnf(ctx goctx.Context, format string, v ...interface{}) {
	GetGeneralEntry(ctx).Warnf(format, v...)
}

// Fatalf formated the fatal level log with format
func Fatalf(ctx goctx.Context, format string, v ...interface{}) {
	GetGeneralEntry(ctx).Fatalf(format, v...)
}

// Panicf formated the panic level log with format
func Panicf(ctx goctx.Context, format string, v ...interface{}) {
	GetGeneralEntry(ctx).Panicf(format, v...)
}

// Logf formated logs by the given level with format
func Logf(ctx goctx.Context, level logrus.Level, format string, v ...interface{}) {
	// fast return if level not enough
	if stdLogger.Level >= level {
		GetGeneralEntry(ctx).Logf(level, format, v...)
	}
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
	ctxMap := ctx.Map()
	for _, key := range blockKeys {
		delete(ctxMap, key)
	}
	return stdLogger.WithFields(logrus.Fields(ctxMap)).
		WithField(goctx.LogKeyAccessTime, SpentTimeInMilliSecond(start)).
		WithField(goctx.LogKeyLogType, AccessType)
}

// GetGeneralEntry return general log entry with preset fields
func GetGeneralEntry(ctx goctx.Context) *logrus.Entry {
	ctxMap := ctx.Map()
	for _, key := range blockKeys {
		delete(ctxMap, key)
	}
	entry := stdLogger.WithFields(logrus.Fields(ctxMap)).
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

func GetAccessLevel() logrus.Level {
	return accessLevel
}