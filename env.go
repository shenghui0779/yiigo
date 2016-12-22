package yiigo

import (
	"path/filepath"
	"strings"
	"sync"

	"github.com/Unknwon/goconfig"
)

var (
	env    *goconfig.ConfigFile
	envMux sync.Mutex
)

/**
 * 初始化ENV配置
 */
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

/**
 * 获取 string 配置
 * @param section string
 * @param option string
 * @param defaultValue string
 * @return string
 */
func GetEnvString(section string, option string, defaultValue string) string {
	if env == nil {
		initEnv()
	}

	val := env.MustValue(section, option, defaultValue)

	return val
}

/**
 * 获取 int 配置
 * @param section string
 * @param option string
 * @param defaultValue int
 * @return int
 */
func GetEnvInt(section string, option string, defaultValue int) int {
	if env == nil {
		initEnv()
	}

	val := env.MustInt(section, option, defaultValue)

	return val
}

/**
 * 获取 int64 配置
 * @param section string
 * @param option string
 * @param defaultValue int64
 * @return int64
 */
func GetEnvInt64(section string, option string, defaultValue int64) int64 {
	if env == nil {
		initEnv()
	}

	val := env.MustInt64(section, option, defaultValue)

	return val
}

/**
 * 获取 float64 配置
 * @param section string
 * @param option string
 * @param defaultValue float64
 * @return float64
 */
func GetEnvFloat64(section string, option string, defaultValue float64) float64 {
	if env == nil {
		initEnv()
	}

	val := env.MustFloat64(section, option, defaultValue)

	return val
}

/**
 * 获取 bool 配置
 * @param section string
 * @param option string
 * @param defaultValue bool
 * @return bool
 */
func GetEnvBool(section string, option string, defaultValue bool) bool {
	if env == nil {
		initEnv()
	}

	val := env.MustBool(section, option, defaultValue)

	return val
}

/**
 * 获取 []string 配置
 * @param section string
 * @param option string
 * @param sep string 字符串分隔符(建议使用：,)
 * @param defaultValue string
 * @return []string
 */
func GetEnvArray(section string, option string, sep string, defaultValue string) []string {
	if env == nil {
		initEnv()
	}

	arr := env.MustValueArray(section, option, sep)

	if len(arr) == 0 {
		arr = strings.Split(defaultValue, sep)
	}

	return arr
}
