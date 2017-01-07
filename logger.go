package yiigo

import (
	"fmt"
	"path/filepath"
	"runtime"

	"github.com/cihub/seelog"
)

/**
 * 初始化日志配置
 */
func InitLogger() {
	path, _ := filepath.Abs("log.xml")
	logger, err := seelog.LoggerFromConfigAsFile(path)

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
	debugMsg := []interface{}{}

	_, file, line, ok := runtime.Caller(1)

	if ok {
		debugMsg = append(debugMsg, fmt.Sprintf("[%s:%d] ", file, line))
	}

	debugMsg = append(debugMsg, msg...)

	seelog.Debug(debugMsg...)
	seelog.Flush()
}

/**
 * 记录 Info 日志
 * @param msg ...interface{}
 */
func LogInfo(msg ...interface{}) {
	infoMsg := []interface{}{}

	_, file, line, ok := runtime.Caller(1)

	if ok {
		infoMsg = append(infoMsg, fmt.Sprintf("[%s:%d] ", file, line))
	}

	infoMsg = append(infoMsg, msg...)

	seelog.Info(infoMsg...)
	seelog.Flush()
}

/**
 * 记录 Warn 日志
 * @param msg ...interface{}
 */
func LogWarn(msg ...interface{}) {
	warnMsg := []interface{}{}

	_, file, line, ok := runtime.Caller(1)

	if ok {
		warnMsg = append(warnMsg, fmt.Sprintf("[%s:%d] ", file, line))
	}

	warnMsg = append(warnMsg, msg...)

	seelog.Warn(warnMsg...)
	seelog.Flush()
}

/**
 * 记录 Error 日志
 * @param msg ...interface{}
 */
func LogError(msg ...interface{}) {
	errorMsg := []interface{}{}

	_, file, line, ok := runtime.Caller(1)

	if ok {
		errorMsg = append(errorMsg, fmt.Sprintf("[%s:%d] ", file, line))
	}

	errorMsg = append(errorMsg, msg...)

	seelog.Error(errorMsg...)
	seelog.Flush()
}

/**
 * 记录 Critical 日志
 * @param msg ...interface{}
 */
func LogCritical(msg ...interface{}) {
	criticalMsg := []interface{}{}

	_, file, line, ok := runtime.Caller(1)

	if ok {
		criticalMsg = append(criticalMsg, fmt.Sprintf("[%s:%d] ", file, line))
	}

	criticalMsg = append(criticalMsg, msg...)

	seelog.Critical(criticalMsg...)
	seelog.Flush()
}

/**
 * 记录 Debug 格式化日志
 * @param format string
 * @param params ...interface{}
 */
func LogDebugf(format string, params ...interface{}) {
	debugParams := []interface{}{}

	_, file, line, ok := runtime.Caller(1)

	if ok {
		debugParams = append(debugParams, file, line)
	}

	debugParams = append(debugParams, params...)
	format = "[%s:%d] " + format

	seelog.Debugf(format, debugParams...)
	seelog.Flush()
}

/**
 * 记录 Info 格式化日志
 * @param format string
 * @param params ...interface{}
 */
func LogInfof(format string, params ...interface{}) {
	infoParams := []interface{}{}

	_, file, line, ok := runtime.Caller(1)

	if ok {
		infoParams = append(infoParams, file, line)
	}

	infoParams = append(infoParams, params...)
	format = "[%s:%d] " + format

	seelog.Infof(format, infoParams...)
	seelog.Flush()
}

/**
 * 记录 Warn 格式化日志
 * @param format string
 * @param params ...interface{}
 */
func LogWarnf(format string, params ...interface{}) {
	warnParams := []interface{}{}

	_, file, line, ok := runtime.Caller(1)

	if ok {
		warnParams = append(warnParams, file, line)
	}

	warnParams = append(warnParams, params...)
	format = "[%s:%d] " + format

	seelog.Warnf(format, warnParams...)
	seelog.Flush()
}

/**
 * 记录 Error 格式化日志
 * @param format string
 * @param params ...interface{}
 */
func LogErrorf(format string, params ...interface{}) {
	errorParams := []interface{}{}

	_, file, line, ok := runtime.Caller(1)

	if ok {
		errorParams = append(errorParams, file, line)
	}

	errorParams = append(errorParams, params...)
	format = "[%s:%d] " + format

	seelog.Errorf(format, errorParams...)
	seelog.Flush()
}

/**
 * 记录 Critical 格式化日志
 * @param format string
 * @param params ...interface{}
 */
func LogCriticalf(format string, params ...interface{}) {
	criticalParams := []interface{}{}

	_, file, line, ok := runtime.Caller(1)

	if ok {
		criticalParams = append(criticalParams, file, line)
	}

	criticalParams = append(criticalParams, params...)
	format = "[%s:%d] " + format

	seelog.Criticalf(format, criticalParams...)
	seelog.Flush()
}
