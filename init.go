package yiigo

import (
	"path/filepath"
	"strings"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/nsqio/go-nsq"
)

// InitOption yiigo初始化选项
type InitOption func(wg *sync.WaitGroup)

// WithMySQL 注册MySQL
func WithMySQL(name string, cfg *DBConfig) InitOption {
	return func(wg *sync.WaitGroup) {
		defer wg.Done()

		initDB(name, MySQL, cfg)
	}
}

// WithPostgres 注册Postgres
func WithPostgres(name string, cfg *DBConfig) InitOption {
	return func(wg *sync.WaitGroup) {
		defer wg.Done()

		initDB(name, Postgres, cfg)
	}
}

// WithSQLite 注册SQLite
func WithSQLite(name string, cfg *DBConfig) InitOption {
	return func(wg *sync.WaitGroup) {
		defer wg.Done()

		initDB(name, SQLite, cfg)
	}
}

// WithMongo 注册MongoDB
// [DSN] mongodb://localhost:27017/?connectTimeoutMS=10000&minPoolSize=10&maxPoolSize=20&maxIdleTimeMS=60000&readPreference=primary
// [Reference] https://docs.mongodb.com/manual/reference/connection-string
func WithMongo(name string, dsn string) InitOption {
	return func(wg *sync.WaitGroup) {
		defer wg.Done()

		initMongoDB(name, dsn)
	}
}

// WithRedis 注册Redis
func WithRedis(name string, cfg *RedisConfig) InitOption {
	return func(wg *sync.WaitGroup) {
		defer wg.Done()

		initRedis(name, cfg)
	}
}

// WithLogger 注册日志
func WithLogger(name string, cfg *LoggerConfig) InitOption {
	return func(wg *sync.WaitGroup) {
		defer wg.Done()

		if v := strings.TrimSpace(cfg.Filename); len(v) != 0 {
			cfg.Filename = filepath.Clean(v)
		}

		initLogger(name, cfg)
	}
}

// WithNSQProducer 设置nsq生产者
func WithNSQProducer(nsqd string, cfg *nsq.Config) InitOption {
	return func(wg *sync.WaitGroup) {
		defer wg.Done()

		initNSQProducer(nsqd, cfg)
	}
}

// WithNSQConsumers 设置nsq消费者
func WithNSQConsumers(lookupd []string, consumers ...NSQConsumer) InitOption {
	return func(wg *sync.WaitGroup) {
		defer wg.Done()

		setNSQConsumers(lookupd, consumers...)
	}
}

// WithWebsocket 设置websocket
func WithWebsocket(upgrader *websocket.Upgrader) InitOption {
	return func(wg *sync.WaitGroup) {
		defer wg.Done()

		wsupgrader = upgrader
	}
}

// Init yiigo初始化
func Init(options ...InitOption) {
	var wg sync.WaitGroup

	for _, f := range options {
		wg.Add(1)

		go f(&wg)
	}

	wg.Wait()
}
