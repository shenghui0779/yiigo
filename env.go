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

// SetEnvFile use `toml` config file.
func SetEnvFile(file string) error {
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
		r := ""

		switch t := v.Index(i).Interface().(type) {
		case string:
			r = t
		case int64:
			r = strconv.FormatInt(t, 10)
		case uint64:
			r = strconv.FormatInt(int64(t), 10)
		case float64:
			r = strconv.FormatFloat(t, 'f', -1, 64)
		case bool:
			r = strconv.FormatBool(t)
		}

		result = append(result, r)
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
		r := 0

		switch t := v.Index(i).Interface().(type) {
		case int64:
			r = int(t)
		case uint64:
			r = int(t)
		case float64:
			r = int(t)
		case string:
			r, _ = strconv.Atoi(t)
		case bool:
			if t {
				r = 1
			}
		}

		result = append(result, r)
	}

	return result
}

// Uint returns a value of uint.
func (e *env) Uint(key string, defaultValue ...uint) uint {
	var dv uint = 0

	if len(defaultValue) > 0 {
		dv = defaultValue[0]
	}

	ev := e.Get(key)

	if ev == nil {
		return dv
	}

	switch t := ev.(type) {
	case int64:
		if t < 0 {
			return 0
		}

		return uint(t)
	case uint64:
		return uint(t)
	case float64:
		if t < 0 {
			return 0
		}

		return uint(t)
	case string:
		v, _ := strconv.ParseUint(t, 10, 0)

		return uint(v)
	case bool:
		if t {
			return 1
		}

		return 0
	default:
		return 0
	}
}

// Uints returns a value of []uint.
func (e *env) Uints(key string, defaultValue ...uint) []uint {
	ev := e.Get(key)

	if ev == nil {
		return defaultValue
	}

	v := reflect.Indirect(reflect.ValueOf(ev))

	if v.Kind() != reflect.Slice {
		return defaultValue
	}

	l := v.Len()

	result := make([]uint, 0, l)

	if l == 0 {
		return result
	}

	for i := 0; i < l; i++ {
		var r uint

		switch t := v.Index(i).Interface().(type) {
		case int64:
			if t >= 0 {
				r = uint(t)
			}
		case uint64:
			r = uint(t)
		case float64:
			if t > 0 {
				r = uint(t)
			}
		case string:
			n, _ := strconv.ParseUint(t, 10, 0)
			r = uint(n)
		case bool:
			if t {
				r = 1
			}
		}

		result = append(result, r)
	}

	return result
}

// Int8 returns a value of int8.
func (e *env) Int8(key string, defaultValue ...int8) int8 {
	var dv int8

	if len(defaultValue) > 0 {
		dv = defaultValue[0]
	}

	ev := e.Get(key)

	if ev == nil {
		return dv
	}

	switch t := ev.(type) {
	case int64:
		return int8(t)
	case uint64:
		return int8(t)
	case float64:
		return int8(t)
	case string:
		v, _ := strconv.ParseInt(t, 10, 8)

		return int8(v)
	case bool:
		if t {
			return 1
		}

		return 0
	default:
		return 0
	}
}

// Int8s returns a value of []int8.
func (e *env) Int8s(key string, defaultValue ...int8) []int8 {
	ev := e.Get(key)

	if ev == nil {
		return defaultValue
	}

	v := reflect.Indirect(reflect.ValueOf(ev))

	if v.Kind() != reflect.Slice {
		return defaultValue
	}

	l := v.Len()

	result := make([]int8, 0, l)

	if l == 0 {
		return result
	}

	for i := 0; i < l; i++ {
		var r int8

		switch t := v.Index(i).Interface().(type) {
		case int64:
			r = int8(t)
		case uint64:
			r = int8(t)
		case float64:
			r = int8(t)
		case string:
			v, _ := strconv.ParseInt(t, 10, 8)

			r = int8(v)
		case bool:
			if t {
				r = 1
			}
		}

		result = append(result, r)
	}

	return result
}

// Uint8 returns a value of uint8.
func (e *env) Uint8(key string, defaultValue ...uint8) uint8 {
	var dv uint8 = 0

	if len(defaultValue) > 0 {
		dv = defaultValue[0]
	}

	ev := e.Get(key)

	if ev == nil {
		return dv
	}

	switch t := ev.(type) {
	case int64:
		if t < 0 {
			return 0
		}

		return uint8(t)
	case uint64:
		return uint8(t)
	case float64:
		if t < 0 {
			return 0
		}

		return uint8(t)
	case string:
		v, _ := strconv.ParseUint(t, 10, 8)

		return uint8(v)
	case bool:
		if t {
			return 1
		}

		return 0
	default:
		return 0
	}
}

// Uint8s returns a value of []uint8.
func (e *env) Uint8s(key string, defaultValue ...uint8) []uint8 {
	ev := e.Get(key)

	if ev == nil {
		return defaultValue
	}

	v := reflect.Indirect(reflect.ValueOf(ev))

	if v.Kind() != reflect.Slice {
		return defaultValue
	}

	l := v.Len()

	result := make([]uint8, 0, l)

	if l == 0 {
		return result
	}

	for i := 0; i < l; i++ {
		var r uint8

		switch t := v.Index(i).Interface().(type) {
		case int64:
			if t >= 0 {
				r = uint8(t)
			}
		case uint64:
			r = uint8(t)
		case float64:
			if t >= 0 {
				r = uint8(t)
			}
		case string:
			n, _ := strconv.ParseUint(t, 10, 8)
			r = uint8(n)
		case bool:
			if t {
				r = 1
			}
		}

		result = append(result, r)
	}

	return result
}

// Int16 returns a value of int16.
func (e *env) Int16(key string, defaultValue ...int16) int16 {
	var dv int16

	if len(defaultValue) > 0 {
		dv = defaultValue[0]
	}

	ev := e.Get(key)

	if ev == nil {
		return dv
	}

	switch t := ev.(type) {
	case int64:
		return int16(t)
	case uint64:
		return int16(t)
	case float64:
		return int16(t)
	case string:
		v, _ := strconv.ParseInt(t, 10, 16)

		return int16(v)
	case bool:
		if t {
			return 1
		}

		return 0
	default:
		return 0
	}
}

// Int16s returns a value of []int16.
func (e *env) Int16s(key string, defaultValue ...int16) []int16 {
	ev := e.Get(key)

	if ev == nil {
		return defaultValue
	}

	v := reflect.Indirect(reflect.ValueOf(ev))

	if v.Kind() != reflect.Slice {
		return defaultValue
	}

	l := v.Len()

	result := make([]int16, 0, l)

	if l == 0 {
		return result
	}

	for i := 0; i < l; i++ {
		var r int16

		switch t := v.Index(i).Interface().(type) {
		case int64:
			r = int16(t)
		case uint64:
			r = int16(t)
		case float64:
			r = int16(t)
		case string:
			v, _ := strconv.ParseInt(t, 10, 16)
			r = int16(v)
		case bool:
			if t {
				r = 1
			}
		}

		result = append(result, r)
	}

	return result
}

// Uint16 returns a value of uint16.
func (e *env) Uint16(key string, defaultValue ...uint16) uint16 {
	var dv uint16 = 0

	if len(defaultValue) > 0 {
		dv = defaultValue[0]
	}

	ev := e.Get(key)

	if ev == nil {
		return dv
	}

	switch t := ev.(type) {
	case int64:
		if t < 0 {
			return 0
		}

		return uint16(t)
	case uint64:
		return uint16(t)
	case float64:
		if t < 0 {
			return 0
		}

		return uint16(t)
	case string:
		v, _ := strconv.ParseUint(t, 10, 16)

		return uint16(v)
	case bool:
		if t {
			return 1
		}

		return 0
	default:
		return 0
	}
}

// Uint16s returns a value of []uint16.
func (e *env) Uint16s(key string, defaultValue ...uint16) []uint16 {
	ev := e.Get(key)

	if ev == nil {
		return defaultValue
	}

	v := reflect.Indirect(reflect.ValueOf(ev))

	if v.Kind() != reflect.Slice {
		return defaultValue
	}

	l := v.Len()

	result := make([]uint16, 0, l)

	if l == 0 {
		return result
	}

	for i := 0; i < l; i++ {
		var r uint16

		switch t := v.Index(i).Interface().(type) {
		case int64:
			if t >= 0 {
				r = uint16(t)
			}
		case uint64:
			r = uint16(t)
		case float64:
			if t >= 0 {
				r = uint16(t)
			}
		case string:
			n, _ := strconv.ParseUint(t, 10, 16)
			r = uint16(n)
		case bool:
			if t {
				r = 1
			}
		}

		result = append(result, r)
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
		v, _ := strconv.ParseInt(t, 10, 32)

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
		var r int32

		switch t := v.Index(i).Interface().(type) {
		case int64:
			r = int32(t)
		case uint64:
			r = int32(t)
		case float64:
			r = int32(t)
		case string:
			v, _ := strconv.ParseInt(t, 10, 32)

			r = int32(v)
		case bool:
			if t {
				r = 1
			}
		}

		result = append(result, r)
	}

	return result
}

// Uint32 returns a value of uint32.
func (e *env) Uint32(key string, defaultValue ...uint32) uint32 {
	var dv uint32 = 0

	if len(defaultValue) > 0 {
		dv = defaultValue[0]
	}

	ev := e.Get(key)

	if ev == nil {
		return dv
	}

	switch t := ev.(type) {
	case int64:
		if t < 0 {
			return 0
		}

		return uint32(t)
	case uint64:
		return uint32(t)
	case float64:
		if t < 0 {
			return 0
		}

		return uint32(t)
	case string:
		v, _ := strconv.ParseUint(t, 10, 32)

		return uint32(v)
	case bool:
		if t {
			return 1
		}

		return 0
	default:
		return 0
	}
}

// Uint32s returns a value of []uint32.
func (e *env) Uint32s(key string, defaultValue ...uint32) []uint32 {
	ev := e.Get(key)

	if ev == nil {
		return defaultValue
	}

	v := reflect.Indirect(reflect.ValueOf(ev))

	if v.Kind() != reflect.Slice {
		return defaultValue
	}

	l := v.Len()

	result := make([]uint32, 0, l)

	if l == 0 {
		return result
	}

	for i := 0; i < l; i++ {
		var r uint32

		switch t := v.Index(i).Interface().(type) {
		case int64:
			if t >= 0 {
				r = uint32(t)
			}
		case uint64:
			r = uint32(t)
		case float64:
			if t >= 0 {
				r = uint32(t)
			}
		case string:
			n, _ := strconv.ParseUint(t, 10, 32)
			r = uint32(n)
		case bool:
			if t {
				r = 1
			}
		}

		result = append(result, r)
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
		v, _ := strconv.ParseInt(t, 10, 64)

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
		var r int64

		switch t := v.Index(i).Interface().(type) {
		case int64:
			r = t
		case uint64:
			r = int64(t)
		case float64:
			r = int64(t)
		case string:
			r, _ = strconv.ParseInt(t, 10, 64)
		case bool:
			if t {
				r = 1
			}
		}

		result = append(result, r)
	}

	return result
}

// Uint64 returns a value of uint64.
func (e *env) Uint64(key string, defaultValue ...uint64) uint64 {
	var dv uint64 = 0

	if len(defaultValue) > 0 {
		dv = defaultValue[0]
	}

	ev := e.Get(key)

	if ev == nil {
		return dv
	}

	switch t := ev.(type) {
	case int64:
		if t < 0 {
			return 0
		}

		return uint64(t)
	case uint64:
		return t
	case float64:
		if t < 0 {
			return 0
		}

		return uint64(t)
	case string:
		v, _ := strconv.ParseUint(t, 10, 64)

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

// Uint64s returns a value of []uint64.
func (e *env) Uint64s(key string, defaultValue ...uint64) []uint64 {
	ev := e.Get(key)

	if ev == nil {
		return defaultValue
	}

	v := reflect.Indirect(reflect.ValueOf(ev))

	if v.Kind() != reflect.Slice {
		return defaultValue
	}

	l := v.Len()

	result := make([]uint64, 0, l)

	if l == 0 {
		return result
	}

	for i := 0; i < l; i++ {
		var r uint64

		switch t := v.Index(i).Interface().(type) {
		case int64:
			if t >= 0 {
				r = uint64(t)
			}
		case uint64:
			r = uint64(t)
		case float64:
			if t >= 0 {
				r = uint64(t)
			}
		case string:
			r, _ = strconv.ParseUint(t, 10, 64)
		case bool:
			if t {
				r = 1
			}
		}

		result = append(result, r)
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
		var r float64

		switch t := v.Index(i).Interface().(type) {
		case float64:
			r = t
		case int64:
			r = float64(t)
		case uint64:
			r = float64(t)
		case string:
			r, _ = strconv.ParseFloat(t, 64)
		case bool:
			if t {
				r = 1
			}
		}

		result = append(result, r)
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

	v, ok := i.(*toml.Tree)

	if !ok {
		return errors.New("yiigo: invalid config of toml")
	}

	return v.Unmarshal(dest)
}

// Get returns the value at key in the env Tree.
func (e *env) Get(key string) interface{} {
	e.mutex.RLock()
	defer e.mutex.RUnlock()

	return e.tree.Get(key)
}
