package yiigo

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"sync"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
	"github.com/pelletier/go-toml"
	"go.uber.org/zap"
)

type DBDriver string

const (
	MySQL    DBDriver = "mysql"
	Postgres DBDriver = "postgres"
	SQLite   DBDriver = "sqlite3"
)

var (
	defaultDB  *sqlx.DB
	dbmap      sync.Map
	defaultOrm *gorm.DB
	ormap      sync.Map
)

type dbConfig struct {
	Driver          string `toml:"driver"`
	Dsn             string `toml:"dsn"`
	MaxOpenConns    int    `toml:"max_open_conns"`
	MaxIdleConns    int    `toml:"max_idle_conns"`
	ConnMaxLifetime int    `toml:"conn_max_lifetime"`
}

func dbDial(cfg *dbConfig, debug bool) (*gorm.DB, error) {
	if !InStrings(cfg.Driver, MySQL, Postgres, SQLite) {
		return nil, fmt.Errorf("yiigo: unknown db driver %s, expects mysql, postgres, sqlite3", cfg.Driver)
	}

	orm, err := gorm.Open(cfg.Driver, cfg.Dsn)

	if err != nil {
		return nil, err
	}

	if debug {
		orm.LogMode(true)
	}

	orm.DB().SetMaxOpenConns(cfg.MaxOpenConns)
	orm.DB().SetMaxIdleConns(cfg.MaxIdleConns)
	orm.DB().SetConnMaxLifetime(time.Duration(cfg.ConnMaxLifetime) * time.Second)

	return orm, nil
}

func initDB(debug bool) {
	tree, ok := env.get("db").(*toml.Tree)

	if !ok {
		return
	}

	keys := tree.Keys()

	if len(keys) == 0 {
		return
	}

	for _, v := range keys {
		node, ok := tree.Get(v).(*toml.Tree)

		if !ok {
			continue
		}

		cfg := new(dbConfig)

		if err := node.Unmarshal(cfg); err != nil {
			logger.Panic("yiigo: db init error", zap.String("name", v), zap.Error(err))
		}

		orm, err := dbDial(cfg, debug)

		if err != nil {
			logger.Panic("yiigo: db init error", zap.String("name", v), zap.Error(err))
		}

		db := sqlx.NewDb(orm.DB(), cfg.Driver)

		if v == AsDefault {
			defaultDB = db
			defaultOrm = orm
		}

		dbmap.Store(v, db)
		ormap.Store(v, orm)

		logger.Info(fmt.Sprintf("yiigo: db.%s is OK.", v))
	}
}

// DB returns a db.
func DB(name ...string) *sqlx.DB {
	if len(name) == 0 {
		if defaultDB == nil {
			logger.Panic(fmt.Sprintf("yiigo: unknown db.%s (forgotten configure?)", AsDefault))
		}

		return defaultDB
	}

	v, ok := dbmap.Load(name[0])

	if !ok {
		logger.Panic(fmt.Sprintf("yiigo: unknown db.%s (forgotten configure?)", name[0]))
	}

	return v.(*sqlx.DB)
}

// Orm returns an orm's db.
func Orm(name ...string) *gorm.DB {
	if len(name) == 0 || name[0] == AsDefault {
		if defaultOrm == nil {
			logger.Panic(fmt.Sprintf("yiigo: unknown db.%s (forgotten configure?)", AsDefault))
		}

		return defaultOrm
	}

	v, ok := ormap.Load(name[0])

	if !ok {
		logger.Panic(fmt.Sprintf("yiigo: unknown db.%s (forgotten configure?)", name[0]))
	}

	return v.(*gorm.DB)
}

var (
	errInsertInvalidType = errors.New("yiigo: invalid data type of InsertSQL() / PGInsertSQL(), expects: struct, *struct, []struct, []*struct, yiigo.X, []yiigo.X")
	errUpdateInvalidType = errors.New("yiigo: invalid data type of UpdateSQL() / PGUpdateSQL(), expects: struct, *struct, yiigo.X")
)

// InsertSQL returns mysql insert sql and binds.
// param data expects: `struct`, `*struct`, `[]struct`, `[]*struct`, `yiigo.X`, `[]yiigo.X`.
func InsertSQL(table string, data interface{}) (string, []interface{}) {
	columns, placeholders, binds := insertSQLBuilder(MySQL, data)
	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)", table, strings.Join(columns, ", "), strings.Join(placeholders, ", "))

	return sqlx.Rebind(sqlx.BindType(MySQL), query), binds
}

// UpdateSQL returns mysql update sql and binds.
// param query expects eg: "UPDATE `table` SET ? WHERE `id` = ?".
// param data expects: `struct`, `*struct`, `yiigo.X`.
func UpdateSQL(query string, data interface{}, condition ...interface{}) (string, []interface{}) {
	return updateSQLBuilder(query, data, condition...)
}

// PGInsertSQL returns postgres insert sql and binds.
// param data expects: `struct`, `*struct`, `[]struct`, `[]*struct`, `yiigo.X`, `[]yiigo.X`.
func PGInsertSQL(table string, data interface{}) (string, []interface{}) {
	columns, placeholders, binds := insertSQLBuilder(MySQL, data)
	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s) RETURNING id", table, strings.Join(columns, ", "), strings.Join(placeholders, ", "))

	return sqlx.Rebind(sqlx.BindType(Postgres), query), binds
}

// PGUpdateSQL returns postgres update sql and binds.
// param query expects eg: "UPDATE `table` SET ? WHERE `id` = ?".
// param data expects: `struct`, `*struct`, `yiigo.X`.
func PGUpdateSQL(query string, data interface{}, condition ...interface{}) (string, []interface{}) {
	return updateSQLBuilder(query, data, condition...)
}

// LiteInsertSQL returns sqlite3 insert sql and binds.
// param data expects: `struct`, `*struct`, `yiigo.X`.
func LiteInsertSQL(table string, data interface{}) (string, []interface{}) {
	columns, placeholders, binds := insertSQLBuilder(MySQL, data)
	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)", table, strings.Join(columns, ", "), strings.Join(placeholders, ", "))

	return sqlx.Rebind(sqlx.BindType(SQLite), query), binds
}

// LiteUpdateSQL returns sqlite3 update sql and binds.
// param query expects eg: "UPDATE `table` SET ? WHERE `id` = ?".
// param data expects: `struct`, `*struct`, `yiigo.X`.
func LiteUpdateSQL(query string, data interface{}, condition ...interface{}) (string, []interface{}) {
	return updateSQLBuilder(query, data, condition...)
}

type SQLBuilder struct {
	driver DBDriver
	table  string
	data   interface{}
}

func insertSQLBuilder(driver string, data interface{}) (columns []string, placeholders []string, binds []interface{}) {
	v := reflect.Indirect(reflect.ValueOf(data))

	switch v.Kind() {
	case reflect.Map:
		if x, ok := data.(X); ok {
			columns, placeholders, binds = singleInsertWithMap(x)
		}
	case reflect.Struct:
		columns, placeholders, binds = singleInsertWithStruct(v)
	case reflect.Slice:
		if driver == SQLite {
			panic(errInsertInvalidType)
		}

		if v.Len() == 0 {
			return
		}

		e := v.Type().Elem()

		switch e.Kind() {
		case reflect.Map:
			x, ok := data.([]X)

			if !ok {
				panic(errInsertInvalidType)
			}

			columns, placeholders, binds = batchInsertWithMap(x)
		case reflect.Struct:
			columns, placeholders, binds = batchInsertWithStruct(v)
		case reflect.Ptr:
			if e.Elem().Kind() != reflect.Struct {
				panic(errInsertInvalidType)
			}

			columns, placeholders, binds = batchInsertWithStruct(v)
		default:
			panic(errInsertInvalidType)
		}
	default:
		panic(errInsertInvalidType)
	}

	return
}

func updateSQLBuilder(query string, data interface{}, condition ...interface{}) (string, []interface{}) {
	var (
		sql   string
		binds []interface{}
	)

	v := reflect.Indirect(reflect.ValueOf(data))

	switch v.Kind() {
	case reflect.Map:
		x, ok := data.(X)

		if !ok {
			panic(errUpdateInvalidType)
		}

		sql, binds = updateWithMap(query, x, condition...)
	case reflect.Struct:
		sql, binds = updateWithStruct(query, v, condition...)
	default:
		panic(errUpdateInvalidType)
	}

	return sql, binds
}

func singleInsertWithMap(data X) ([]string, []string, []interface{}) {
	fieldNum := len(data)

	columns := make([]string, 0, fieldNum)
	placeholders := make([]string, 0, fieldNum)
	binds := make([]interface{}, 0, fieldNum)

	for k, v := range data {
		columns = append(columns, k)
		placeholders = append(placeholders, "?")
		binds = append(binds, v)
	}

	return columns, placeholders, binds
}

func singleInsertWithStruct(v reflect.Value) ([]string, []string, []interface{}) {
	fieldNum := v.NumField()

	columns := make([]string, 0, fieldNum)
	placeholders := make([]string, 0, fieldNum)
	binds := make([]interface{}, 0, fieldNum)

	t := v.Type()

	for i := 0; i < fieldNum; i++ {
		column := t.Field(i).Tag.Get("db")

		if column == "-" {
			continue
		}

		if column == "" {
			column = t.Field(i).Name
		}

		columns = append(columns, column)
		placeholders = append(placeholders, "?")
		binds = append(binds, v.Field(i).Interface())
	}

	return columns, placeholders, binds
}

func batchInsertWithMap(data []X) ([]string, []string, []interface{}) {
	fieldNum := len(data[0])

	fields := make([]string, 0, fieldNum)
	columns := make([]string, 0, fieldNum)
	placeholders := make([]string, 0, fieldNum)
	binds := make([]interface{}, 0, fieldNum*len(data))

	for k := range data[0] {
		fields = append(fields, k)

		columns = append(columns, k)
	}

	for _, x := range data {
		phrs := make([]string, 0, fieldNum)

		for _, v := range fields {
			phrs = append(phrs, "?")
			binds = append(binds, x[v])
		}

		placeholders = append(placeholders, fmt.Sprintf("(%s)", strings.Join(phrs, ", ")))
	}

	return columns, placeholders, binds
}

func batchInsertWithStruct(v reflect.Value) ([]string, []string, []interface{}) {
	first := reflect.Indirect(v.Index(0))

	fieldNum := first.NumField()

	columns := make([]string, 0, fieldNum)
	placeholders := make([]string, 0, fieldNum)
	binds := make([]interface{}, 0, fieldNum*v.Len())

	t := first.Type()

	for i := 0; i < count; i++ {
		phrs := make([]string, 0, fieldNum)

		for j := 0; j < fieldNum; j++ {
			column := t.Field(j).Tag.Get("db")

			if column == "-" {
				continue
			}

			if i == 0 {
				if column == "" {
					column = t.Field(j).Name
				}

				columns = append(columns, column)
			}

			phrs = append(phrs, "?")
			binds = append(binds, reflect.Indirect(v.Index(i)).Field(j).Interface())
		}

		placeholders = append(placeholders, fmt.Sprintf("(%s)", strings.Join(phrs, ", ")))
	}

	return columns, placeholders, binds
}

func updateWithMap(query string, data X, conditions ...interface{}) (string, []interface{}) {
	dataLen := len(data)

	sets := make([]string, 0, dataLen)
	binds := make([]interface{}, 0, dataLen+len(conditions))

	for k, v := range data {
		sets = append(sets, fmt.Sprintf("%s = ?", k))
		binds = append(binds, v)
	}

	query = strings.Replace(query, "?", strings.Join(sets, ", "), 1)
	binds = append(binds, conditions...)

	return query, binds
}

func updateWithStruct(query string, v reflect.Value, conditions ...interface{}) (string, []interface{}) {
	fieldNum := v.NumField()

	sets := make([]string, 0, fieldNum)
	binds := make([]interface{}, 0, fieldNum+len(conditions))

	t := v.Type()

	for i := 0; i < fieldNum; i++ {
		column := t.Field(i).Tag.Get("db")

		if column == "-" {
			continue
		}

		if column == "" {
			column = t.Field(i).Name
		}

		sets = append(sets, fmt.Sprintf("%s = ?", column))
		binds = append(binds, v.Field(i).Interface())
	}

	query = strings.Replace(query, "?", strings.Join(sets, ", "), 1)
	binds = append(binds, conditions...)

	return query, binds
}
