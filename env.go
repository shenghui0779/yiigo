package yiigo

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/pelletier/go-toml"
	"github.com/philchia/agollo/v3"
	"go.uber.org/zap"
)

// ErrConfigNil returned when config not found.
var ErrConfigNil = errors.New("yiigo: config not found")

type config struct {
	tree   *toml.Tree
	apollo *apollo
	mutex  sync.RWMutex
}

type apollo struct {
	namespace []string
	debug     bool
}

func (c *config) get(key string) interface{} {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	return c.tree.Get(key)
}

func (c *config) getFromApollo(key string) string {
	if c.apollo == nil || c.apollo.debug || key == "" {
		return ""
	}

	arr := strings.Split(key, ".")

	if len(arr) == 1 || !InStrings(arr[0], c.apollo.namespace...) {
		return agollo.GetStringValue(arr[0], "")
	}

	return agollo.GetStringValueWithNameSpace(arr[0], arr[1], "")
}

func (c *config) setApollo(namespace []string, debug bool) {
	c.apollo = &apollo{
		namespace: namespace,
		debug:     debug,
	}
}

// Env config value
type EnvValue struct {
	value interface{}
}

// String returns a value of string.
func (e *EnvValue) String(defaultValue ...string) string {
	dv := ""

	if len(defaultValue) > 0 {
		dv = defaultValue[0]
	}

	if e.value == nil {
		return dv
	}

	switch t := e.value.(type) {
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
func (e *EnvValue) Strings(defaultValue ...string) []string {
	if e.value == nil {
		return defaultValue
	}

	v := reflect.Indirect(reflect.ValueOf(e.value))

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
func (e *EnvValue) Int(defaultValue ...int) int {
	dv := 0

	if len(defaultValue) > 0 {
		dv = defaultValue[0]
	}

	if e.value == nil {
		return dv
	}

	switch t := e.value.(type) {
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
func (e *EnvValue) Ints(defaultValue ...int) []int {
	if e.value == nil {
		return defaultValue
	}

	v := reflect.Indirect(reflect.ValueOf(e.value))

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
func (e *EnvValue) Uint(defaultValue ...uint) uint {
	var dv uint

	if len(defaultValue) > 0 {
		dv = defaultValue[0]
	}

	if e.value == nil {
		return dv
	}

	switch t := e.value.(type) {
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
func (e *EnvValue) Uints(defaultValue ...uint) []uint {
	if e.value == nil {
		return defaultValue
	}

	v := reflect.Indirect(reflect.ValueOf(e.value))

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
func (e *EnvValue) Int8(defaultValue ...int8) int8 {
	var dv int8

	if len(defaultValue) > 0 {
		dv = defaultValue[0]
	}

	if e.value == nil {
		return dv
	}

	switch t := e.value.(type) {
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
func (e *EnvValue) Int8s(defaultValue ...int8) []int8 {
	if e.value == nil {
		return defaultValue
	}

	v := reflect.Indirect(reflect.ValueOf(e.value))

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
func (e *EnvValue) Uint8(defaultValue ...uint8) uint8 {
	var dv uint8

	if len(defaultValue) > 0 {
		dv = defaultValue[0]
	}

	if e.value == nil {
		return dv
	}

	switch t := e.value.(type) {
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
func (e *EnvValue) Uint8s(defaultValue ...uint8) []uint8 {
	if e.value == nil {
		return defaultValue
	}

	v := reflect.Indirect(reflect.ValueOf(e.value))

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
func (e *EnvValue) Int16(defaultValue ...int16) int16 {
	var dv int16

	if len(defaultValue) > 0 {
		dv = defaultValue[0]
	}

	if e.value == nil {
		return dv
	}

	switch t := e.value.(type) {
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
func (e *EnvValue) Int16s(defaultValue ...int16) []int16 {
	if e.value == nil {
		return defaultValue
	}

	v := reflect.Indirect(reflect.ValueOf(e.value))

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
func (e *EnvValue) Uint16(defaultValue ...uint16) uint16 {
	var dv uint16

	if len(defaultValue) > 0 {
		dv = defaultValue[0]
	}

	if e.value == nil {
		return dv
	}

	switch t := e.value.(type) {
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
func (e *EnvValue) Uint16s(defaultValue ...uint16) []uint16 {
	if e.value == nil {
		return defaultValue
	}

	v := reflect.Indirect(reflect.ValueOf(e.value))

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
func (e *EnvValue) Int32(defaultValue ...int32) int32 {
	var dv int32

	if len(defaultValue) > 0 {
		dv = defaultValue[0]
	}

	if e.value == nil {
		return dv
	}

	switch t := e.value.(type) {
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
func (e *EnvValue) Int32s(defaultValue ...int32) []int32 {
	if e.value == nil {
		return defaultValue
	}

	v := reflect.Indirect(reflect.ValueOf(e.value))

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
func (e *EnvValue) Uint32(defaultValue ...uint32) uint32 {
	var dv uint32

	if len(defaultValue) > 0 {
		dv = defaultValue[0]
	}

	if e.value == nil {
		return dv
	}

	switch t := e.value.(type) {
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
func (e *EnvValue) Uint32s(defaultValue ...uint32) []uint32 {
	if e.value == nil {
		return defaultValue
	}

	v := reflect.Indirect(reflect.ValueOf(e.value))

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
func (e *EnvValue) Int64(defaultValue ...int64) int64 {
	var dv int64

	if len(defaultValue) > 0 {
		dv = defaultValue[0]
	}

	if e.value == nil {
		return dv
	}

	switch t := e.value.(type) {
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
func (e *EnvValue) Int64s(defaultValue ...int64) []int64 {
	if e.value == nil {
		return defaultValue
	}

	v := reflect.Indirect(reflect.ValueOf(e.value))

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
func (e *EnvValue) Uint64(defaultValue ...uint64) uint64 {
	var dv uint64

	if len(defaultValue) > 0 {
		dv = defaultValue[0]
	}

	if e.value == nil {
		return dv
	}

	switch t := e.value.(type) {
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
func (e *EnvValue) Uint64s(defaultValue ...uint64) []uint64 {
	if e.value == nil {
		return defaultValue
	}

	v := reflect.Indirect(reflect.ValueOf(e.value))

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
			r = t
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
func (e *EnvValue) Float64(defaultValue ...float64) float64 {
	var dv float64

	if len(defaultValue) > 0 {
		dv = defaultValue[0]
	}

	if e.value == nil {
		return dv
	}

	switch t := e.value.(type) {
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
func (e *EnvValue) Float64s(defaultValue ...float64) []float64 {
	if e.value == nil {
		return defaultValue
	}

	v := reflect.Indirect(reflect.ValueOf(e.value))

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
func (e *EnvValue) Bool(defaultValue ...bool) bool {
	var dv bool

	if len(defaultValue) > 0 {
		dv = defaultValue[0]
	}

	if e.value == nil {
		return dv
	}

	switch t := e.value.(type) {
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
func (e *EnvValue) Time(layout string, defaultValue ...time.Time) time.Time {
	var dv time.Time

	if len(defaultValue) > 0 {
		dv = defaultValue[0]
	}

	if e.value == nil {
		return dv
	}

	switch t := e.value.(type) {
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
func (e *EnvValue) Map() map[string]interface{} {
	m := make(map[string]interface{})

	if e.value == nil {
		return m
	}

	if v, ok := e.value.(*toml.Tree); ok {
		m = v.ToMap()
	}

	return m
}

// Unmarshal attempts to unmarshal the Tree into a Go struct pointed by dest.
func (e *EnvValue) Unmarshal(dest interface{}) error {
	if e.value == nil {
		return ErrConfigNil
	}

	v, ok := e.value.(*toml.Tree)

	if !ok {
		return errors.New("yiigo: toml syntax error")
	}

	return v.Unmarshal(dest)
}

var env *config

func loadConfigFile() {
	path, err := filepath.Abs("yiigo.toml")

	if err != nil {
		logger.Panic("yiigo: load config file error", zap.Error(err))
	}

	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			if f, err := os.Create(path); err == nil {
				f.WriteString(defaultEnvContent)
				f.Close()
			}
		} else if os.IsPermission(err) {
			os.Chmod(path, os.ModePerm)
		}
	}

	t, err := toml.LoadFile(path)

	if err != nil {
		logger.Panic("yiigo: load config file error", zap.Error(err))
	}

	env = &config{tree: t}
}

// Env returns an env value
func Env(key string) *EnvValue {
	if v := env.getFromApollo(key); v != "" {
		return &EnvValue{value: v}
	}

	return &EnvValue{value: env.get(key)}
}

var defaultEnvContent = `[app]
env = "dev" # dev | beta | prod
debug = true

[apollo]
# appid = "test"
# cluster = "default"
# address = "127.0.0.1:8080"
# namespace = []
# cache_dir = "./"
# accesskey_secret = ""
# insecure_skip_verify = true

[db]

    # [db.default]
    # driver = "mysql"
    # dsn = "username:password@tcp(localhost:3306)/dbname?timeout=10s&charset=utf8mb4&collation=utf8mb4_general_ci&parseTime=True&loc=Local"
    # max_open_conns = 20
    # max_idle_conns = 10
    # conn_max_lifetime = 60 # 秒

[mongo]

	# [mongo.default]
	# dsn = "mongodb://username:password@localhost:27017"
	# connect_timeout = 10 # 秒
	# min_pool_size = 10
	# max_pool_size = 10
	# max_conn_idle_time = 60 # 秒
	# mode = "primary" # primary | primary_preferred | secondary | secondary_preferred | nearest

[redis]

	# [redis.default]
	# address = "127.0.0.1:6379"
	# password = ""
	# database = 0
	# connect_timeout = 10 # 秒
	# read_timeout = 10 # 秒
	# write_timeout = 10 # 秒
	# pool_size = 10
	# pool_limit = 20
	# idle_timeout = 60 # 秒
	# wait_timeout = 10 # 秒
	# prefill_parallelism = 0

[email]

	# [email.default]
	# host = "smtp.exmail.qq.com"
	# port = 25
	# username = ""
	# password = ""

[log]

    [log.default]
    path = "logs/app.log"
    max_size = 500
    max_age = 0
    max_backups = 0
    compress = true
`
