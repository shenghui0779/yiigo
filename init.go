package yiigo

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/gorilla/websocket"
	"github.com/nsqio/go-nsq"
	"github.com/redis/go-redis/v9"
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

// WithRedis 注册Redis
func WithRedis(name string, cfg *redis.UniversalOptions) InitOption {
	return func() {
		if len(name) == 0 || cfg == nil {
			return
		}

		if err := initRedis(name, cfg); err != nil {
			logger.Panic(fmt.Sprintf("err redis.%s init", name), zap.Strings("addr", cfg.Addrs), zap.Error(err))
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

// WithNSQ 设置NSQ
func WithNSQ(nsqd string, lookupd []string, cfg *nsq.Config, consumers ...NSQConsumer) InitOption {
	return func() {
		if err := initNSQ(nsqd, lookupd, cfg, consumers...); err != nil {
			logger.Panic("err nsq init", zap.String("nsqd", nsqd), zap.Strings("lookupd", lookupd), zap.Error(err))
		}

		logger.Info("nsq producer is OK")
	}
}

// WithWebsocket 设置websocket
func WithWebsocket(up *websocket.Upgrader) InitOption {
	return func() {
		upgrader = up
	}
}

// Init yiigo初始化
func Init(options ...InitOption) {
	for _, f := range options {
		f()
	}

	logger.Info("yiigo init complete!")
}
