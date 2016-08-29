package yiigo

import (
	"path/filepath"
	"sync"

	"github.com/Unknwon/goconfig"
)

var (
	config    *goconfig.ConfigFile
	configMux sync.Mutex
)

func initConfig() {
	configMux.Lock()
	defer configMux.Unlock()

	if config == nil {
		var err error

		appCfg, _ := filepath.Abs("config/app.ini")
		dbCfg, _ := filepath.Abs("config/db.ini")
		cacheCfg, _ := filepath.Abs("config/cache.ini")

		config, err = goconfig.LoadConfigFile(appCfg, dbCfg, cacheCfg)

		if err != nil {
			LogCritical("load configuration file error: ", err.Error())
			return
		}
	}
}

func GetConfigString(section string, option string, defaultValue string) string {
	if config == nil {
		initConfig()
	}

	conf := config.MustValue(section, option, defaultValue)

	return conf
}

func GetConfigInt(section string, option string, defaultValue int) int {
	if config == nil {
		initConfig()
	}

	conf := config.MustInt(section, option, defaultValue)

	return conf
}

func GetConfigInt64(section string, option string, defaultValue int64) int64 {
	if config == nil {
		initConfig()
	}

	conf := config.MustInt64(section, option, defaultValue)

	return conf
}

func GetConfigFloat64(section string, option string, defaultValue float64) float64 {
	if config == nil {
		initConfig()
	}

	conf := config.MustFloat64(section, option, defaultValue)

	return conf
}

func GetConfigBool(section string, option string, defaultValue bool) bool {
	if config == nil {
		initConfig()
	}

	conf := config.MustBool(section, option, defaultValue)

	return conf
}

/**
 * sep ×Ö·û´®·Ö¸ô·û
 */
func GetConfigArray(section string, option string, sep string) []string {
	if config == nil {
		initConfig()
	}

	conf := config.MustValueArray(section, option, sep)

	return conf
}
