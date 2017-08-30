package yiigo

import (
	"path/filepath"
	"time"

	"gopkg.in/ini.v1"
)

var env *ini.File

// loadEnv load env file
func loadEnv(path string) {
	var err error

	abs, _ := filepath.Abs(path)
	env, err = ini.Load(abs)

	if err != nil {
		panic(err)
	}

	env.BlockMode = false
}

// EnvString get string config
func EnvString(section string, key string, defaultValue string) string {
	if env == nil {
		return defaultValue
	}

	v := env.Section(section).Key(key).MustString(defaultValue)

	return v
}

// EnvInt get int config
func EnvInt(section string, key string, defaultValue int) int {
	if env == nil {
		return defaultValue
	}

	v := env.Section(section).Key(key).MustInt(defaultValue)

	return v
}

// EnvInt64 get int64 config
func EnvInt64(section string, key string, defaultValue int64) int64 {
	if env == nil {
		return defaultValue
	}

	v := env.Section(section).Key(key).MustInt64(defaultValue)

	return v
}

// EnvFloat64 get float64 config
func EnvFloat64(section string, key string, defaultValue float64) float64 {
	if env == nil {
		return defaultValue
	}

	v := env.Section(section).Key(key).MustFloat64(defaultValue)

	return v
}

// EnvBool get bool config
func EnvBool(section string, key string, defaultValue bool) bool {
	if env == nil {
		return defaultValue
	}

	v := env.Section(section).Key(key).MustBool(defaultValue)

	return v
}

// EnvDuration get duration config
func EnvDuration(section string, key string, defaultValue time.Duration) time.Duration {
	if env == nil {
		return defaultValue
	}

	v := env.Section(section).Key(key).MustDuration(defaultValue)

	return v
}

func childSections(section string) []*ini.Section {
	if env == nil {
		return []*ini.Section{}
	}

	return env.Section(section).ChildSections()
}
