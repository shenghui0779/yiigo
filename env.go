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
func (e *env) String(key string, defVal ...string) string {
	val := ""

	if len(defVal) > 0 {
		val = defVal[0]
	}

	conf := e.Get(key)

	if conf == nil {
		return val
	}

	if v, ok := conf.(string); ok {
		val = v
	}

	return val
}

// Int return int
func (e *env) Int(key string, defVal ...int) int {
	val := 0

	if len(defVal) > 0 {
		val = defVal[0]
	}

	conf := e.Get(key)

	if conf == nil {
		return val
	}

	if v, ok := conf.(int); ok {
		val = v
	}

	return val
}

// Int64 return int64
func (e *env) Int64(key string, defVal ...int64) int64 {
	var val int64

	if len(defVal) > 0 {
		val = defVal[0]
	}

	conf := e.Get(key)

	if conf == nil {
		return val
	}

	if v, ok := conf.(int64); ok {
		val = v
	}

	return val
}

// Float64 return float64
func (e *env) Float64(key string, defVal ...float64) float64 {
	var val float64

	if len(defVal) > 0 {
		val = defVal[0]
	}

	conf := e.Get(key)

	if conf == nil {
		return val
	}

	if v, ok := conf.(float64); ok {
		val = v
	}

	return val
}

// Bool return bool
func (e *env) Bool(key string, defVal ...bool) bool {
	var val bool

	if len(defVal) > 0 {
		val = defVal[0]
	}

	conf := e.Get(key)

	if conf == nil {
		return val
	}

	if v, ok := conf.(bool); ok {
		val = v
	}

	return val
}

// ToMap return map[string]interface{}
func (e *env) ToMap(key string) map[string]interface{} {
	v := e.Get(key)

	if v == nil {
		return nil
	}

	if node, ok := v.(*toml.Tree); ok {
		return node.ToMap()
	}

	return nil
}

// Unmarshal attempts to unmarshal the Tree into a Go struct pointed by dest
func (e *env) Unmarshal(key string, dest interface{}) error {
	conf := e.Get(key)

	if conf == nil {
		return ErrNil
	}

	if node, ok := conf.(*toml.Tree); ok {
		err := node.Unmarshal(dest)

		return err
	}

	return errors.New("value is not a tree of toml")
}

// Get return interface{}
func (e *env) Get(key string) interface{} {
	e.mutex.RLock()
	defer e.mutex.RUnlock()

	v := e.tree.Get(key)

	return v
}
