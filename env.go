package yiigo

import (
	"path/filepath"

	"gopkg.in/ini.v1"
)

var env *ini.File

/**
 * 加载ENV配置
 */
func loadEnv(path string) error {
	var err error

	abs, _ := filepath.Abs(path)

	env, err = ini.Load(abs)

	if err != nil {
		return err
	}

	env.BlockMode = false

	return nil
}

/**
 * 获取 string 配置
 * @param section string
 * @param key string
 * @param defaultValue string
 * @return string
 */
func GetEnvString(section string, key string, defaultValue string) string {
	if env == nil {
		return defaultValue
	}

	v := env.Section(section).Key(key).MustString(defaultValue)

	return v
}

/**
 * 获取 int 配置
 * @param section string
 * @param key string
 * @param defaultValue int
 * @return int
 */
func GetEnvInt(section string, key string, defaultValue int) int {
	if env == nil {
		return defaultValue
	}

	v := env.Section(section).Key(key).MustInt(defaultValue)

	return v
}

/**
 * 获取 int64 配置
 * @param section string
 * @param key string
 * @param defaultValue int64
 * @return int64
 */
func GetEnvInt64(section string, key string, defaultValue int64) int64 {
	if env == nil {
		return defaultValue
	}

	v := env.Section(section).Key(key).MustInt64(defaultValue)

	return v
}

/**
 * 获取 float64 配置
 * @param section string
 * @param key string
 * @param defaultValue float64
 * @return float64
 */
func GetEnvFloat64(section string, key string, defaultValue float64) float64 {
	if env == nil {
		return defaultValue
	}

	v := env.Section(section).Key(key).MustFloat64(defaultValue)

	return v
}

/**
 * 获取 bool 配置
 * @param section string
 * @param key string
 * @param defaultValue bool
 * @return bool
 */
func GetEnvBool(section string, key string, defaultValue bool) bool {
	if env == nil {
		return defaultValue
	}

	v := env.Section(section).Key(key).MustBool(defaultValue)

	return v
}

func childSections(section string) []*ini.Section {
	if env == nil {
		return []*ini.Section{}
	}

	return env.Section(section).ChildSections()
}
