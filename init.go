package yiigo

import (
	"path/filepath"
	"strings"
	"sync"
)

type cfgdb struct {
	name   string
	driver DBDriver
	config *DBConfig
}

type cfgmongo struct {
	name string
	dsn  string
}

type cfgredis struct {
	name   string
	config *RedisConfig
}

type cfgnsq struct {
	nsqd      string
	lookupd   []string
	consumers []NSQConsumer
}

type cfglogger struct {
	name   string
	config *LoggerConfig
}

type initConfig struct {
	logger []*cfglogger
	db     []*cfgdb
	mongo  []*cfgmongo
	redis  []*cfgredis
	nsq    *cfgnsq
}

// InitOption configures how we set up the yiigo initialization.
type InitOption func(c *initConfig)

// WithMySQL register mysql db.
func WithMySQL(name string, cfg *DBConfig) InitOption {
	return func(c *initConfig) {
		c.db = append(c.db, &cfgdb{
			name:   name,
			driver: MySQL,
			config: cfg,
		})
	}
}

// WithPostgres register postgres db.
func WithPostgres(name string, cfg *DBConfig) InitOption {
	return func(c *initConfig) {
		c.db = append(c.db, &cfgdb{
			name:   name,
			driver: Postgres,
			config: cfg,
		})
	}
}

// WithSQLite register sqlite db.
func WithSQLite(name string, cfg *DBConfig) InitOption {
	return func(c *initConfig) {
		c.db = append(c.db, &cfgdb{
			name:   name,
			driver: SQLite,
			config: cfg,
		})
	}
}

// WithMongo register mongodb.
//
// [DSN] mongodb://localhost:27017/?connectTimeoutMS=10000&minPoolSize=10&maxPoolSize=20&maxIdleTimeMS=60000&readPreference=primary
//
// [reference] https://docs.mongodb.com/manual/reference/connection-string
func WithMongo(name string, dsn string) InitOption {
	return func(c *initConfig) {
		c.mongo = append(c.mongo, &cfgmongo{
			name: name,
			dsn:  dsn,
		})
	}
}

// WithRedis register redis.
func WithRedis(name string, cfg *RedisConfig) InitOption {
	return func(c *initConfig) {
		c.redis = append(c.redis, &cfgredis{
			name:   name,
			config: cfg,
		})
	}
}

// WithNSQ initialize nsq.
func WithNSQ(nsqd string, lookupd []string, consumers ...NSQConsumer) InitOption {
	return func(c *initConfig) {
		c.nsq = &cfgnsq{
			nsqd:      nsqd,
			lookupd:   lookupd,
			consumers: consumers,
		}
	}
}

// WithLogger register logger.
func WithLogger(name string, cfg *LoggerConfig) InitOption {
	return func(c *initConfig) {
		logger := &cfglogger{
			name:   name,
			config: cfg,
		}

		if v := strings.TrimSpace(logger.config.Filename); len(v) != 0 {
			logger.config.Filename = filepath.Clean(v)
		}

		c.logger = append(c.logger, logger)
	}
}

// Init yiigo initialization.
func Init(options ...InitOption) {
	cfg := new(initConfig)

	for _, f := range options {
		f(cfg)
	}

	if len(cfg.logger) != 0 {
		for _, v := range cfg.logger {
			initLogger(v.name, v.config)
		}
	}

	var wg sync.WaitGroup

	if len(cfg.db) != 0 {
		wg.Add(1)

		go func() {
			defer wg.Done()

			for _, v := range cfg.db {
				initDB(v.name, v.driver, v.config)
			}
		}()
	}

	if len(cfg.mongo) != 0 {
		wg.Add(1)

		go func() {
			defer wg.Done()

			for _, v := range cfg.mongo {
				initMongoDB(v.name, v.dsn)
			}
		}()
	}

	if len(cfg.redis) != 0 {
		wg.Add(1)

		go func() {
			defer wg.Done()

			for _, v := range cfg.redis {
				initRedis(v.name, v.config)
			}
		}()
	}

	if cfg.nsq != nil {
		wg.Add(1)

		go func() {
			defer wg.Done()

			initNSQ(cfg.nsq.nsqd, cfg.nsq.lookupd, cfg.nsq.consumers...)
		}()
	}

	wg.Wait()
}
