package yiigo

import (
	"errors"
	"os"
	"path/filepath"
	"runtime/debug"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/pelletier/go-toml"
	"go.uber.org/zap"
)

// Environment is the interface for config
type Environment interface {
	// Get returns an env value
	Get(key string) EnvValue

	// LoadEnvFromFile loads env from file
	LoadEnvFromFile(path string) error

	// LoadEnvFromBytes loads env from bytes
	LoadEnvFromBytes(b []byte) error

	// Reload reloads env config
	Reload() error

	// Watcher watching env change and reload
	Watcher(onchange func(event fsnotify.Event))
}

// EnvValue is the interface for config value
type EnvValue interface {
	// Int returns a value of int64.
	Int(defaultValue ...int64) int64

	// Ints returns a value of []int64.
	Ints(defaultValue ...int64) []int64

	// Float returns a value of float64.
	Float(defaultValue ...float64) float64

	// Floats returns a value of []float64.
	Floats(defaultValue ...float64) []float64

	// String returns a value of string.
	String(defaultValue ...string) string

	// Strings returns a value of []string.
	Strings(defaultValue ...string) []string

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
	mutex sync.RWMutex
	tree  *toml.Tree
	path  string
}

func (c *config) Get(key string) EnvValue {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	if c.tree == nil {
		return new(cfgValue)
	}

	return &cfgValue{value: c.tree.Get(key)}
}

func (c *config) LoadEnvFromFile(path string) error {
	t, err := toml.LoadFile(path)

	if err != nil {
		return err
	}

	c.tree = t
	c.path = path

	return nil
}

func (c *config) LoadEnvFromBytes(b []byte) error {
	t, err := toml.LoadBytes(b)

	if err != nil {
		return err
	}

	c.tree = t

	return nil
}

func (c *config) Reload() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if len(c.path) == 0 {
		return nil
	}

	t, err := toml.LoadFile(c.path)

	if err != nil {
		return err
	}

	c.tree = t

	return nil
}

func (c *config) Watcher(onchange func(event fsnotify.Event)) {
	watcher, err := fsnotify.NewWatcher()

	if err != nil {
		logger.Error("yiigo: env watcher error", zap.Error(err))

		return
	}

	defer watcher.Close()

	envDir, _ := filepath.Split(c.path)
	realEnvFile, _ := filepath.EvalSymlinks(c.path)

	var wg sync.WaitGroup

	wg.Add(1)

	go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.Error("yiigo: env watcher panic", zap.Any("error", r), zap.ByteString("stack", debug.Stack()))
			}

			wg.Done()
		}()

		writeOrCreateMask := fsnotify.Write | fsnotify.Create

		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok { // 'Events' channel is closed
					return
				}

				eventFile := filepath.Clean(event.Name)
				currentEnvFile, _ := filepath.EvalSymlinks(c.path)

				// the env file was modified or created || the real path to the env file changed (eg: k8s ConfigMap replacement)
				if (eventFile == c.path && event.Op&writeOrCreateMask != 0) || (currentEnvFile != "" && currentEnvFile != realEnvFile) {
					realEnvFile = currentEnvFile

					if err := c.Reload(); err != nil {
						logger.Error("yiigo: env reload error", zap.Error(err))
					}

					// reassign the 'Debug' variable
					Debug = c.Get("app.debug").Bool()

					if onchange != nil {
						onchange(event)
					}
				} else if eventFile == c.path && event.Op&fsnotify.Remove&fsnotify.Remove != 0 {
					logger.Warn("yiigo: env file removed")
				}
			case err, ok := <-watcher.Errors:
				if ok { // 'Errors' channel is not closed
					logger.Error("yiigo: env watcher error", zap.Error(err))
				}

				return
			}
		}
	}()

	watcher.Add(envDir)

	wg.Wait()
}

type cfgValue struct {
	value interface{}
}

func (c *cfgValue) Int(defaultValue ...int64) int64 {
	var dv int64

	if len(defaultValue) != 0 {
		dv = defaultValue[0]
	}

	if c.value == nil {
		return dv
	}

	result, ok := c.value.(int64)

	if !ok {
		return 0
	}

	return result
}

func (c *cfgValue) Ints(defaultValue ...int64) []int64 {
	if c.value == nil {
		return defaultValue
	}

	arr, ok := c.value.([]interface{})

	if !ok {
		return []int64{}
	}

	l := len(arr)

	result := make([]int64, 0, l)

	for _, v := range arr {
		if i, ok := v.(int64); ok {
			result = append(result, i)
		}
	}

	if len(result) < l {
		return []int64{}
	}

	return result
}

func (c *cfgValue) Float(defaultValue ...float64) float64 {
	var dv float64

	if len(defaultValue) != 0 {
		dv = defaultValue[0]
	}

	if c.value == nil {
		return dv
	}

	result, ok := c.value.(float64)

	if !ok {
		return 0
	}

	return result
}

func (c *cfgValue) Floats(defaultValue ...float64) []float64 {
	if c.value == nil {
		return defaultValue
	}

	arr, ok := c.value.([]interface{})

	if !ok {
		return []float64{}
	}

	l := len(arr)

	result := make([]float64, 0, l)

	for _, v := range arr {
		if f, ok := v.(float64); ok {
			result = append(result, f)
		}
	}

	if len(result) < l {
		return []float64{}
	}

	return result
}

func (c *cfgValue) String(defaultValue ...string) string {
	dv := ""

	if len(defaultValue) != 0 {
		dv = defaultValue[0]
	}

	if c.value == nil {
		return dv
	}

	result, ok := c.value.(string)

	if !ok {
		return ""
	}

	return result
}

func (c *cfgValue) Strings(defaultValue ...string) []string {
	if c.value == nil {
		return defaultValue
	}

	arr, ok := c.value.([]interface{})

	if !ok {
		return []string{}
	}

	l := len(arr)

	result := make([]string, 0, l)

	for _, v := range arr {
		if s, ok := v.(string); ok {
			result = append(result, s)
		}
	}

	if len(result) < l {
		return []string{}
	}

	return result
}

func (c *cfgValue) Bool(defaultValue ...bool) bool {
	var dv bool

	if len(defaultValue) != 0 {
		dv = defaultValue[0]
	}

	if c.value == nil {
		return dv
	}

	result, ok := c.value.(bool)

	if !ok {
		return false
	}

	return result
}

func (c *cfgValue) Time(layout string, defaultValue ...time.Time) time.Time {
	var dv time.Time

	if len(defaultValue) != 0 {
		dv = defaultValue[0]
	}

	if c.value == nil {
		return dv
	}

	var result time.Time

	switch t := c.value.(type) {
	case time.Time:
		result = t
	case string:
		result, _ = time.Parse(layout, t)
	}

	return result
}

func (c *cfgValue) Map() X {
	if c.value == nil {
		return X{}
	}

	v, ok := c.value.(*toml.Tree)

	if !ok {
		return X{}
	}

	return v.ToMap()
}

func (c *cfgValue) Unmarshal(dest interface{}) error {
	if c.value == nil {
		return nil
	}

	v, ok := c.value.(*toml.Tree)

	if !ok {
		return errors.New("yiigo: invalid env value, expects *toml.Tree")
	}

	return v.Unmarshal(dest)
}

var env Environment = new(config)

func initEnv(settings *initSettings) {
	path, err := filepath.Abs(filepath.Join(settings.envDir, "yiigo.toml"))

	if err != nil {
		logger.Panic("yiigo: load config file error", zap.Error(err))
	}

	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			if len(settings.envDir) != 0 {
				if err := os.MkdirAll(settings.envDir, 0755); err != nil {
					logger.Panic("yiigo: load config file error", zap.Error(err))
				}
			}

			f, err := os.Create(path)

			if err != nil {
				logger.Panic("yiigo: load config file error", zap.Error(err))
			}

			f.WriteString(defaultEnvContent)
			f.Close()
		} else if os.IsPermission(err) {
			os.Chmod(path, 0755)
		}
	}

	if err := env.LoadEnvFromFile(path); err != nil {
		logger.Panic("yiigo: load config file error", zap.Error(err))
	}

	// assign the 'Debug' variable
	Debug = Env("app.debug").Bool()

	if settings.envWatcher {
		go env.Watcher(settings.envOnChange)
	}
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
	# dsn = "mongodb://localhost:27017/?connectTimeoutMS=10000&minPoolSize=10&maxPoolSize=20&maxIdleTimeMS=60000&readPreference=primary"

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
