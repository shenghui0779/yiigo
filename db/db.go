package db

import (
	"database/sql"
	"time"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/jackc/pgx/v5/stdlib"
	_ "github.com/mattn/go-sqlite3"
)

// Config 数据库初始化配置
type Config struct {
	// Driver 驱动名称
	Driver string
	// DSN 数据源名称
	// [-- MySQL] username:password@tcp(localhost:3306)/dbname?timeout=10s&charset=utf8mb4&collation=utf8mb4_general_ci&parseTime=True&loc=Local
	// [Postgres] host=localhost port=5432 user=root password=secret dbname=test connect_timeout=10 sslmode=disable
	// [- SQLite] file::memory:?cache=shared
	DSN string
	// Options 配置选项
	Options *Options
}

// Options 数据库配置选项
type Options struct {
	// MaxOpenConns 设置最大可打开的连接数
	MaxOpenConns int
	// MaxIdleConns 连接池最大闲置连接数
	MaxIdleConns int
	// ConnMaxLifetime 连接的最大生命时长
	ConnMaxLifetime time.Duration
	// ConnMaxIdleTime 连接最大闲置时间
	ConnMaxIdleTime time.Duration
}

func New(cfg *Config) (*sql.DB, error) {
	db, err := sql.Open(cfg.Driver, cfg.DSN)
	if err != nil {
		return nil, err
	}
	if err = db.Ping(); err != nil {
		db.Close()
		return nil, err
	}

	if cfg.Options != nil {
		db.SetMaxOpenConns(cfg.Options.MaxOpenConns)
		db.SetMaxIdleConns(cfg.Options.MaxIdleConns)
		db.SetConnMaxLifetime(cfg.Options.ConnMaxLifetime)
		db.SetConnMaxIdleTime(cfg.Options.ConnMaxIdleTime)
	}

	return db, nil
}
