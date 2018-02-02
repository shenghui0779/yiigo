package yiigo

import (
	"fmt"
	"reflect"
	"strings"
	"sync"

	"gopkg.in/ini.v1"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

var (
	// DB default connection
	DB    *sqlx.DB
	dbmap sync.Map
)

// SQLExpr SQL expression
type SQLExpr struct {
	Expr string
	Args []interface{}
}

func initMySQL() error {
	sections := childSections("mysql")

	if len(sections) > 0 {
		return initMultiDB(sections)
	}

	return initSingleDB()
}

func initSingleDB() error {
	var err error

	section := env.Section("mysql")

	DB, err = dbDial(section)

	if err != nil {
		return fmt.Errorf("mysql error: %s", err.Error())
	}

	return nil
}

func initMultiDB(sections []*ini.Section) error {
	for _, v := range sections {
		db, err := dbDial(v)

		if err != nil {
			return fmt.Errorf("mysql error: %s", err.Error())
		}

		dbmap.Store(v.Name(), db)
	}

	if v, ok := dbmap.Load("mysql.default"); ok {
		DB = v.(*sqlx.DB)
	}

	return nil
}

func dbDial(section *ini.Section) (*sqlx.DB, error) {
	host := section.Key("host").MustString("localhost")
	port := section.Key("port").MustInt(3306)
	username := section.Key("username").MustString("root")
	password := section.Key("password").MustString("")
	database := section.Key("database").MustString("test")
	charset := section.Key("charset").MustString("utf8mb4")
	collection := section.Key("collection").MustString("utf8mb4_general_ci")
	maxOpenConns := section.Key("maxOpenConns").MustInt(20)
	maxIdleConns := section.Key("maxIdleConns").MustInt(10)

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&collation=%s&parseTime=True&loc=Local", username, password, host, port, database, charset, collection)

	db, err := sqlx.Connect("mysql", dsn)

	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(maxOpenConns)
	db.SetMaxIdleConns(maxIdleConns)

	return db, nil
}

// DBConn get db connection
func DBConn(conn ...string) (*sqlx.DB, error) {
	c := "default"

	if len(conn) > 0 {
		c = conn[0]
	}

	schema := fmt.Sprintf("mysql.%s", c)

	v, ok := dbmap.Load(schema)

	if !ok {
		return nil, fmt.Errorf("mysql %s is not connected", schema)
	}

	return v.(*sqlx.DB), nil
}

// InsertSQL returns insert sql and binds
// data expect struct, []struct, yiigo.X, []yiigo.X
func InsertSQL(table string, data interface{}) (string, []interface{}) {
	v := reflect.Indirect(reflect.ValueOf(data))

	sql := ""
	binds := []interface{}{}

	switch v.Kind() {
	case reflect.Map:
		if x, ok := data.(X); ok {
			sql, binds = singleInsertWithMap(sql, x)
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

// UpdateSQL returns update sql and binds
// data expect struct, yiigo.X
func UpdateSQL(sql string, data interface{}, args ...interface{}) (string, []interface{}) {
	v := reflect.Indirect(reflect.ValueOf(data))

	_sql := ""
	binds := []interface{}{}

	switch v.Kind() {
	case reflect.Map:
		if x, ok := data.(X); ok {
			_sql, binds = updateWithMap(sql, x, args...)
		}
	case reflect.Struct:
		_sql, binds = updateWithStruct(sql, v, args...)
	}

	return _sql, binds
}

// Expr returns expression, eg: yiigo.Expr("price * ? + ?", 2, 100)
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

func batchInsertWithStruct(table string, v reflect.Value, count int) (string, []interface{}) {
	first := reflect.Indirect(v.Index(0))

	if first.Kind() != reflect.Struct {
		panic("the data must be a slice to struct")
	}

	fieldNum := first.NumField()

	columns := make([]string, 0, fieldNum)
	placeholders := make([]string, 0, fieldNum)
	binds := make([]interface{}, 0, fieldNum*count)

	t := first.Type()

	for i := 0; i < fieldNum; i++ {
		column := t.Field(i).Tag.Get("db")

		if column == "" {
			column = t.Field(i).Name
		}

		columns = append(columns, fmt.Sprintf("`%s`", column))
	}

	for i := 0; i < count; i++ {
		phrs := make([]string, 0, fieldNum)

		for j := 0; j < fieldNum; j++ {
			phrs = append(phrs, "?")
			binds = append(binds, reflect.Indirect(v.Index(i)).Field(j).Interface())
		}

		placeholders = append(placeholders, fmt.Sprintf("(%s)", strings.Join(phrs, ", ")))
	}

	sql := fmt.Sprintf("INSERT INTO `%s` (%s) VALUES %s", table, strings.Join(columns, ", "), strings.Join(placeholders, ","))

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

func updateWithMap(sql string, data X, args ...interface{}) (string, []interface{}) {
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

	sql = strings.Replace(sql, "?", strings.Join(sets, ", "), 1)
	binds = append(binds, args...)

	return sql, binds
}

func updateWithStruct(sql string, v reflect.Value, args ...interface{}) (string, []interface{}) {
	fieldNum := v.NumField()

	sets := make([]string, 0, fieldNum)
	binds := make([]interface{}, 0, fieldNum+len(args))

	t := v.Type()

	for i := 0; i < fieldNum; i++ {
		column := t.Field(i).Tag.Get("db")

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

	sql = strings.Replace(sql, "?", strings.Join(sets, ", "), 1)
	binds = append(binds, args...)

	return sql, binds
}
