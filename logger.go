package yiigo

import (
	"fmt"
	"path/filepath"

	"github.com/cihub/seelog"
)

func InitLogger(name string) {
	path, _ := filepath.Abs(fmt.Sprintf("config/%s.xml", name))
	logger, logErr := seelog.LoggerFromConfigAsFile(path)

	if logErr != nil {
		seelog.Critical("load log file error: ", logErr.Error())
		seelog.Flush()
		return
	}

	seelog.ReplaceLogger(logger)

	fmt.Println("Init Logger")
}

func LogDebug(msg ...interface{}) {
	seelog.Debug(msg...)
	seelog.Flush()
}

func LogInfo(msg ...interface{}) {
	seelog.Info(msg...)
	seelog.Flush()
}

func LogWarn(msg ...interface{}) {
	seelog.Warn(msg...)
	seelog.Flush()
}

func LogError(msg ...interface{}) {
	seelog.Error(msg...)
	seelog.Flush()
}

func LogCritical(msg ...interface{}) {
	seelog.Critical(msg...)
	seelog.Flush()
}

func LogDebugf(format string, params ...interface{}) {
	seelog.Debugf(format, params...)
	seelog.Flush()
}

func LogInfof(format string, params ...interface{}) {
	seelog.Infof(format, params...)
	seelog.Flush()
}

func LogWarnf(format string, params ...interface{}) {
	seelog.Warnf(format, params...)
	seelog.Flush()
}

func LogErrorf(format string, params ...interface{}) {
	seelog.Errorf(format, params...)
	seelog.Flush()
}

func LogCriticalf(format string, params ...interface{}) {
	seelog.Criticalf(format, params...)
	seelog.Flush()
}
