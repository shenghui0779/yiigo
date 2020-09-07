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

const (
	MySQL    = "mysql"
	Postgres = "postgres"
	SQLite   = "sqlite3"
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
	return insertSQLBuilder(MySQL, table, data)
}

// UpdateSQL returns mysql update sql and binds.
// param query expects eg: "UPDATE `table` SET ? WHERE `id` = ?".
// param data expects: `struct`, `*struct`, `yiigo.X`.
func UpdateSQL(query string, data interface{}, condition ...interface{}) (string, []interface{}) {
	return updateSQLBuilder(MySQL, query, data, condition...)
}

// PGInsertSQL returns postgres insert sql and binds.
// param data expects: `struct`, `*struct`, `[]struct`, `[]*struct`, `yiigo.X`, `[]yiigo.X`.
func PGInsertSQL(table string, data interface{}) (string, []interface{}) {
	return insertSQLBuilder(Postgres, table, data)
}

// PGUpdateSQL returns postgres update sql and binds.
// param query expects eg: "UPDATE `table` SET $1 WHERE `id` = $2".
// param data expects: `struct`, `*struct`, `yiigo.X`.
func PGUpdateSQL(query string, data interface{}, condition ...interface{}) (string, []interface{}) {
	return updateSQLBuilder(Postgres, query, data, condition...)
}

// LiteInsertSQL returns sqlite3 insert sql and binds.
// param data expects: `struct`, `*struct`, `yiigo.X`.
func LiteInsertSQL(table string, data interface{}) (string, []interface{}) {
	return insertSQLBuilder(Postgres, table, data)
}

// LiteUpdateSQL returns sqlite3 update sql and binds.
// param query expects eg: "UPDATE `table` SET ? WHERE `id` = ?".
// param data expects: `struct`, `*struct`, `yiigo.X`.
func LiteUpdateSQL(query string, data interface{}, condition ...interface{}) (string, []interface{}) {
	return updateSQLBuilder(Postgres, query, data, condition...)
}

func insertSQLBuilder(driver, table string, data interface{}) (string, []interface{}) {
	var (
		sql   string
		binds []interface{}
	)

	v := reflect.Indirect(reflect.ValueOf(data))

	switch v.Kind() {
	case reflect.Map:
		if x, ok := data.(X); ok {
			sql, binds = singleInsertWithMap(driver, table, x)
		}
	case reflect.Struct:
		sql, binds = singleInsertWithStruct(driver, table, v)
	case reflect.Slice:
		if driver == SQLite {
			panic(errInsertInvalidType)
		}

		count := v.Len()

		if count == 0 {
			return sql, binds
		}

		e := v.Type().Elem()

		switch e.Kind() {
		case reflect.Map:
			x, ok := data.([]X)

			if !ok {
				panic(errInsertInvalidType)
			}

			sql, binds = batchInsertWithMap(driver, table, x, count)
		case reflect.Struct:
			sql, binds = batchInsertWithStruct(driver, table, v, count)
		case reflect.Ptr:
			if e.Elem().Kind() != reflect.Struct {
				panic(errInsertInvalidType)
			}

			sql, binds = batchInsertWithStruct(driver, table, v, count)
		default:
			panic(errInsertInvalidType)
		}
	default:
		panic(errInsertInvalidType)
	}

	return sql, binds
}

func updateSQLBuilder(driver, query string, data interface{}, condition ...interface{}) (string, []interface{}) {
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

		sql, binds = updateWithMap(driver, query, x, condition...)
	case reflect.Struct:
		sql, binds = updateWithStruct(driver, query, v, condition...)
	default:
		panic(errUpdateInvalidType)
	}

	return sql, binds
}

func singleInsertWithMap(driver, table string, data X) (string, []interface{}) {
	fieldNum := len(data)

	columns := make([]string, 0, fieldNum)
	placeholders := make([]string, 0, fieldNum)

	sql := ""
	binds := make([]interface{}, 0, fieldNum)

	switch driver {
	case MySQL:
		for k, v := range data {
			columns = append(columns, fmt.Sprintf("`%s`", k))
			placeholders = append(placeholders, "?")
			binds = append(binds, v)
		}

		sql = fmt.Sprintf("INSERT INTO `%s` (%s) VALUES (%s)", table, strings.Join(columns, ", "), strings.Join(placeholders, ", "))
	case Postgres:
		bindIndex := 0

		for k, v := range data {
			bindIndex++

			columns = append(columns, fmt.Sprintf(`"%s"`, k))
			placeholders = append(placeholders, fmt.Sprintf("$%d", bindIndex))
			binds = append(binds, v)
		}

		sql = fmt.Sprintf(`INSERT INTO "%s" (%s) VALUES (%s) RETURNING "id"`, table, strings.Join(columns, ", "), strings.Join(placeholders, ", "))
	case SQLite:
		for k, v := range data {
			columns = append(columns, fmt.Sprintf("%s", k))
			placeholders = append(placeholders, "?")
			binds = append(binds, v)
		}

		sql = fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)", table, strings.Join(columns, ", "), strings.Join(placeholders, ", "))
	}

	return sql, binds
}

func singleInsertWithStruct(driver, table string, v reflect.Value) (string, []interface{}) {
	fieldNum := v.NumField()

	columns := make([]string, 0, fieldNum)
	placeholders := make([]string, 0, fieldNum)

	sql := ""
	binds := make([]interface{}, 0, fieldNum)

	t := v.Type()

	switch driver {
	case MySQL:
		for i := 0; i < fieldNum; i++ {
			column := t.Field(i).Tag.Get("db")

			if column == "-" {
				continue
			}

			if column == "" {
				column = t.Field(i).Name
			}

			columns = append(columns, fmt.Sprintf("`%s`", column))
			placeholders = append(placeholders, "?")
			binds = append(binds, v.Field(i).Interface())
		}

		sql = fmt.Sprintf("INSERT INTO `%s` (%s) VALUES (%s)", table, strings.Join(columns, ", "), strings.Join(placeholders, ", "))
	case Postgres:
		bindIndex := 0

		for i := 0; i < fieldNum; i++ {
			column := t.Field(i).Tag.Get("db")

			if column == "-" {
				continue
			}

			bindIndex++

			if column == "" {
				column = t.Field(i).Name
			}

			columns = append(columns, fmt.Sprintf(`"%s"`, column))
			placeholders = append(placeholders, fmt.Sprintf("$%d", bindIndex))
			binds = append(binds, v.Field(i).Interface())
		}

		sql = fmt.Sprintf(`INSERT INTO "%s" (%s) VALUES (%s) RETURNING "id"`, table, strings.Join(columns, ", "), strings.Join(placeholders, ", "))
	case SQLite:
		for i := 0; i < fieldNum; i++ {
			column := t.Field(i).Tag.Get("db")

			if column == "-" {
				continue
			}

			if column == "" {
				column = t.Field(i).Name
			}

			columns = append(columns, fmt.Sprintf("%s", column))
			placeholders = append(placeholders, "?")
			binds = append(binds, v.Field(i).Interface())
		}

		sql = fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)", table, strings.Join(columns, ", "), strings.Join(placeholders, ", "))
	}

	return sql, binds
}

func batchInsertWithMap(driver, table string, data []X, count int) (string, []interface{}) {
	fieldNum := len(data[0])

	fields := make([]string, 0, fieldNum)
	columns := make([]string, 0, fieldNum)
	placeholders := make([]string, 0, fieldNum)

	sql := ""
	binds := make([]interface{}, 0, fieldNum*count)

	switch driver {
	case MySQL:
		for k := range data[0] {
			fields = append(fields, k)

			columns = append(columns, fmt.Sprintf("`%s`", k))
		}

		for _, x := range data {
			phrs := make([]string, 0, fieldNum)

			for _, v := range fields {
				phrs = append(phrs, "?")
				binds = append(binds, x[v])
			}

			placeholders = append(placeholders, fmt.Sprintf("(%s)", strings.Join(phrs, ", ")))
		}

		sql = fmt.Sprintf("INSERT INTO `%s` (%s) VALUES %s", table, strings.Join(columns, ", "), strings.Join(placeholders, ", "))
	case Postgres:
		for k := range data[0] {
			fields = append(fields, k)

			columns = append(columns, fmt.Sprintf(`"%s"`, k))
		}

		bindIndex := 0

		for _, x := range data {
			phrs := make([]string, 0, fieldNum)

			for _, v := range fields {
				bindIndex++

				phrs = append(phrs, fmt.Sprintf("$%d", bindIndex))
				binds = append(binds, x[v])
			}

			placeholders = append(placeholders, fmt.Sprintf("(%s)", strings.Join(phrs, ", ")))
		}

		sql = fmt.Sprintf(`INSERT INTO "%s" (%s) VALUES %s`, table, strings.Join(columns, ", "), strings.Join(placeholders, ", "))
	}

	return sql, binds
}

func batchInsertWithStruct(driver, table string, v reflect.Value, count int) (string, []interface{}) {
	first := reflect.Indirect(v.Index(0))

	fieldNum := first.NumField()

	columns := make([]string, 0, fieldNum)
	placeholders := make([]string, 0, fieldNum)

	sql := ""
	binds := make([]interface{}, 0, fieldNum*count)

	t := first.Type()

	switch driver {
	case MySQL:
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

					columns = append(columns, fmt.Sprintf("`%s`", column))
				}

				phrs = append(phrs, "?")
				binds = append(binds, reflect.Indirect(v.Index(i)).Field(j).Interface())
			}

			placeholders = append(placeholders, fmt.Sprintf("(%s)", strings.Join(phrs, ", ")))
		}

		sql = fmt.Sprintf("INSERT INTO `%s` (%s) VALUES %s", table, strings.Join(columns, ", "), strings.Join(placeholders, ", "))
	case Postgres:
		bindIndex := 0

		for i := 0; i < count; i++ {
			phrs := make([]string, 0, fieldNum)

			for j := 0; j < fieldNum; j++ {
				column := t.Field(j).Tag.Get("db")

				if column == "-" {
					continue
				}

				bindIndex++

				if i == 0 {
					if column == "" {
						column = t.Field(j).Name
					}

					columns = append(columns, fmt.Sprintf(`"%s"`, column))
				}

				phrs = append(phrs, fmt.Sprintf("$%d", bindIndex))
				binds = append(binds, reflect.Indirect(v.Index(i)).Field(j).Interface())
			}

			placeholders = append(placeholders, fmt.Sprintf("(%s)", strings.Join(phrs, ", ")))
		}

		sql = fmt.Sprintf(`INSERT INTO "%s" (%s) VALUES %s`, table, strings.Join(columns, ", "), strings.Join(placeholders, ", "))
	}

	return sql, binds
}

func updateWithMap(driver, query string, data X, condition ...interface{}) (string, []interface{}) {
	dataLen := len(data)
	condLen := len(condition)

	sql := ""
	sets := make([]string, 0, dataLen)
	binds := make([]interface{}, 0, dataLen+condLen)

	switch driver {
	case MySQL:
		for k, v := range data {
			sets = append(sets, fmt.Sprintf("`%s` = ?", k))
			binds = append(binds, v)
		}

		sql = strings.Replace(query, "?", strings.Join(sets, ", "), 1)
		binds = append(binds, condition...)
	case Postgres:
		bindIndex := 0

		for k, v := range data {
			bindIndex++

			sets = append(sets, fmt.Sprintf(`"%s" = $%d`, k, bindIndex))
			binds = append(binds, v)
		}

		oldnew := make([]string, 0, condLen*2)

		for i := 1; i <= condLen; i++ {
			oldnew = append(oldnew, fmt.Sprintf("$%d", i+1), fmt.Sprintf("$%d", dataLen+i))
		}

		r := strings.NewReplacer(oldnew...)
		query = r.Replace(query)

		sql = strings.Replace(query, "$1", strings.Join(sets, ", "), 1)
		binds = append(binds, condition...)
	case SQLite:
		for k, v := range data {
			sets = append(sets, fmt.Sprintf("%s = ?", k))
			binds = append(binds, v)
		}

		sql = strings.Replace(query, "?", strings.Join(sets, ", "), 1)
		binds = append(binds, condition...)
	}

	return sql, binds
}

func updateWithStruct(driver, query string, v reflect.Value, condition ...interface{}) (string, []interface{}) {
	fieldNum := v.NumField()
	condLen := len(condition)

	sql := ""
	sets := make([]string, 0, fieldNum)
	binds := make([]interface{}, 0, fieldNum+condLen)

	t := v.Type()

	switch driver {
	case MySQL:
		for i := 0; i < fieldNum; i++ {
			column := t.Field(i).Tag.Get("db")

			if column == "-" {
				continue
			}

			if column == "" {
				column = t.Field(i).Name
			}

			sets = append(sets, fmt.Sprintf("`%s` = ?", column))
			binds = append(binds, v.Field(i).Interface())
		}

		sql = strings.Replace(query, "?", strings.Join(sets, ", "), 1)
		binds = append(binds, condition...)
	case Postgres:
		bindIndex := 0

		for i := 0; i < fieldNum; i++ {
			column := t.Field(i).Tag.Get("db")

			if column == "-" {
				continue
			}

			bindIndex++

			if column == "" {
				column = t.Field(i).Name
			}

			sets = append(sets, fmt.Sprintf(`"%s" = $%d`, column, bindIndex))
			binds = append(binds, v.Field(i).Interface())
		}

		oldnew := make([]string, 0, condLen*2)

		for i := 1; i <= condLen; i++ {
			oldnew = append(oldnew, fmt.Sprintf("$%d", i+1), fmt.Sprintf("$%d", bindIndex+i))
		}

		r := strings.NewReplacer(oldnew...)
		query = r.Replace(query)

		sql = strings.Replace(query, "$1", strings.Join(sets, ", "), 1)
		binds = append(binds, condition...)
	case SQLite:
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

		sql = strings.Replace(query, "?", strings.Join(sets, ", "), 1)
		binds = append(binds, condition...)
	}

	return sql, binds
}
