package yiigo

import (
	"fmt"
	"runtime/debug"

	log "github.com/go-ozzo/ozzo-log"
)

var logger *log.Logger

/**
 * 初始化日志配置
 */
func initLogger() {
	logger = log.NewLogger()
	logger = logger.GetLogger("app", func(l *log.Logger, e *log.Entry) string {
		return fmt.Sprintf("%s [%v] %s", e.Time.Format("2006-01-02 15:04:05"), e.Level, e.Message)
	})

	if EnvBool("app", "debug", true) {
		t := log.NewConsoleTarget()
		logger.Targets = append(logger.Targets, t)
	} else {
		t := log.NewFileTarget()
		t.FileName = EnvString("log", "path", "app.log")
		logger.Targets = append(logger.Targets, t)
	}
}

// LogOpen 打开日志
func LogOpen() {
	logger.Open()
}

// LogClose 关闭日志
func LogClose() {
	logger.Close()
}

/**
 * LogDebug 记录 Debug 日志
 * @param format string
 * @param params ...interface{}
 */
func LogDebug(format string, params ...interface{}) {
	logger.Debug(format, params...)
}

/**
 * LogInfo 记录 Info 日志
 * @param format string
 * @param params ...interface{}
 */
func LogInfo(format string, params ...interface{}) {
	logger.Info(format, params...)
}

/**
 * LogWarn 记录 Warn 日志
 * @param format string
 * @param params ...interface{}
 */
func LogWarn(format string, params ...interface{}) {
	logger.Warning(format, params...)
}

/**
 * LogError 记录 Error 日志
 * @param format string
 * @param params ...interface{}
 */
func LogError(format string, params ...interface{}) {
	format += "\n%s"
	params = append(params, string(debug.Stack()))

	logger.Error(format, params...)
}

/**
 * LogCritical 记录 Critical 日志
 * @param format string
 * @param params ...interface{}
 */
func LogCritical(format string, params ...interface{}) {
	format += "\n%s"
	params = append(params, string(debug.Stack()))

	logger.Critical(format, params...)
}

/**
 * LogAlert 记录 Warn 日志
 * @param format string
 * @param params ...interface{}
 */
func LogAlert(format string, params ...interface{}) {
	logger.Alert(format, params...)
}

/**
 * LogEmergency 记录 Warn 日志
 * @param format string
 * @param params ...interface{}
 */
func LogEmergency(format string, params ...interface{}) {
	logger.Emergency(format, params...)
}
