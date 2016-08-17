package yiigo

import (
	"fmt"
	"path/filepath"

	"github.com/dlintw/goconf"
)

var (
	config    *goconf.ConfigFile
	configErr error
)

func InitConfig(name string) {
	path, _ := filepath.Abs(fmt.Sprintf("config/%s.config", name))
	config, configErr = goconf.ReadConfigFile(path)

	if configErr != nil {
		LogCritical("load configuration file error: ", configErr.Error())
		return
	}

	fmt.Println("Init Configuration")
}

func GetConfigString(section string, option string, defaultValue string) string {
	conf, err := config.GetString(section, option)

	if err != nil {
		return defaultValue
	}

	return conf
}

func GetConfigInt(section string, option string, defaultValue int) int {
	conf, err := config.GetInt(section, option)

	if err != nil {
		return defaultValue
	}

	return conf
}

func GetConfigFloat64(section string, option string, defaultValue float64) float64 {
	conf, err := config.GetFloat64(section, option)

	if err != nil {
		return defaultValue
	}

	return conf
}

func GetConfigBool(section string, option string, defaultValue bool) bool {
	conf, err := config.GetBool(section, option)

	if err != nil {
		return defaultValue
	}

	return conf
}
