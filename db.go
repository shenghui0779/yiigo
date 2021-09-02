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
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
	"go.uber.org/zap"
)

// DBDriver db driver
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

type dbSettings struct {
	maxOpenConns    int
	maxIdleConns    int
	connMaxIdleTime time.Duration
	connMaxLifetime time.Duration
}

// DBOption configures how we set up the db.
type DBOption func(s *dbSettings)

// WithDBMaxOpenConns specifies the `MaxOpenConns` for db.
func WithDBMaxOpenConns(n int) DBOption {
	return func(s *dbSettings) {
		s.maxOpenConns = n
	}
}

// WithDBMaxIdleConns specifies the `MaxIdleConns` for db.
func WithDBMaxIdleConns(n int) DBOption {
	return func(s *dbSettings) {
		s.maxIdleConns = n
	}
}

// WithDBConnMaxIdleTime specifies the `ConnMaxIdleTime` for db.
func WithDBConnMaxIdleTime(t time.Duration) DBOption {
	return func(s *dbSettings) {
		s.connMaxIdleTime = t
	}
}

// WithDBConnMaxLifetime specifies the `ConnMaxLifetime` for db.
func WithDBConnMaxLifetime(t time.Duration) DBOption {
	return func(s *dbSettings) {
		s.connMaxLifetime = t
	}
}

func initDB(name string, driver DBDriver, dsn string, options ...DBOption) {
	db, err := sql.Open(string(driver), dsn)

	if err != nil {
		logger.Panic("yiigo: db init error", zap.String("name", name), zap.Error(err))
	}

	if err = db.Ping(); err != nil {
		db.Close()

		logger.Panic("yiigo: db init error", zap.String("name", name), zap.Error(err))
	}

	settings := &dbSettings{
		maxOpenConns:    20,
		maxIdleConns:    10,
		connMaxIdleTime: 60,
		connMaxLifetime: 600,
	}

	for _, f := range options {
		f(settings)
	}

	db.SetMaxOpenConns(settings.maxOpenConns)
	db.SetMaxIdleConns(settings.maxIdleConns)
	db.SetConnMaxIdleTime(settings.connMaxIdleTime)
	db.SetConnMaxLifetime(settings.connMaxLifetime)

	sqlxDB := sqlx.NewDb(db, string(driver))
	entDriver := entsql.OpenDB(string(driver), db)

	if name == Default {
		defaultDB = sqlxDB
		defaultEntDriver = entDriver
	}

	dbmap.Store(name, sqlxDB)
	entmap.Store(name, entDriver)

	logger.Info(fmt.Sprintf("yiigo: db.%s is OK.", name))
}

// DB returns a db.
func DB(name ...string) *sqlx.DB {
	if len(name) == 0 || name[0] == Default {
		if defaultDB == nil {
			logger.Panic(fmt.Sprintf("yiigo: unknown db.%s (forgotten configure?)", Default))
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
	if len(name) == 0 || name[0] == Default {
		if defaultEntDriver == nil {
			logger.Panic(fmt.Sprintf("yiigo: unknown db.%s (forgotten configure?)", Default))
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
// The provided context is used until the transaction is committed or rolledback.
func DBTransaction(ctx context.Context, db *sqlx.DB, process func(ctx context.Context, tx *sqlx.Tx) error) error {
	tx, err := db.BeginTxx(ctx, nil)

	if err != nil {
		return err
	}

	defer func() {
		if r := recover(); r != nil {
			logger.Fatal("yiigo: db transaction process panic", zap.Any("error", r), zap.ByteString("stack", debug.Stack()))

			txRollback(tx)
		}
	}()

	if err = process(ctx, tx); err != nil {
		txRollback(tx)

		return err
	}

	if err = tx.Commit(); err != nil {
		txRollback(tx)

		return err
	}

	return nil
}

func txRollback(tx *sqlx.Tx) {
	if err := tx.Rollback(); err != nil && err != sql.ErrTxDone {
		logger.Error("yiigo: db transaction rollback error", zap.Error(err))
	}
}
