package yiigo

import (
	"path/filepath"
	"sync"

	"github.com/Unknwon/goconfig"
)

var (
	env    *goconfig.ConfigFile
	envMux sync.Mutex
)

func initEnv() {
	envMux.Lock()
	defer envMux.Unlock()

	if env == nil {
		var err error

		envFile, _ := filepath.Abs("env.ini")

		env, err = goconfig.LoadConfigFile(envFile)

		if err != nil {
			LogCritical("load env file error: ", err.Error())
			return
		}
	}
}

func GetEnvString(section string, option string, defaultValue string) string {
	if env == nil {
		initEnv()
	}

	val := env.MustValue(section, option, defaultValue)

	return val
}

func GetEnvInt(section string, option string, defaultValue int) int {
	if env == nil {
		initEnv()
	}

	val := env.MustInt(section, option, defaultValue)

	return val
}

func GetEnvInt64(section string, option string, defaultValue int64) int64 {
	if env == nil {
		initEnv()
	}

	val := env.MustInt64(section, option, defaultValue)

	return val
}

func GetEnvFloat64(section string, option string, defaultValue float64) float64 {
	if env == nil {
		initEnv()
	}

	val := env.MustFloat64(section, option, defaultValue)

	return val
}

func GetEnvBool(section string, option string, defaultValue bool) bool {
	if env == nil {
		initEnv()
	}

	val := env.MustBool(section, option, defaultValue)

	return val
}

/**
 * sep 字符串分隔符
 */
func GetEnvArray(section string, option string, sep string) []string {
	if env == nil {
		initEnv()
	}

	val := env.MustValueArray(section, option, sep)

	return val
}
