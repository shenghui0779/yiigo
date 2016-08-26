package yiigo

import (
	"path/filepath"
	"sync"

	"github.com/dlintw/goconf"
)

var (
	config    *goconf.ConfigFile
	configMux sync.Mutex
)

func initConfig() {
	configMux.Lock()
	defer configMux.Unlock()

	if config == nil {
		var err error

		path, _ := filepath.Abs("config/app.config")
		config, err = goconf.ReadConfigFile(path)

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

	conf, err := config.GetString(section, option)

	if err != nil {
		return defaultValue
	}

	return conf
}

func GetConfigInt(section string, option string, defaultValue int) int {
	if config == nil {
		initConfig()
	}

	conf, err := config.GetInt(section, option)

	if err != nil {
		return defaultValue
	}

	return conf
}

func GetConfigFloat64(section string, option string, defaultValue float64) float64 {
	if config == nil {
		initConfig()
	}

	conf, err := config.GetFloat64(section, option)

	if err != nil {
		return defaultValue
	}

	return conf
}

func GetConfigBool(section string, option string, defaultValue bool) bool {
	if config == nil {
		initConfig()
	}

	conf, err := config.GetBool(section, option)

	if err != nil {
		return defaultValue
	}

	return conf
}
