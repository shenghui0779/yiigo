package yiigo

import (
	"path/filepath"
	"sync"

	"github.com/cihub/seelog"
)

var (
	logger seelog.LoggerInterface
	logMux sync.Mutex
)

func initLogger() {
	logMux.Lock()
	defer logMux.Unlock()

	if logger == nil {
		var err error
		path, _ := filepath.Abs("config/log.xml")
		logger, err = seelog.LoggerFromConfigAsFile(path)

		if err != nil {
			seelog.Critical("load log file error: ", err.Error())
			seelog.Flush()
			return
		}

		seelog.ReplaceLogger(logger)
	}
}

func LogDebug(msg ...interface{}) {
	if logger == nil {
		initLogger()
	}

	seelog.Debug(msg...)
	seelog.Flush()
}

func LogInfo(msg ...interface{}) {
	if logger == nil {
		initLogger()
	}

	seelog.Info(msg...)
	seelog.Flush()
}

func LogWarn(msg ...interface{}) {
	if logger == nil {
		initLogger()
	}

	seelog.Warn(msg...)
	seelog.Flush()
}

func LogError(msg ...interface{}) {
	if logger == nil {
		initLogger()
	}

	seelog.Error(msg...)
	seelog.Flush()
}

func LogCritical(msg ...interface{}) {
	if logger == nil {
		initLogger()
	}

	seelog.Critical(msg...)
	seelog.Flush()
}

func LogDebugf(format string, params ...interface{}) {
	if logger == nil {
		initLogger()
	}

	seelog.Debugf(format, params...)
	seelog.Flush()
}

func LogInfof(format string, params ...interface{}) {
	if logger == nil {
		initLogger()
	}

	seelog.Infof(format, params...)
	seelog.Flush()
}

func LogWarnf(format string, params ...interface{}) {
	if logger == nil {
		initLogger()
	}

	seelog.Warnf(format, params...)
	seelog.Flush()
}

func LogErrorf(format string, params ...interface{}) {
	if logger == nil {
		initLogger()
	}

	seelog.Errorf(format, params...)
	seelog.Flush()
}

func LogCriticalf(format string, params ...interface{}) {
	if logger == nil {
		initLogger()
	}

	seelog.Criticalf(format, params...)
	seelog.Flush()
}
