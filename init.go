package yiigo

import (
	"path/filepath"
	"sync"
)

var debugMode bool

type cfgenv struct {
	path    string
	options []EnvOption
}

type cfglogger struct {
	name    string
	path    string
	options []LoggerOption
}

type cfgdb struct {
	name    string
	driver  DBDriver
	dsn     string
	options []DBOption
}

type cfgmongo struct {
	name string
	dsn  string
}

type cfgredis struct {
	name    string
	address string
	options []RedisOption
}

type cfgnsq struct {
	nsqd      string
	lookupd   []string
	consumers []NSQConsumer
}

type cfgmailer struct {
	name     string
	host     string
	port     int
	account  string
	password string
}

type initSettings struct {
	debug  bool
	env    *cfgenv
	logger []*cfglogger
	db     []*cfgdb
	mongo  []*cfgmongo
	redis  []*cfgredis
	nsq    *cfgnsq
	mailer []*cfgmailer
}

// InitOption configures how we set up the yiigo initialization.
type InitOption func(s *initSettings)

// WithDebug specifies the debug mode
func WithDebug() InitOption {
	return func(s *initSettings) {
		s.debug = true
	}
}

// WithEnvFile register env file, only support toml.
func WithEnvFile(path string, options ...EnvOption) InitOption {
	return func(s *initSettings) {
		s.env = &cfgenv{
			path:    filepath.Clean(path),
			options: options,
		}
	}
}

// WithLogger register logger
func WithLogger(name, path string, options ...LoggerOption) InitOption {
	return func(s *initSettings) {
		s.logger = append(s.logger, &cfglogger{
			name:    name,
			path:    path,
			options: options,
		})
	}
}

// WithDB register db
// [MySQL] username:password@tcp(localhost:3306)/dbname?timeout=10s&charset=utf8mb4&collation=utf8mb4_general_ci&parseTime=True&loc=Local
// [PgSQL] host=localhost port=5432 user=root password=secret dbname=test connect_timeout=10 sslmode=disable
// [SQLite] file::memory:?cache=shared"
func WithDB(name string, driver DBDriver, dsn string, options ...DBOption) InitOption {
	return func(s *initSettings) {
		s.db = append(s.db, &cfgdb{
			name:    name,
			driver:  driver,
			dsn:     dsn,
			options: options,
		})
	}
}

// WithMongo register mongodb
// mongodb://localhost:27017/?connectTimeoutMS=10000&minPoolSize=10&maxPoolSize=20&maxIdleTimeMS=60000&readPreference=primary
// [reference] https://docs.mongodb.com/manual/reference/connection-string
func WithMongo(name string, dsn string) InitOption {
	return func(s *initSettings) {
		s.mongo = append(s.mongo, &cfgmongo{
			name: name,
			dsn:  dsn,
		})
	}
}

// WithRedis register redis
func WithRedis(name, address string, options ...RedisOption) InitOption {
	return func(s *initSettings) {
		s.redis = append(s.redis, &cfgredis{
			name:    name,
			address: address,
			options: options,
		})
	}
}

// WithNSQ specifies initialize the nsq
func WithNSQ(nsqd string, lookupd []string, consumers ...NSQConsumer) InitOption {
	return func(s *initSettings) {
		s.nsq = &cfgnsq{
			nsqd:      nsqd,
			lookupd:   lookupd,
			consumers: consumers,
		}
	}
}

// Init yiigo initialization
func Init(options ...InitOption) {
	settings := new(initSettings)

	for _, f := range options {
		f(settings)
	}

	debugMode = settings.debug

	var wg sync.WaitGroup

	if settings.env != nil {
		wg.Add(1)

		go func() {
			defer wg.Done()

			initEnv(settings.env.path, settings.env.options...)
		}()
	}

	if len(settings.logger) != 0 {
		wg.Add(1)

		go func() {
			defer wg.Done()

			for _, v := range settings.logger {
				initLogger(v.name, v.path, v.options...)
			}
		}()
	}

	if len(settings.db) != 0 {
		wg.Add(1)

		go func() {
			defer wg.Done()

			for _, v := range settings.db {
				initDB(v.name, v.driver, v.dsn, v.options...)
			}
		}()
	}

	if len(settings.mongo) != 0 {
		wg.Add(1)

		go func() {
			defer wg.Done()

			for _, v := range settings.mongo {
				initMongoDB(v.name, v.dsn)
			}
		}()
	}

	if len(settings.redis) != 0 {
		wg.Add(1)

		go func() {
			defer wg.Done()

			for _, v := range settings.redis {
				initRedis(v.name, v.address, v.options...)
			}
		}()
	}

	if settings.nsq != nil {
		wg.Add(1)

		go func() {
			defer wg.Done()

			initNSQ(settings.nsq.nsqd, settings.nsq.lookupd, settings.nsq.consumers...)
		}()
	}

	if len(settings.mailer) != 0 {
		wg.Add(1)

		go func() {
			defer wg.Done()

			for _, v := range settings.mailer {
				initMailer(v.name, v.host, v.port, v.account, v.account)
			}
		}()
	}

	wg.Wait()
}
