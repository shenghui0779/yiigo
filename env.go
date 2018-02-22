package yiigo

import (
	"errors"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	toml "github.com/pelletier/go-toml"
)

type env struct {
	tree  *toml.Tree
	mutex sync.RWMutex
}

// Env enviroment
var Env *env

// ErrNil returned when config not found.
var ErrNil = errors.New("config not found")

// loadEnv load env file
func loadEnv(path string) {
	abs, _ := filepath.Abs(path)
	tomlTree, err := toml.LoadFile(abs)

	if err != nil {
		panic(err)
	}

	Env = &env{tree: tomlTree}
}

// String return string
func (e *env) String(key string, defaultValue ...string) string {
	dv := ""

	if len(defaultValue) > 0 {
		dv = defaultValue[0]
	}

	i := e.Get(key)

	switch t := i.(type) {
	case string:
		return t
	case int64:
		return strconv.FormatInt(t, 10)
	case uint64:
		return strconv.FormatInt(int64(t), 10)
	case float64:
		return strconv.FormatFloat(t, 'f', -1, 64)
	case bool:
		return strconv.FormatBool(t)
	default:
		return dv
	}
}

// Int return int
func (e *env) Int(key string, defaultValue ...int) int {
	dv := 0

	if len(defaultValue) > 0 {
		dv = defaultValue[0]
	}

	i := e.Get(key)

	switch t := i.(type) {
	case int64:
		return int(t)
	case uint64:
		return int(t)
	case float64:
		return int(t)
	case string:
		if v, err := strconv.ParseInt(t, 0, 0); err == nil {
			return int(v)
		}

		return dv
	case bool:
		if t {
			return 1
		}

		return 0
	default:
		return dv
	}
}

// Int64 return int64
func (e *env) Int64(key string, defaultValue ...int64) int64 {
	var dv int64

	if len(defaultValue) > 0 {
		dv = defaultValue[0]
	}

	i := e.Get(key)

	switch t := i.(type) {
	case int64:
		return t
	case uint64:
		return int64(t)
	case float64:
		return int64(t)
	case string:
		if v, err := strconv.ParseInt(t, 0, 0); err == nil {
			return v
		}

		return dv
	case bool:
		if t {
			return 1
		}

		return 0
	default:
		return dv
	}
}

// Float64 return float64
func (e *env) Float64(key string, defaultValue ...float64) float64 {
	var dv float64

	if len(defaultValue) > 0 {
		dv = defaultValue[0]
	}

	i := e.Get(key)

	switch t := i.(type) {
	case float64:
		return t
	case int64:
		return float64(t)
	case uint64:
		return float64(t)
	case string:
		if v, err := strconv.ParseFloat(t, 64); err == nil {
			return v
		}

		return dv
	case bool:
		if t {
			return 1
		}

		return 0
	default:
		return dv
	}
}

// Bool return bool
func (e *env) Bool(key string, defaultValue ...bool) bool {
	var dv bool

	if len(defaultValue) > 0 {
		dv = defaultValue[0]
	}

	i := e.Get(key)

	switch t := i.(type) {
	case bool:
		return t
	case int64:
		if t != 0 {
			return true
		}

		return false
	case uint64:
		if t != 0 {
			return true
		}

		return false
	case string:
		if v, err := strconv.ParseBool(t); err == nil {
			return v
		}

		return dv
	default:
		return dv
	}
}

// Time return time.Time
func (e *env) Time(key string, defaultValue ...time.Time) time.Time {
	var dv time.Time

	if len(defaultValue) > 0 {
		dv = defaultValue[0]
	}

	i := e.Get(key)

	switch t := i.(type) {
	case time.Time:
		return t
	case string:
		if v, err := time.Parse("2006-01-02 15:04:05", t); err == nil {
			return v
		}

		return dv
	case int64:
		return time.Unix(t, 0)
	case uint64:
		return time.Unix(int64(t), 0)
	default:
		return dv
	}
}

// ToMap return map[string]interface{}
func (e *env) ToMap(key string) map[string]interface{} {
	i := e.Get(key)

	if v, ok := i.(*toml.Tree); ok {
		return v.ToMap()
	}

	return nil
}

// Unmarshal attempts to unmarshal the Tree into a Go struct pointed by dest
func (e *env) Unmarshal(key string, dest interface{}) error {
	i := e.Get(key)

	if i == nil {
		return ErrNil
	}

	if v, ok := i.(*toml.Tree); ok {
		err := v.Unmarshal(dest)

		return err
	}

	return errors.New("value is not a tree of toml")
}

// Get return interface{}
func (e *env) Get(key string) interface{} {
	e.mutex.RLock()
	defer e.mutex.RUnlock()

	i := e.tree.Get(key)

	return i
}
