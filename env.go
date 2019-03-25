package yiigo

import (
	"errors"
	"path/filepath"
	"reflect"
	"strconv"
	"sync"
	"time"

	"github.com/pelletier/go-toml"
)

type env struct {
	tree  *toml.Tree
	mutex sync.RWMutex
}

// Env enviroment
var Env *env

// ErrEnvNil returned when config not found.
var ErrEnvNil = errors.New("yiigo: env config not found")

// UseEnv use `toml` config file.
func UseEnv(file string) error {
	path, err := filepath.Abs(file)

	if err != nil {
		return err
	}

	tomlTree, err := toml.LoadFile(path)

	if err != nil {
		return err
	}

	Env = &env{tree: tomlTree}

	return nil
}

// String returns a value of string.
func (e *env) String(key string, defaultValue ...string) string {
	dv := ""

	if len(defaultValue) > 0 {
		dv = defaultValue[0]
	}

	ev := e.Get(key)

	if ev == nil {
		return dv
	}

	switch t := ev.(type) {
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
		return ""
	}
}

// Strings returns a value of []string.
func (e *env) Strings(key string, defaultValue ...string) []string {
	ev := e.Get(key)

	if ev == nil {
		return defaultValue
	}

	v := reflect.Indirect(reflect.ValueOf(ev))

	if v.Kind() != reflect.Slice {
		return []string{}
	}

	l := v.Len()

	result := make([]string, 0, l)

	if l == 0 {
		return result
	}

	for i := 0; i < l; i++ {
		switch t := v.Index(i).Interface().(type) {
		case string:
			result = append(result, t)
		case int64:
			result = append(result, strconv.FormatInt(t, 10))
		case uint64:
			result = append(result, strconv.FormatInt(int64(t), 10))
		case float64:
			result = append(result, strconv.FormatFloat(t, 'f', -1, 64))
		case bool:
			result = append(result, strconv.FormatBool(t))
		default:
			result = append(result, "")
		}
	}

	return result
}

// Int returns a value of int.
func (e *env) Int(key string, defaultValue ...int) int {
	dv := 0

	if len(defaultValue) > 0 {
		dv = defaultValue[0]
	}

	ev := e.Get(key)

	if ev == nil {
		return dv
	}

	switch t := ev.(type) {
	case int64:
		return int(t)
	case uint64:
		return int(t)
	case float64:
		return int(t)
	case string:
		v, _ := strconv.Atoi(t)

		return v
	case bool:
		if t {
			return 1
		}

		return 0
	default:
		return 0
	}
}

// Ints returns a value of []int.
func (e *env) Ints(key string, defaultValue ...int) []int {
	ev := e.Get(key)

	if ev == nil {
		return defaultValue
	}

	v := reflect.Indirect(reflect.ValueOf(ev))

	if v.Kind() != reflect.Slice {
		return defaultValue
	}

	l := v.Len()

	result := make([]int, 0, l)

	if l == 0 {
		return result
	}

	for i := 0; i < l; i++ {
		switch t := v.Index(i).Interface().(type) {
		case int64:
			result = append(result, int(t))
		case uint64:
			result = append(result, int(t))
		case float64:
			result = append(result, int(t))
		case string:
			n, _ := strconv.Atoi(t)
			result = append(result, n)
		case bool:
			if t {
				result = append(result, 1)
			} else {
				result = append(result, 0)
			}
		default:
			result = append(result, 0)
		}
	}

	return result
}

// Int32 returns a value of int32.
func (e *env) Int32(key string, defaultValue ...int32) int32 {
	var dv int32

	if len(defaultValue) > 0 {
		dv = defaultValue[0]
	}

	ev := e.Get(key)

	if ev == nil {
		return dv
	}

	switch t := ev.(type) {
	case int64:
		return int32(t)
	case uint64:
		return int32(t)
	case float64:
		return int32(t)
	case string:
		v, _ := strconv.ParseInt(t, 0, 0)

		return int32(v)
	case bool:
		if t {
			return 1
		}

		return 0
	default:
		return 0
	}
}

// Int32s returns a value of []int32.
func (e *env) Int32s(key string, defaultValue ...int32) []int32 {
	ev := e.Get(key)

	if ev == nil {
		return defaultValue
	}

	v := reflect.Indirect(reflect.ValueOf(ev))

	if v.Kind() != reflect.Slice {
		return defaultValue
	}

	l := v.Len()

	result := make([]int32, 0, l)

	if l == 0 {
		return result
	}

	for i := 0; i < l; i++ {
		switch t := v.Index(i).Interface().(type) {
		case int64:
			result = append(result, int32(t))
		case uint64:
			result = append(result, int32(t))
		case float64:
			result = append(result, int32(t))
		case string:
			v, _ := strconv.ParseInt(t, 0, 0)

			result = append(result, int32(v))
		case bool:
			if t {
				result = append(result, 1)
			} else {
				result = append(result, 0)
			}
		default:
			result = append(result, 0)
		}
	}

	return result
}

// Int64 returns a value of int64.
func (e *env) Int64(key string, defaultValue ...int64) int64 {
	var dv int64

	if len(defaultValue) > 0 {
		dv = defaultValue[0]
	}

	ev := e.Get(key)

	if ev == nil {
		return dv
	}

	switch t := ev.(type) {
	case int64:
		return t
	case uint64:
		return int64(t)
	case float64:
		return int64(t)
	case string:
		v, _ := strconv.ParseInt(t, 0, 0)

		return v
	case bool:
		if t {
			return 1
		}

		return 0
	default:
		return 0
	}
}

// Int64s returns a value of []int64.
func (e *env) Int64s(key string, defaultValue ...int64) []int64 {
	ev := e.Get(key)

	if ev == nil {
		return defaultValue
	}

	v := reflect.Indirect(reflect.ValueOf(ev))

	if v.Kind() != reflect.Slice {
		return []int64{}
	}

	l := v.Len()

	result := make([]int64, 0, l)

	if l == 0 {
		return result
	}

	for i := 0; i < l; i++ {
		switch t := v.Index(i).Interface().(type) {
		case int64:
			result = append(result, t)
		case uint64:
			result = append(result, int64(t))
		case float64:
			result = append(result, int64(t))
		case string:
			n, _ := strconv.ParseInt(t, 10, 64)

			result = append(result, n)
		case bool:
			if t {
				result = append(result, 1)
			} else {
				result = append(result, 0)
			}
		default:
			result = append(result, 0)
		}
	}

	return result
}

// Float64 returns a value of float64.
func (e *env) Float64(key string, defaultValue ...float64) float64 {
	var dv float64

	if len(defaultValue) > 0 {
		dv = defaultValue[0]
	}

	ev := e.Get(key)

	if ev == nil {
		return dv
	}

	switch t := ev.(type) {
	case float64:
		return t
	case int64:
		return float64(t)
	case uint64:
		return float64(t)
	case string:
		v, _ := strconv.ParseFloat(t, 64)

		return v
	case bool:
		if t {
			return 1
		}

		return 0
	default:
		return 0
	}
}

// Float64s returns a value of []float64.
func (e *env) Float64s(key string, defaultValue ...float64) []float64 {
	ev := e.Get(key)

	if ev == nil {
		return defaultValue
	}

	v := reflect.Indirect(reflect.ValueOf(ev))

	if v.Kind() != reflect.Slice {
		return []float64{}
	}

	l := v.Len()

	result := make([]float64, 0, l)

	if l == 0 {
		return result
	}

	for i := 0; i < l; i++ {
		switch t := v.Index(i).Interface().(type) {
		case float64:
			result = append(result, t)
		case int64:
			result = append(result, float64(t))
		case uint64:
			result = append(result, float64(t))
		case string:
			n, _ := strconv.ParseFloat(t, 64)

			result = append(result, n)
		case bool:
			if t {
				result = append(result, 1)
			} else {
				result = append(result, 0)
			}
		default:
			result = append(result, 0)
		}
	}

	return result
}

// Bool returns a value of bool.
func (e *env) Bool(key string, defaultValue ...bool) bool {
	var dv bool

	if len(defaultValue) > 0 {
		dv = defaultValue[0]
	}

	ev := e.Get(key)

	if ev == nil {
		return dv
	}

	switch t := ev.(type) {
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
		v, _ := strconv.ParseBool(t)

		return v
	default:
		return false
	}
}

// Time returns a value of time.Time.
// Layout is required when the env value is a string.
func (e *env) Time(key string, layout string, defaultValue ...time.Time) time.Time {
	var dv time.Time

	if len(defaultValue) > 0 {
		dv = defaultValue[0]
	}

	ev := e.Get(key)

	if ev == nil {
		return dv
	}

	switch t := ev.(type) {
	case time.Time:
		return t
	case string:
		v, _ := time.Parse(layout, t)

		return v
	case int64:
		return time.Unix(t, 0)
	case uint64:
		return time.Unix(int64(t), 0)
	default:
		return time.Time{}
	}
}

// Map returns a value of map[string]interface{}.
func (e *env) Map(key string) map[string]interface{} {
	m := make(map[string]interface{})

	if v, ok := e.Get(key).(*toml.Tree); ok {
		m = v.ToMap()
	}

	return m
}

// Unmarshal attempts to unmarshal the Tree into a Go struct pointed by dest.
func (e *env) Unmarshal(key string, dest interface{}) error {
	i := e.Get(key)

	if i == nil {
		return ErrEnvNil
	}

	if v, ok := i.(*toml.Tree); ok {
		err := v.Unmarshal(dest)

		return err
	}

	return errors.New("yiigo: env config is not a tree of toml")
}

// Get returns the value at key in the env Tree.
func (e *env) Get(key string) interface{} {
	e.mutex.RLock()
	defer e.mutex.RUnlock()

	return e.tree.Get(key)
}
