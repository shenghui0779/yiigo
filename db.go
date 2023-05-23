package yiigo

import (
	"context"
	"database/sql"
	"fmt"
	"runtime/debug"
	"sync"
	"time"

	entsql "entgo.io/ent/dialect/sql"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"go.uber.org/zap"
)

// DBDriver 数据库驱动
type DBDriver string

const (
	MySQL    DBDriver = "mysql"
	Postgres DBDriver = "pgx"
	SQLite   DBDriver = "sqlite3"
)

var (
	defaultDB *sqlx.DB
	dbmap     sync.Map

	defaultEntDriver *entsql.Driver
	entmap           sync.Map
)

// DBConfig 数据库初始化配置
type DBConfig struct {
	// DSN 数据源名称
	// [-- MySQL] username:password@tcp(localhost:3306)/dbname?timeout=10s&charset=utf8mb4&collation=utf8mb4_general_ci&parseTime=True&loc=Local
	// [Postgres] host=localhost port=5432 user=root password=secret dbname=test connect_timeout=10 sslmode=disable
	// [- SQLite] file::memory:?cache=shared
	DSN string `json:"dsn"`

	// Options 配置选项
	Options *DBOptions `json:"options"`
}

// DBOptions 数据库配置选项
type DBOptions struct {
	// MaxOpenConns 设置最大可打开的连接数
	// -1：不限；默认：20
	MaxOpenConns int `json:"max_open_conns"`

	// MaxIdleConns 连接池最大闲置连接数
	// -1：不限；默认：10
	MaxIdleConns int `json:"max_idle_conns"`

	// ConnMaxLifetime 连接的最大生命时长
	// -1：不限；默认：10分钟
	ConnMaxLifetime time.Duration `json:"conn_max_lifetime"`

	// ConnMaxIdleTime 连接最大闲置时间
	// -1：不限；默认：5分钟
	ConnMaxIdleTime time.Duration `json:"conn_max_idle_time"`
}

func (o *DBOptions) rebuild(opt *DBOptions) {
	if opt.MaxOpenConns > 0 {
		o.MaxOpenConns = opt.MaxOpenConns
	} else {
		if opt.MaxOpenConns == -1 {
			o.MaxOpenConns = 0
		}
	}

	if opt.MaxIdleConns > 0 {
		o.MaxIdleConns = opt.MaxIdleConns
	} else {
		if opt.MaxIdleConns == -1 {
			o.MaxIdleConns = 0
		}
	}

	if opt.ConnMaxLifetime > 0 {
		o.ConnMaxLifetime = opt.ConnMaxLifetime
	} else {
		if opt.ConnMaxLifetime == -1 {
			o.ConnMaxLifetime = 0
		}
	}

	if opt.ConnMaxIdleTime > 0 {
		o.ConnMaxIdleTime = opt.ConnMaxIdleTime
	} else {
		if opt.ConnMaxIdleTime == -1 {
			o.ConnMaxIdleTime = 0
		}
	}
}

func initDB(name string, driver DBDriver, cfg *DBConfig) {
	db, err := sql.Open(string(driver), cfg.DSN)

	if err != nil {
		logger.Panic(fmt.Sprintf("err db.%s open", name), zap.String("dsn", cfg.DSN), zap.Error(err))
	}

	if err = db.Ping(); err != nil {
		db.Close()

		logger.Panic(fmt.Sprintf("err db.%s ping", name), zap.String("dsn", cfg.DSN), zap.Error(err))
	}

	opt := &DBOptions{
		MaxOpenConns:    20,
		MaxIdleConns:    10,
		ConnMaxLifetime: 10 * time.Minute,
		ConnMaxIdleTime: 5 * time.Minute,
	}

	if cfg.Options != nil {
		opt.rebuild(cfg.Options)
	}

	db.SetMaxOpenConns(opt.MaxOpenConns)
	db.SetMaxIdleConns(opt.MaxIdleConns)
	db.SetConnMaxLifetime(opt.ConnMaxLifetime)
	db.SetConnMaxIdleTime(opt.ConnMaxIdleTime)

	sqlxDB := sqlx.NewDb(db, string(driver))
	entDriver := entsql.OpenDB(string(driver), db)

	if name == Default {
		defaultDB = sqlxDB
		defaultEntDriver = entDriver
	}

	dbmap.Store(name, sqlxDB)
	entmap.Store(name, entDriver)

	logger.Info(fmt.Sprintf("db.%s is OK", name))
}

// DB 返回一个sqlx数据库实例
func DB(name ...string) *sqlx.DB {
	if len(name) == 0 || name[0] == Default {
		if defaultDB == nil {
			logger.Panic(fmt.Sprintf("unknown db.%s (forgotten configure?)", Default))
		}

		return defaultDB
	}

	v, ok := dbmap.Load(name[0])

	if !ok {
		logger.Panic(fmt.Sprintf("unknown db.%s (forgotten configure?)", name[0]))
	}

	return v.(*sqlx.DB)
}

// EntDriver 返回一个ent驱动实例
func EntDriver(name ...string) *entsql.Driver {
	if len(name) == 0 || name[0] == Default {
		if defaultEntDriver == nil {
			logger.Panic(fmt.Sprintf("unknown db.%s (forgotten configure?)", Default))
		}

		return defaultEntDriver
	}

	v, ok := entmap.Load(name[0])

	if !ok {
		logger.Panic(fmt.Sprintf("unknown db.%s (forgotten configure?)", name[0]))
	}

	return v.(*entsql.Driver)
}

// DBTransaction 执行数据库事物
func DBTransaction(ctx context.Context, db *sqlx.DB, f func(ctx context.Context, tx *sqlx.Tx) error) error {
	tx, err := db.BeginTxx(ctx, nil)

	if err != nil {
		return err
	}

	defer func() {
		if r := recover(); r != nil {
			logger.Error("db tx panic", zap.Any("error", r), zap.ByteString("stack", debug.Stack()))

			rollback(tx)
		}
	}()

	if err = f(ctx, tx); err != nil {
		rollback(tx)

		return err
	}

	if err = tx.Commit(); err != nil {
		rollback(tx)

		return err
	}

	return nil
}

func rollback(tx *sqlx.Tx) {
	if err := tx.Rollback(); err != nil && err != sql.ErrTxDone {
		logger.Error("err db tx rollback", zap.Error(err))
	}
}
