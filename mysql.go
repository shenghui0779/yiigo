package yiigo

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"sync"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	toml "github.com/pelletier/go-toml"
)

type mysqlConf struct {
	Name         string `toml:"name"`
	Host         string `toml:"host"`
	Port         int    `toml:"port"`
	Username     string `toml:"username"`
	Password     string `toml:"password"`
	Database     string `toml:"database"`
	Charset      string `toml:"charset"`
	Collection   string `toml:"collection"`
	MaxOpenConns int    `toml:"maxOpenConns"`
	MaxIdleConns int    `toml:"maxIdleConns"`
}

// SQLExpr SQL expression
type SQLExpr struct {
	Expr string
	Args []interface{}
}

var (
	// DB default connection
	DB    *sqlx.DB
	dbmap sync.Map
)

func initMySQL() error {
	var err error

	result := Env.Get("mysql")

	switch node := result.(type) {
	case *toml.Tree:
		conf := &mysqlConf{}
		err = node.Unmarshal(conf)

		if err != nil {
			break
		}

		err = initSingleDB(conf)
	case []*toml.Tree:
		conf := make([]*mysqlConf, 0, len(node))

		for _, v := range node {
			c := &mysqlConf{}
			err = v.Unmarshal(c)

			if err != nil {
				break
			}

			conf = append(conf, c)
		}

		err = initMultiDB(conf)
	default:
		return errors.New("mysql error config")
	}

	if err != nil {
		return fmt.Errorf("mysql error: %s", err.Error())
	}

	return nil
}

func initSingleDB(conf *mysqlConf) error {
	var err error

	DB, err = dbDial(conf)

	return err
}

func initMultiDB(conf []*mysqlConf) error {
	for _, v := range conf {
		db, err := dbDial(v)

		if err != nil {
			return err
		}

		dbmap.Store(v.Name, db)
	}

	if v, ok := dbmap.Load("default"); ok {
		DB = v.(*sqlx.DB)
	}

	return nil
}

func dbDial(conf *mysqlConf) (*sqlx.DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&collation=%s&parseTime=True&loc=Local", conf.Username, conf.Password, conf.Host, conf.Port, conf.Database, conf.Charset, conf.Collection)

	db, err := sqlx.Connect("mysql", dsn)

	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(conf.MaxOpenConns)
	db.SetMaxIdleConns(conf.MaxIdleConns)

	return db, nil
}

// DBConn returns a db connection.
func DBConn(conn ...string) (*sqlx.DB, error) {
	schema := "default"

	if len(conn) > 0 {
		schema = conn[0]
	}

	v, ok := dbmap.Load(schema)

	if !ok {
		return nil, fmt.Errorf("mysql %s is not connected", schema)
	}

	return v.(*sqlx.DB), nil
}

// InsertSQL returns insert sql and binds.
// param data expects struct, []struct, yiigo.X, []yiigo.X.
func InsertSQL(table string, data interface{}) (string, []interface{}) {
	v := reflect.Indirect(reflect.ValueOf(data))

	sql := ""
	binds := []interface{}{}

	switch v.Kind() {
	case reflect.Map:
		if x, ok := data.(X); ok {
			sql, binds = singleInsertWithMap(table, x)
		}
	case reflect.Struct:
		sql, binds = singleInsertWithStruct(table, v)
	case reflect.Slice:
		if count := v.Len(); count > 0 {
			elemKind := v.Type().Elem().Kind()

			if elemKind == reflect.Map {
				if x, ok := data.([]X); ok {
					sql, binds = batchInsertWithMap(table, x, count)
				}

				break
			}

			if elemKind == reflect.Struct {
				sql, binds = batchInsertWithStruct(table, v, count)

				break
			}
		}
	}

	return sql, binds
}

// UpdateSQL returns update sql and binds.
// param data expects struct, yiigo.X.
func UpdateSQL(query string, data interface{}, args ...interface{}) (string, []interface{}) {
	v := reflect.Indirect(reflect.ValueOf(data))

	sql := ""
	binds := []interface{}{}

	switch v.Kind() {
	case reflect.Map:
		if x, ok := data.(X); ok {
			sql, binds = updateWithMap(query, x, args...)
		}
	case reflect.Struct:
		sql, binds = updateWithStruct(query, v, args...)
	}

	return sql, binds
}

// Expr returns an expression, eg: yiigo.Expr("price * ? + ?", 2, 100).
func Expr(expr string, args ...interface{}) *SQLExpr {
	return &SQLExpr{Expr: expr, Args: args}
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

	fmt.Println(columns)

	for _, x := range data {
		phrs := make([]string, 0, fieldNum)

		for _, v := range fields {
			phrs = append(phrs, "?")
			binds = append(binds, x[v])
		}

		placeholders = append(placeholders, fmt.Sprintf("(%s)", strings.Join(phrs, ", ")))
	}

	sql := fmt.Sprintf("INSERT INTO `%s` (%s) VALUES %s", table, strings.Join(columns, ", "), strings.Join(placeholders, ","))

	return sql, binds
}

func batchInsertWithStruct(table string, v reflect.Value, count int) (string, []interface{}) {
	first := v.Index(0)

	if first.Kind() != reflect.Struct {
		panic("the data must be a slice to struct")
	}

	fieldNum := first.NumField()

	columns := make([]string, 0, fieldNum)
	placeholders := make([]string, 0, fieldNum)
	binds := make([]interface{}, 0, fieldNum*count)

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

				columns = append(columns, fmt.Sprintf("`%s`", column))
			}

			phrs = append(phrs, "?")
			binds = append(binds, v.Index(i).Field(j).Interface())
		}

		placeholders = append(placeholders, fmt.Sprintf("(%s)", strings.Join(phrs, ", ")))
	}

	sql := fmt.Sprintf("INSERT INTO `%s` (%s) VALUES %s", table, strings.Join(columns, ", "), strings.Join(placeholders, ","))

	return sql, binds
}

func updateWithMap(query string, data X, args ...interface{}) (string, []interface{}) {
	sets := []string{}
	binds := []interface{}{}

	for k, v := range data {
		if e, ok := v.(*SQLExpr); ok {
			sets = append(sets, fmt.Sprintf("`%s` = %s", k, e.Expr))
			binds = append(binds, e.Args...)
		} else {
			sets = append(sets, fmt.Sprintf("`%s` = ?", k))
			binds = append(binds, v)
		}
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

		if column == "-" {
			continue
		}

		if column == "" {
			column = t.Field(i).Name
		}

		field := v.Field(i).Interface()

		if e, ok := field.(*SQLExpr); ok {
			sets = append(sets, fmt.Sprintf("`%s` = %s", column, e.Expr))
			binds = append(binds, e.Args...)
		} else {
			sets = append(sets, fmt.Sprintf("`%s` = ?", column))
			binds = append(binds, field)
		}
	}

	sql := strings.Replace(query, "?", strings.Join(sets, ", "), 1)
	binds = append(binds, args...)

	return sql, binds
}
