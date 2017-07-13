package yiigo

import (
	"path/filepath"
	"runtime/debug"

	"github.com/cihub/seelog"
)

/**
 * 初始化日志配置
 */
func initLogger(path string) {
	abs, _ := filepath.Abs(path)
	logger, err := seelog.LoggerFromConfigAsFile(abs)

	if err != nil {
		panic(err)
	}

	seelog.ReplaceLogger(logger)
}

/**
 * 记录 Debug 日志
 * @param msg ...interface{}
 */
func LogDebug(msg ...interface{}) {
	seelog.Debug(msg...)
	seelog.Flush()
}

/**
 * 记录 Info 日志
 * @param msg ...interface{}
 */
func LogInfo(msg ...interface{}) {
	seelog.Info(msg...)
	seelog.Flush()
}

/**
 * 记录 Warn 日志
 * @param msg ...interface{}
 */
func LogWarn(msg ...interface{}) {
	seelog.Warn(msg...)
	seelog.Flush()
}

/**
 * 记录 Error 日志
 * @param msg ...interface{}
 */
func LogError(msg ...interface{}) {
	msg = append(msg, "\n", string(debug.Stack()))

	seelog.Error(msg...)
	seelog.Flush()
}

/**
 * 记录 Critical 日志
 * @param msg ...interface{}
 */
func LogCritical(msg ...interface{}) {
	msg = append(msg, "\n", string(debug.Stack()))

	seelog.Critical(msg...)
	seelog.Flush()
}

/**
 * 记录 Debug 格式化日志
 * @param format string
 * @param params ...interface{}
 */
func LogDebugf(format string, params ...interface{}) {
	seelog.Debugf(format, params...)
	seelog.Flush()
}

/**
 * 记录 Info 格式化日志
 * @param format string
 * @param params ...interface{}
 */
func LogInfof(format string, params ...interface{}) {
	seelog.Infof(format, params...)
	seelog.Flush()
}

/**
 * 记录 Warn 格式化日志
 * @param format string
 * @param params ...interface{}
 */
func LogWarnf(format string, params ...interface{}) {
	seelog.Warnf(format, params...)
	seelog.Flush()
}

/**
 * 记录 Error 格式化日志
 * @param format string
 * @param params ...interface{}
 */
func LogErrorf(format string, params ...interface{}) {
	format += "\n%s"
	params = append(params, string(debug.Stack()))

	seelog.Errorf(format, params...)
	seelog.Flush()
}

/**
 * 记录 Critical 格式化日志
 * @param format string
 * @param params ...interface{}
 */
func LogCriticalf(format string, params ...interface{}) {
	format += "\n%s"
	params = append(params, string(debug.Stack()))

	seelog.Criticalf(format, params...)
	seelog.Flush()
}
