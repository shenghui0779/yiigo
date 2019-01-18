package yiigo

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"sync"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	toml "github.com/pelletier/go-toml"
)

type mysqlConf struct {
	Name            string `toml:"name"`
	Host            string `toml:"host"`
	Port            int    `toml:"port"`
	Username        string `toml:"username"`
	Password        string `toml:"password"`
	Database        string `toml:"database"`
	Charset         string `toml:"charset"`
	Collection      string `toml:"collection"`
	ConnTimeout     int    `toml:"connTimeout"`
	MaxOpenConns    int    `toml:"maxOpenConns"`
	MaxIdleConns    int    `toml:"maxIdleConns"`
	ConnMaxLifetime int    `toml:"connMaxLifetime"`
}

var (
	// DB default mysql connection
	DB    *sqlx.DB
	dbmap sync.Map
)

var errInsertInvalidType = errors.New("yiigo: invalid data type for InsertSQL(), expects: struct, *struct, []struct, []*struct, yiigo.X, []yiigo.X")
var errUpdateInvalidType = errors.New("yiigo: invalid data type for UpdateSQL(), expects: struct, *struct, yiigo.X")

// initMySQL init MySQL
func initMySQL() error {
	result := Env.Get("mysql")

	if result == nil {
		return nil
	}

	switch node := result.(type) {
	case *toml.Tree:
		conf := new(mysqlConf)

		if err := node.Unmarshal(conf); err != nil {
			return err
		}

		if err := initSingleDB(conf); err != nil {
			return err
		}
	case []*toml.Tree:
		conf := make([]*mysqlConf, 0, len(node))

		for _, v := range node {
			c := new(mysqlConf)

			if err := v.Unmarshal(c); err != nil {
				return err
			}

			conf = append(conf, c)
		}

		if err := initMultiDB(conf); err != nil {
			return err
		}
	default:
		return errors.New("yiigo: invalid mysql config")
	}

	return nil
}

func initSingleDB(conf *mysqlConf) error {
	var err error

	DB, err = dbDial(conf)

	if err != nil {
		return fmt.Errorf("yiigo: mysql.default connect error: %s", err.Error())
	}

	dbmap.Store("default", DB)

	return nil
}

func initMultiDB(conf []*mysqlConf) error {
	for _, v := range conf {
		db, err := dbDial(v)

		if err != nil {
			return fmt.Errorf("yiigo: mysql.%s connect error: %s", v.Name, err.Error())
		}

		dbmap.Store(v.Name, db)
	}

	if v, ok := dbmap.Load("default"); ok {
		DB = v.(*sqlx.DB)
	}

	return nil
}

func dbDial(conf *mysqlConf) (*sqlx.DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?timeout=%ds&charset=%s&collation=%s&parseTime=True&loc=Local", conf.Username, conf.Password, conf.Host, conf.Port, conf.Database, conf.ConnTimeout, conf.Charset, conf.Collection)

	db, err := sqlx.Connect("mysql", dsn)

	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(conf.MaxOpenConns)
	db.SetMaxIdleConns(conf.MaxIdleConns)
	db.SetConnMaxLifetime(time.Duration(conf.ConnMaxLifetime) * time.Second)

	return db, nil
}

// DBConn returns a mysql connection.
func DBConn(conn ...string) (*sqlx.DB, error) {
	schema := "default"

	if len(conn) > 0 {
		schema = conn[0]
	}

	v, ok := dbmap.Load(schema)

	if !ok {
		return nil, fmt.Errorf("yiigo: mysql.%s is not connected", schema)
	}

	return v.(*sqlx.DB), nil
}

// InsertSQL returns mysql insert sql and binds.
// param data expects struct, *struct, []struct, []*struct, yiigo.X, []yiigo.X.
func InsertSQL(table string, data interface{}) (string, []interface{}) {
	var (
		sql   string
		binds []interface{}
	)

	v := reflect.Indirect(reflect.ValueOf(data))

	switch v.Kind() {
	case reflect.Map:
		if x, ok := data.(X); ok {
			sql, binds = singleInsertWithMap(table, x)
		}
	case reflect.Struct:
		sql, binds = singleInsertWithStruct(table, v)
	case reflect.Slice:
		if count := v.Len(); count > 0 {
			e := v.Type().Elem()

			switch e.Kind() {
			case reflect.Map:
				x, ok := data.([]X)

				if !ok {
					panic(errInsertInvalidType)
				}

				sql, binds = batchInsertWithMap(table, x, count)
			case reflect.Struct:
				sql, binds = batchInsertWithStruct(table, v, count)
			case reflect.Ptr:
				if e.Elem().Kind() != reflect.Struct {
					panic(errInsertInvalidType)
				}

				sql, binds = batchInsertWithStruct(table, v, count)
			default:
				panic(errInsertInvalidType)
			}
		}
	default:
		panic(errInsertInvalidType)
	}

	return sql, binds
}

// UpdateSQL returns mysql update sql and binds.
// param query expects eg: "UPDATE `table` SET ? WHERE id = ?".
// param data expects struct, *struct, yiigo.X.
func UpdateSQL(query string, data interface{}, args ...interface{}) (string, []interface{}) {
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

		sql, binds = updateWithMap(query, x, args...)
	case reflect.Struct:
		sql, binds = updateWithStruct(query, v, args...)
	default:
		panic(errUpdateInvalidType)
	}

	return sql, binds
}

func singleInsertWithMap(table string, data X) (string, []interface{}) {
	fieldNum := len(data)

	columns := make([]string, 0, fieldNum)
	placeholders := make([]string, 0, fieldNum)
	binds := make([]interface{}, 0, fieldNum)

	for k, v := range data {
		columns = append(columns, fmt.Sprintf("`%s`", k))
		placeholders = append(placeholders, "?")
		binds = append(binds, v)
	}

	sql := fmt.Sprintf("INSERT INTO `%s` (%s) VALUES (%s)", table, strings.Join(columns, ", "), strings.Join(placeholders, ", "))

	return sql, binds
}

func singleInsertWithStruct(table string, v reflect.Value) (string, []interface{}) {
	fieldNum := v.NumField()

	columns := make([]string, 0, fieldNum)
	placeholders := make([]string, 0, fieldNum)
	binds := make([]interface{}, 0, fieldNum)

	t := v.Type()

	for i := 0; i < fieldNum; i++ {
		column := t.Field(i).Tag.Get("db")

		if column == "" {
			column = t.Field(i).Name
		}

		columns = append(columns, fmt.Sprintf("`%s`", column))
		placeholders = append(placeholders, "?")
		binds = append(binds, v.Field(i).Interface())
	}

	sql := fmt.Sprintf("INSERT INTO `%s` (%s) VALUES (%s)", table, strings.Join(columns, ", "), strings.Join(placeholders, ", "))

	return sql, binds
}

func batchInsertWithMap(table string, data []X, count int) (string, []interface{}) {
	fieldNum := len(data[0])

	fields := make([]string, 0, fieldNum)
	columns := make([]string, 0, fieldNum)
	placeholders := make([]string, 0, fieldNum)
	binds := make([]interface{}, 0, fieldNum*count)

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

	sql := fmt.Sprintf("INSERT INTO `%s` (%s) VALUES %s", table, strings.Join(columns, ", "), strings.Join(placeholders, ", "))

	return sql, binds
}

func batchInsertWithStruct(table string, v reflect.Value, count int) (string, []interface{}) {
	first := reflect.Indirect(v.Index(0))

	fieldNum := first.NumField()

	columns := make([]string, 0, fieldNum)
	placeholders := make([]string, 0, fieldNum)
	binds := make([]interface{}, 0, fieldNum*count)

	t := first.Type()

	for i := 0; i < count; i++ {
		phrs := make([]string, 0, fieldNum)

		for j := 0; j < fieldNum; j++ {
			column := t.Field(j).Tag.Get("db")

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

	sql := fmt.Sprintf("INSERT INTO `%s` (%s) VALUES %s", table, strings.Join(columns, ", "), strings.Join(placeholders, ", "))

	return sql, binds
}

func updateWithMap(query string, data X, args ...interface{}) (string, []interface{}) {
	dataLen := len(data)

	sets := make([]string, 0, dataLen)
	binds := make([]interface{}, 0, dataLen+len(args))

	for k, v := range data {
		sets = append(sets, fmt.Sprintf("`%s` = ?", k))
		binds = append(binds, v)
	}

	sql := strings.Replace(query, "?", strings.Join(sets, ", "), 1)
	binds = append(binds, args...)

	return sql, binds
}

func updateWithStruct(query string, v reflect.Value, args ...interface{}) (string, []interface{}) {
	fieldNum := v.NumField()

	sets := make([]string, 0, fieldNum)
	binds := make([]interface{}, 0, fieldNum+len(args))

	t := v.Type()

	for i := 0; i < fieldNum; i++ {
		column := t.Field(i).Tag.Get("db")

		if column == "" {
			column = t.Field(i).Name
		}

		sets = append(sets, fmt.Sprintf("`%s` = ?", column))
		binds = append(binds, v.Field(i).Interface())
	}

	sql := strings.Replace(query, "?", strings.Join(sets, ", "), 1)
	binds = append(binds, args...)

	return sql, binds
}
