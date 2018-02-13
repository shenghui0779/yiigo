package yiigo

import (
	"errors"
	"path/filepath"
	"sync"

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

	if v, ok := i.(string); ok {
		return v
	}

	return dv
}

// Int return int
func (e *env) Int(key string, defaultValue ...int) int {
	dv := 0

	if len(defaultValue) > 0 {
		dv = defaultValue[0]
	}

	i := e.Get(key)

	if v, ok := i.(int64); ok {
		return int(v)
	}

	return dv
}

// Int64 return int64
func (e *env) Int64(key string, defaultValue ...int64) int64 {
	var dv int64

	if len(defaultValue) > 0 {
		dv = defaultValue[0]
	}

	i := e.Get(key)

	if v, ok := i.(int64); ok {
		return v
	}

	return dv
}

// Float64 return float64
func (e *env) Float64(key string, defaultValue ...float64) float64 {
	var dv float64

	if len(defaultValue) > 0 {
		dv = defaultValue[0]
	}

	i := e.Get(key)

	if v, ok := i.(float64); ok {
		return v
	}

	return dv
}

// Bool return bool
func (e *env) Bool(key string, defaultValue ...bool) bool {
	var dv bool

	if len(defaultValue) > 0 {
		dv = defaultValue[0]
	}

	i := e.Get(key)

	if v, ok := i.(bool); ok {
		return v
	}

	return dv
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
