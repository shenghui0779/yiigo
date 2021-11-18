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

type DBOptions struct {
	// MaxOpenConns is the maximum number of open connections to the database.
	// Use value -1 for no timeout and 0 for default.
	// Default is 20.
	MaxOpenConns int

	// MaxIdleConns is the maximum number of connections in the idle connection pool.
	// Use value -1 for no timeout and 0 for default.
	// Default is 10.
	MaxIdleConns int

	// ConnMaxLifetime is the maximum amount of time a connection may be reused.
	// Use value -1 for no timeout and 0 for default.
	// Default is 60 seconds.
	ConnMaxLifetime time.Duration

	// ConnMaxIdleTime is the maximum amount of time a connection may be idle.
	// Use value -1 for no timeout and 0 for default.
	// Default is 5 minutes.
	ConnMaxIdleTime time.Duration
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

func initDB(name string, driver DBDriver, dsn string, opt *DBOptions) {
	db, err := sql.Open(string(driver), dsn)

	if err != nil {
		logger.Panic("[yiigo] db init error", zap.String("name", name), zap.Error(err))
	}

	if err = db.Ping(); err != nil {
		db.Close()

		logger.Panic("[yiigo] db init error", zap.String("name", name), zap.Error(err))
	}

	options := &DBOptions{
		MaxOpenConns:    20,
		MaxIdleConns:    10,
		ConnMaxLifetime: 60 * time.Minute,
		ConnMaxIdleTime: 5 * time.Minute,
	}

	if opt != nil {
		options.rebuild(opt)
	}

	db.SetMaxOpenConns(options.MaxOpenConns)
	db.SetMaxIdleConns(options.MaxIdleConns)
	db.SetConnMaxLifetime(options.ConnMaxLifetime)
	db.SetConnMaxIdleTime(options.ConnMaxIdleTime)

	sqlxDB := sqlx.NewDb(db, string(driver))
	entDriver := entsql.OpenDB(string(driver), db)

	if name == Default {
		defaultDB = sqlxDB
		defaultEntDriver = entDriver
	}

	dbmap.Store(name, sqlxDB)
	entmap.Store(name, entDriver)

	logger.Info(fmt.Sprintf("[yiigo] db.%s is OK", name))
}

// DB returns a db.
func DB(name ...string) *sqlx.DB {
	if len(name) == 0 || name[0] == Default {
		if defaultDB == nil {
			logger.Panic(fmt.Sprintf("[yiigo] unknown db.%s (forgotten configure?)", Default))
		}

		return defaultDB
	}

	v, ok := dbmap.Load(name[0])

	if !ok {
		logger.Panic(fmt.Sprintf("[yiigo] unknown db.%s (forgotten configure?)", name[0]))
	}

	return v.(*sqlx.DB)
}

// EntDriver returns an ent dialect.Driver.
func EntDriver(name ...string) *entsql.Driver {
	if len(name) == 0 || name[0] == Default {
		if defaultEntDriver == nil {
			logger.Panic(fmt.Sprintf("[yiigo] unknown db.%s (forgotten configure?)", Default))
		}

		return defaultEntDriver
	}

	v, ok := entmap.Load(name[0])

	if !ok {
		logger.Panic(fmt.Sprintf("[yiigo] unknown db.%s (forgotten configure?)", name[0]))
	}

	return v.(*entsql.Driver)
}

// DBTxHandler db tx callback func.
type DBTxHandler func(ctx context.Context, tx *sqlx.Tx) error

// DBTransaction Executes db transaction with callback function.
// The provided context is used until the transaction is committed or rolledback.
func DBTransaction(ctx context.Context, db *sqlx.DB, callback DBTxHandler) error {
	tx, err := db.BeginTxx(ctx, nil)

	if err != nil {
		return err
	}

	defer func() {
		if r := recover(); r != nil {
			logger.Error("[yiigo] db transaction handler panic", zap.Any("error", r), zap.ByteString("stack", debug.Stack()))

			dbTxRollback(tx)
		}
	}()

	if err = callback(ctx, tx); err != nil {
		dbTxRollback(tx)

		return err
	}

	if err = tx.Commit(); err != nil {
		dbTxRollback(tx)

		return err
	}

	return nil
}

func dbTxRollback(tx *sqlx.Tx) {
	if err := tx.Rollback(); err != nil && err != sql.ErrTxDone {
		logger.Error("[yiigo] db transaction rollback error", zap.Error(err))
	}
}
