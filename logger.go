package yiigo

import (
	"runtime/debug"

	"github.com/cihub/seelog"
)

// initLogger open a log
func initLogger(path string) {
	logger, _ := seelog.LoggerFromConfigAsFile(path)
	seelog.ReplaceLogger(logger)
}

// Debug debug
func Debug(v ...interface{}) {
	seelog.Debug(v...)
}

// Info info
func Info(v ...interface{}) {
	seelog.Info(v...)
}

// Warn warning
func Warn(v ...interface{}) {
	seelog.Warn(v...)
}

// Error error
func Error(v ...interface{}) {
	seelog.Error(v...)
}

// Err error with debug
func Err(v ...interface{}) {
	v = append(v, "\n", string(debug.Stack()))
	seelog.Error(v...)
}

// Critical critical
func Critical(v ...interface{}) {
	seelog.Critical(v...)
}

// Debugf debug
func Debugf(format string, v ...interface{}) {
	seelog.Debugf(format, v...)
}

// Infof info
func Infof(format string, v ...interface{}) {
	seelog.Infof(format, v...)
}

// Warnf warning
func Warnf(format string, v ...interface{}) {
	seelog.Warnf(format, v...)
}

// Errorf error
func Errorf(format string, v ...interface{}) {
	seelog.Errorf(format, v...)
}

// Errf error with debug
func Errf(format string, v ...interface{}) {
	format += "\n%s"
	v = append(v, string(debug.Stack()))

	seelog.Errorf(format, v...)
}

// Criticalf critical
func Criticalf(format string, v ...interface{}) {
	seelog.Criticalf(format, v...)
}

// Critf critical with debug
func Critf(format string, v ...interface{}) {
	format += "\n%s"
	v = append(v, string(debug.Stack()))

	seelog.Criticalf(format, v...)
}

// Flush processes all currently queued messages and all currently buffered messages immediately
func Flush() {
	seelog.Flush()
}
