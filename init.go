package yiigo

import (
	"path/filepath"
	"strings"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/nsqio/go-nsq"
)

// InitOption configures how we set up the yiigo initialization.
type InitOption func(wg *sync.WaitGroup)

// WithMySQL register mysql db.
func WithMySQL(name string, cfg *DBConfig) InitOption {
	return func(wg *sync.WaitGroup) {
		defer wg.Done()

		initDB(name, MySQL, cfg)
	}
}

// WithPostgres register postgres db.
func WithPostgres(name string, cfg *DBConfig) InitOption {
	return func(wg *sync.WaitGroup) {
		defer wg.Done()

		initDB(name, Postgres, cfg)
	}
}

// WithSQLite register sqlite db.
func WithSQLite(name string, cfg *DBConfig) InitOption {
	return func(wg *sync.WaitGroup) {
		defer wg.Done()

		initDB(name, SQLite, cfg)
	}
}

// WithMongo register mongodb.
// [DSN] mongodb://localhost:27017/?connectTimeoutMS=10000&minPoolSize=10&maxPoolSize=20&maxIdleTimeMS=60000&readPreference=primary
// [Reference] https://docs.mongodb.com/manual/reference/connection-string
func WithMongo(name string, dsn string) InitOption {
	return func(wg *sync.WaitGroup) {
		defer wg.Done()

		initMongoDB(name, dsn)
	}
}

// WithRedis register redis.
func WithRedis(name string, cfg *RedisConfig) InitOption {
	return func(wg *sync.WaitGroup) {
		defer wg.Done()

		initRedis(name, cfg)
	}
}

// WithLogger register logger.
func WithLogger(name string, cfg *LoggerConfig) InitOption {
	return func(wg *sync.WaitGroup) {
		defer wg.Done()

		if v := strings.TrimSpace(cfg.Filename); len(v) != 0 {
			cfg.Filename = filepath.Clean(v)
		}

		initLogger(name, cfg)
	}
}

// WithNSQProducer specifies the nsq producer.
func WithNSQProducer(nsqd string, cfg *nsq.Config) InitOption {
	return func(wg *sync.WaitGroup) {
		defer wg.Done()

		initNSQProducer(nsqd, cfg)
	}
}

// WithNSQConsumers set the nsq consumers.
func WithNSQConsumers(lookupd []string, consumers ...NSQConsumer) InitOption {
	return func(wg *sync.WaitGroup) {
		defer wg.Done()

		setNSQConsumers(lookupd, consumers...)
	}
}

// WithWebsocket specifies the websocket upgrader.
func WithWebsocket(upgrader *websocket.Upgrader) InitOption {
	return func(wg *sync.WaitGroup) {
		defer wg.Done()

		wsupgrader = upgrader
	}
}

// Init yiigo initialization.
func Init(options ...InitOption) {
	var wg sync.WaitGroup

	for _, f := range options {
		wg.Add(1)

		go f(&wg)
	}

	wg.Wait()
}
