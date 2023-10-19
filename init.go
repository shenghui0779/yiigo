package yiigo

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/gorilla/websocket"
	"github.com/nsqio/go-nsq"
	"go.uber.org/zap"
)

// InitOption yiigo初始化选项
type InitOption func()

// WithMySQL 注册MySQL
func WithMySQL(name string, cfg *DBConfig) InitOption {
	return func() {
		if len(name) == 0 || cfg == nil {
			return
		}

		if err := initDB(name, MySQL, cfg); err != nil {
			logger.Panic(fmt.Sprintf("err mysql.%s init", name), zap.String("dsn", cfg.DSN), zap.Error(err))
		}

		logger.Info(fmt.Sprintf("mysql.%s is OK", name))
	}
}

// WithPostgres 注册Postgres
func WithPostgres(name string, cfg *DBConfig) InitOption {
	return func() {
		if len(name) == 0 || cfg == nil {
			return
		}

		if err := initDB(name, Postgres, cfg); err != nil {
			logger.Panic(fmt.Sprintf("err postgres.%s init", name), zap.String("dsn", cfg.DSN), zap.Error(err))
		}

		logger.Info(fmt.Sprintf("postgres.%s is OK", name))
	}
}

// WithSQLite 注册SQLite
func WithSQLite(name string, cfg *DBConfig) InitOption {
	return func() {
		if len(name) == 0 || cfg == nil {
			return
		}

		if err := initDB(name, SQLite, cfg); err != nil {
			logger.Panic(fmt.Sprintf("err sqlite.%s init", name), zap.String("dsn", cfg.DSN), zap.Error(err))
		}

		logger.Info(fmt.Sprintf("sqlite.%s is OK", name))
	}
}

// WithMongo 注册MongoDB
// [DSN] mongodb://localhost:27017/?connectTimeoutMS=10000&minPoolSize=10&maxPoolSize=20&maxIdleTimeMS=60000&readPreference=primary
// [Reference] https://docs.mongodb.com/manual/reference/connection-string
func WithMongo(name string, dsn string) InitOption {
	return func() {
		if len(name) == 0 || len(dsn) == 0 {
			return
		}

		if err := initMongoDB(name, dsn); err != nil {
			logger.Panic(fmt.Sprintf("err mongodb.%s init", name), zap.String("dsn", dsn), zap.Error(err))
		}

		logger.Info(fmt.Sprintf("mongodb.%s is OK", name))
	}
}

// WithRedis 注册Redis
func WithRedis(name string, cfg *RedisConfig) InitOption {
	return func() {
		if len(name) == 0 || cfg == nil {
			return
		}

		if err := initRedis(name, cfg); err != nil {
			logger.Panic(fmt.Sprintf("err redis.%s init", name), zap.String("addr", cfg.Addr), zap.Error(err))
		}

		logger.Info(fmt.Sprintf("redis.%s is OK", name))
	}
}

// WithLogger 注册日志
func WithLogger(name string, cfg *LoggerConfig) InitOption {
	return func() {
		if len(name) == 0 || cfg == nil {
			return
		}

		if v := strings.TrimSpace(cfg.Filename); len(v) != 0 {
			cfg.Filename = filepath.Clean(v)
		}

		initLogger(name, cfg)

		logger.Info(fmt.Sprintf("logger.%s is OK", name))
	}
}

// WithNSQProducer 设置nsq生产者
func WithNSQProducer(nsqd string, cfg *nsq.Config) InitOption {
	return func() {
		if err := initNSQProducer(nsqd, cfg); err != nil {
			logger.Panic("err nsq producer init", zap.String("nsqd", nsqd), zap.Error(err))
		}

		logger.Info("nsq producer is OK")
	}
}

// WithNSQConsumers 设置nsq消费者
func WithNSQConsumers(lookupd []string, consumers ...NSQConsumer) InitOption {
	return func() {
		if err := setNSQConsumers(lookupd, consumers...); err != nil {
			logger.Panic("err set nsq consumers", zap.Strings("lookupd", lookupd), zap.Error(err))
		}

		logger.Info("nsq consumers set OK")
	}
}

// WithWebsocket 设置websocket
func WithWebsocket(upgrader *websocket.Upgrader) InitOption {
	return func() {
		wsupgrader = upgrader
	}
}

// Init yiigo初始化
func Init(options ...InitOption) {
	for _, f := range options {
		go f()
	}

	logger.Info("yiigo init complete!")
}
