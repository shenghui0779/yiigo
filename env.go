package yiigo

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"sync"
	"time"

	"github.com/pelletier/go-toml"
	"go.uber.org/zap"
)

// Environment is the interface for config
type Environment interface {
	// Get returns an env value
	Get(key string) EnvValue
}

// EnvValue is the interface for config value
type EnvValue interface {
	// String returns a value of string.
	String(defaultValue ...string) string

	// Strings returns a value of []string.
	Strings(defaultValue ...string) []string

	// Int returns a value of int.
	Int(defaultValue ...int) int

	// Ints returns a value of []int.
	Ints(defaultValue ...int) []int

	// Uint returns a value of uint.
	Uint(defaultValue ...uint) uint

	// Uints returns a value of []uint.
	Uints(defaultValue ...uint) []uint

	// Int8 returns a value of int8.
	Int8(defaultValue ...int8) int8

	// Int8s returns a value of []int8.
	Int8s(defaultValue ...int8) []int8

	// Uint8 returns a value of uint8.
	Uint8(defaultValue ...uint8) uint8

	// Uint8s returns a value of []uint8.
	Uint8s(defaultValue ...uint8) []uint8

	// Int16 returns a value of int16.
	Int16(defaultValue ...int16) int16

	// Int16s returns a value of []int16.
	Int16s(defaultValue ...int16) []int16

	// Uint16 returns a value of uint16.
	Uint16(defaultValue ...uint16) uint16

	// Uint16s returns a value of []uint16.
	Uint16s(defaultValue ...uint16) []uint16

	// Int32 returns a value of int32.
	Int32(defaultValue ...int32) int32

	// Int32s returns a value of []int32.
	Int32s(defaultValue ...int32) []int32

	// Uint32 returns a value of uint32.
	Uint32(defaultValue ...uint32) uint32

	// Uint32s returns a value of []uint32.
	Uint32s(defaultValue ...uint32) []uint32

	// Int64 returns a value of int64.
	Int64(defaultValue ...int64) int64

	// Int64s returns a value of []int64.
	Int64s(defaultValue ...int64) []int64

	// Uint64 returns a value of uint64.
	Uint64(defaultValue ...uint64) uint64

	// Uint64s returns a value of []uint64.
	Uint64s(defaultValue ...uint64) []uint64

	// Float64 returns a value of float64.
	Float64(defaultValue ...float64) float64

	// Float64s returns a value of []float64.
	Float64s(defaultValue ...float64) []float64

	// Bool returns a value of bool.
	Bool(defaultValue ...bool) bool

	// Time returns a value of time.Time.
	// Layout is required when the env value is a string.
	Time(layout string, defaultValue ...time.Time) time.Time

	// Map returns a value of X.
	Map() X

	// Unmarshal attempts to unmarshal the value into a Go struct pointed by dest.
	Unmarshal(dest interface{}) error
}

type config struct {
	tree  *toml.Tree
	mutex sync.RWMutex
}

func (c *config) Get(key string) EnvValue {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	return &configValue{value: c.tree.Get(key)}
}

type configValue struct {
	value interface{}
}

func (c *configValue) String(defaultValue ...string) string {
	dv := ""

	if len(defaultValue) != 0 {
		dv = defaultValue[0]
	}

	if c.value == nil {
		return dv
	}

	switch t := c.value.(type) {
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

func (c *configValue) Strings(defaultValue ...string) []string {
	if c.value == nil {
		return defaultValue
	}

	v := reflect.Indirect(reflect.ValueOf(c.value))

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

func (c *configValue) Int(defaultValue ...int) int {
	dv := 0

	if len(defaultValue) != 0 {
		dv = defaultValue[0]
	}

	if c.value == nil {
		return dv
	}

	switch t := c.value.(type) {
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

func (c *configValue) Ints(defaultValue ...int) []int {
	if c.value == nil {
		return defaultValue
	}

	v := reflect.Indirect(reflect.ValueOf(c.value))

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

func (c *configValue) Uint(defaultValue ...uint) uint {
	var dv uint

	if len(defaultValue) != 0 {
		dv = defaultValue[0]
	}

	if c.value == nil {
		return dv
	}

	switch t := c.value.(type) {
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

func (c *configValue) Uints(defaultValue ...uint) []uint {
	if c.value == nil {
		return defaultValue
	}

	v := reflect.Indirect(reflect.ValueOf(c.value))

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

func (c *configValue) Int8(defaultValue ...int8) int8 {
	var dv int8

	if len(defaultValue) != 0 {
		dv = defaultValue[0]
	}

	if c.value == nil {
		return dv
	}

	switch t := c.value.(type) {
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

func (c *configValue) Int8s(defaultValue ...int8) []int8 {
	if c.value == nil {
		return defaultValue
	}

	v := reflect.Indirect(reflect.ValueOf(c.value))

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

func (c *configValue) Uint8(defaultValue ...uint8) uint8 {
	var dv uint8

	if len(defaultValue) != 0 {
		dv = defaultValue[0]
	}

	if c.value == nil {
		return dv
	}

	switch t := c.value.(type) {
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

func (c *configValue) Uint8s(defaultValue ...uint8) []uint8 {
	if c.value == nil {
		return defaultValue
	}

	v := reflect.Indirect(reflect.ValueOf(c.value))

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

func (c *configValue) Int16(defaultValue ...int16) int16 {
	var dv int16

	if len(defaultValue) != 0 {
		dv = defaultValue[0]
	}

	if c.value == nil {
		return dv
	}

	switch t := c.value.(type) {
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

func (c *configValue) Int16s(defaultValue ...int16) []int16 {
	if c.value == nil {
		return defaultValue
	}

	v := reflect.Indirect(reflect.ValueOf(c.value))

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

func (c *configValue) Uint16(defaultValue ...uint16) uint16 {
	var dv uint16

	if len(defaultValue) != 0 {
		dv = defaultValue[0]
	}

	if c.value == nil {
		return dv
	}

	switch t := c.value.(type) {
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

func (c *configValue) Uint16s(defaultValue ...uint16) []uint16 {
	if c.value == nil {
		return defaultValue
	}

	v := reflect.Indirect(reflect.ValueOf(c.value))

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

func (c *configValue) Int32(defaultValue ...int32) int32 {
	var dv int32

	if len(defaultValue) != 0 {
		dv = defaultValue[0]
	}

	if c.value == nil {
		return dv
	}

	switch t := c.value.(type) {
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

func (c *configValue) Int32s(defaultValue ...int32) []int32 {
	if c.value == nil {
		return defaultValue
	}

	v := reflect.Indirect(reflect.ValueOf(c.value))

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

func (c *configValue) Uint32(defaultValue ...uint32) uint32 {
	var dv uint32

	if len(defaultValue) != 0 {
		dv = defaultValue[0]
	}

	if c.value == nil {
		return dv
	}

	switch t := c.value.(type) {
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

func (c *configValue) Uint32s(defaultValue ...uint32) []uint32 {
	if c.value == nil {
		return defaultValue
	}

	v := reflect.Indirect(reflect.ValueOf(c.value))

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

func (c *configValue) Int64(defaultValue ...int64) int64 {
	var dv int64

	if len(defaultValue) != 0 {
		dv = defaultValue[0]
	}

	if c.value == nil {
		return dv
	}

	switch t := c.value.(type) {
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

func (c *configValue) Int64s(defaultValue ...int64) []int64 {
	if c.value == nil {
		return defaultValue
	}

	v := reflect.Indirect(reflect.ValueOf(c.value))

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

func (c *configValue) Uint64(defaultValue ...uint64) uint64 {
	var dv uint64

	if len(defaultValue) != 0 {
		dv = defaultValue[0]
	}

	if c.value == nil {
		return dv
	}

	switch t := c.value.(type) {
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

func (c *configValue) Uint64s(defaultValue ...uint64) []uint64 {
	if c.value == nil {
		return defaultValue
	}

	v := reflect.Indirect(reflect.ValueOf(c.value))

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

func (c *configValue) Float64(defaultValue ...float64) float64 {
	var dv float64

	if len(defaultValue) != 0 {
		dv = defaultValue[0]
	}

	if c.value == nil {
		return dv
	}

	switch t := c.value.(type) {
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

func (c *configValue) Float64s(defaultValue ...float64) []float64 {
	if c.value == nil {
		return defaultValue
	}

	v := reflect.Indirect(reflect.ValueOf(c.value))

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

func (c *configValue) Bool(defaultValue ...bool) bool {
	var dv bool

	if len(defaultValue) != 0 {
		dv = defaultValue[0]
	}

	if c.value == nil {
		return dv
	}

	switch t := c.value.(type) {
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

func (c *configValue) Time(layout string, defaultValue ...time.Time) time.Time {
	var dv time.Time

	if len(defaultValue) != 0 {
		dv = defaultValue[0]
	}

	if c.value == nil {
		return dv
	}

	switch t := c.value.(type) {
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

func (c *configValue) Map() X {
	if c.value == nil {
		return X{}
	}

	v, ok := c.value.(*toml.Tree)

	if !ok {
		return X{}
	}

	return v.ToMap()
}

func (c *configValue) Unmarshal(dest interface{}) error {
	if c.value == nil {
		return nil
	}

	v, ok := c.value.(*toml.Tree)

	if !ok {
		return errors.New("yiigo: toml syntax error")
	}

	return v.Unmarshal(dest)
}

var env Environment

func initEnv() {
	if err := LoadEnvFromFile("yiigo.toml"); err != nil {
		logger.Panic("yiigo: load config file error", zap.Error(err))
	}
}

// LoadEnvFromFile load env from file
func LoadEnvFromFile(path string) error {
	path, err := filepath.Abs(path)

	if err != nil {
		return err
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
		return err
	}

	env = &config{tree: t}

	return nil
}

// LoadEnvFromBytes load env from bytes
func LoadEnvFromBytes(b []byte) error {
	t, err := toml.LoadBytes(b)

	if err != nil {
		return err
	}

	env = &config{tree: t}

	return nil
}

// Env returns an env value
func Env(key string) EnvValue {
	return env.Get(key)
}

var defaultEnvContent = `[app]
env = "dev"
debug = true

[db]

    # [db.default]
    # driver = "mysql"
    # dsn = "username:password@tcp(localhost:3306)/dbname?timeout=10s&charset=utf8mb4&collation=utf8mb4_general_ci&parseTime=True&loc=Local"
    # max_open_conns = 20
	# max_idle_conns = 10
	# conn_max_idle_time = 60
    # conn_max_lifetime = 600

[mongo]

	# [mongo.default]
	# dsn = "mongodb://username:password@localhost:27017"
	# connect_timeout = 10
	# min_pool_size = 10
	# max_pool_size = 20
	# max_conn_idle_time = 60
	# mode = "primary"

[redis]

	# [redis.default]
	# address = "127.0.0.1:6379"
	# password = ""
	# database = 0
	# connect_timeout = 10
	# read_timeout = 10
	# write_timeout = 10
	# pool_size = 10
	# pool_limit = 20
	# idle_timeout = 60
	# wait_timeout = 10
	# prefill_parallelism = 0

# [nsq]
# lookupd = ["127.0.0.1:4161"]
# nsqd = "127.0.0.1:4150"

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
