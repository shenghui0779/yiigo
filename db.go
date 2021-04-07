package yiigo

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"time"

	entsql "entgo.io/ent/dialect/sql"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
	"go.uber.org/zap"
)

type DBDriver string

const (
	MySQL    DBDriver = "mysql"
	Postgres DBDriver = "postgres"
	SQLite   DBDriver = "sqlite3"
)

var (
	defaultDB *sqlx.DB
	dbmap     sync.Map

	defaultEntDriver *entsql.Driver
	entmap           sync.Map
)

type dbConfig struct {
	Driver          string `toml:"driver"`
	DSN             string `toml:"dsn"`
	MaxOpenConns    int    `toml:"max_open_conns"`
	MaxIdleConns    int    `toml:"max_idle_conns"`
	ConnMaxIdleTime int    `toml:"conn_max_idle_time"`
	ConnMaxLifetime int    `toml:"conn_max_lifetime"`
}

func dbDial(cfg *dbConfig) (*sql.DB, error) {
	if !InStrings(cfg.Driver, []string{string(MySQL), string(Postgres), string(SQLite)}) {
		return nil, fmt.Errorf("yiigo: unknown db driver %s, expects mysql, postgres, sqlite3", cfg.Driver)
	}

	db, err := sql.Open(cfg.Driver, cfg.DSN)

	if err != nil {
		return nil, err
	}

	if err = db.Ping(); err != nil {
		db.Close()

		return nil, err
	}

	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxIdleTime(time.Duration(cfg.ConnMaxIdleTime) * time.Second)
	db.SetConnMaxLifetime(time.Duration(cfg.ConnMaxLifetime) * time.Second)

	return db, nil
}

func initDB() {
	configs := make(map[string]*dbConfig)

	if err := env.Get("db").Unmarshal(&configs); err != nil {
		logger.Panic("yiigo: db init error", zap.Error(err))
	}

	if len(configs) == 0 {
		return
	}

	for name, cfg := range configs {
		db, err := dbDial(cfg)

		if err != nil {
			logger.Panic("yiigo: db init error", zap.String("name", name), zap.Error(err))
		}

		sqlxDB := sqlx.NewDb(db, cfg.Driver)
		entDriver := entsql.OpenDB(cfg.Driver, db)

		if name == defaultConn {
			defaultDB = sqlxDB
			defaultEntDriver = entDriver
		}

		dbmap.Store(name, sqlxDB)
		entmap.Store(name, entDriver)

		logger.Info(fmt.Sprintf("yiigo: db.%s is OK.", name))
	}
}

// DB returns a db.
func DB(name ...string) *sqlx.DB {
	if len(name) == 0 {
		if defaultDB == nil {
			logger.Panic(fmt.Sprintf("yiigo: unknown db.%s (forgotten configure?)", defaultConn))
		}

		return defaultDB
	}

	v, ok := dbmap.Load(name[0])

	if !ok {
		logger.Panic(fmt.Sprintf("yiigo: unknown db.%s (forgotten configure?)", name[0]))
	}

	return v.(*sqlx.DB)
}

// EntDriver returns an ent dialect.Driver.
func EntDriver(name ...string) *entsql.Driver {
	if len(name) == 0 || name[0] == defaultConn {
		if defaultEntDriver == nil {
			logger.Panic(fmt.Sprintf("yiigo: unknown db.%s (forgotten configure?)", defaultConn))
		}

		return defaultEntDriver
	}

	v, ok := entmap.Load(name[0])

	if !ok {
		logger.Panic(fmt.Sprintf("yiigo: unknown db.%s (forgotten configure?)", name[0]))
	}

	return v.(*entsql.Driver)
}

// DBTransaction Executes db transaction with callback function.
func DBTransaction(db *sqlx.DB, process func(tx *sqlx.Tx) error) error {
	tx, err := db.Beginx()

	if err != nil {
		return err
	}

	defer txRecover(tx)

	if err = process(tx); err != nil {
		txRollback(tx)

		return err
	}

	if err = tx.Commit(); err != nil {
		txRollback(tx)

		return err
	}

	return nil
}

// DBTransactionX Executes db transaction with callback function.
// The provided context is used until the transaction is committed or rolledback.
func DBTransactionX(ctx context.Context, db *sqlx.DB, process func(tx *sqlx.Tx) error) error {
	tx, err := db.BeginTxx(ctx, nil)

	if err != nil {
		return err
	}

	defer txRecover(tx)

	if err = process(tx); err != nil {
		txRollback(tx)

		return err
	}

	if err = tx.Commit(); err != nil {
		txRollback(tx)

		return err
	}

	return nil
}

func txRecover(tx *sqlx.Tx) {
	if r := recover(); r != nil {
		logger.Fatal("yiigo: db transaction process panic", zap.Any("error", r))

		txRollback(tx)
	}
}

func txRollback(tx *sqlx.Tx) {
	if err := tx.Rollback(); err != nil {
		logger.Error("yiigo: db transaction rollback error", zap.Error(err))
	}
}
