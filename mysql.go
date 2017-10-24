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
	DB    *sqlx.DB
	dbmap map[string]*sqlx.DB
	dbmux sync.RWMutex
)

// SQL expression
type expr struct {
	expr string
	args []interface{}
}

// initMySQL init mysql connection
func initMySQL() error {
	sections := childSections("mysql")

	if len(sections) > 0 {
		return initMultiDB(sections)
	}

	return initSingleDB()
}

// initSingleDB init single db connection
func initSingleDB() error {
	var err error

	host := EnvString("mysql", "host", "localhost")
	port := EnvInt("mysql", "port", 3306)
	username := EnvString("mysql", "username", "root")
	password := EnvString("mysql", "password", "")
	database := EnvString("mysql", "database", "test")
	charset := EnvString("mysql", "charset", "utf8mb4")
	collection := EnvString("mysql", "collection", "utf8mb4_general_ci")

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&collation=%s&parseTime=True&loc=Local", username, password, host, port, database, charset, collection)

	DB, err = sqlx.Connect("mysql", dsn)

	if err != nil {
		DB.Close()
		return err
	}

	DB.SetMaxOpenConns(EnvInt("mysql", "maxOpenConns", 20))
	DB.SetMaxIdleConns(EnvInt("mysql", "maxIdleConns", 10))

	return nil
}

// initMultiDB init multi db connections
func initMultiDB(sections []*ini.Section) error {
	dbmap = make(map[string]*sqlx.DB, len(sections))

	for _, v := range sections {
		host := v.Key("host").MustString("localhost")
		port := v.Key("port").MustInt(3306)
		username := v.Key("username").MustString("root")
		password := v.Key("password").MustString("")
		database := v.Key("database").MustString("test")
		charset := v.Key("charset").MustString("utf8mb4")
		collection := v.Key("collection").MustString("utf8mb4_general_ci")

		dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&collation=%s&parseTime=True&loc=Local", username, password, host, port, database, charset, collection)

		db, err := sqlx.Connect("mysql", dsn)

		if err != nil {
			db.Close()
			return err
		}

		db.SetMaxOpenConns(v.Key("maxOpenConns").MustInt(20))
		db.SetMaxIdleConns(v.Key("maxIdleConns").MustInt(10))

		dbmap[v.Name()] = db
	}

	if db, ok := dbmap["mysql.default"]; ok {
		DB = db
	}

	return nil
}

// DBConn get db
func DBConn(conn ...string) (*sqlx.DB, error) {
	dbmux.RLock()
	defer dbmux.RUnlock()

	c := "default"

	if len(conn) > 0 {
		c = conn[0]
	}

	schema := fmt.Sprintf("mysql.%s", c)

	db, ok := dbmap[schema]

	if !ok {
		return nil, fmt.Errorf("database %s is not connected", schema)
	}

	return db, nil
}

// InsertSQL returns insert sql and binds
func InsertSQL(table string, data interface{}) (string, []interface{}) {
	v := reflect.Indirect(reflect.ValueOf(data))

	sql := ""
	binds := []interface{}{}

	switch v.Kind() {
	case reflect.Struct:
		sql, binds = singleInsert(table, v)
	case reflect.Slice:
		sql, binds = batchInsert(table, v)
	}

	return sql, binds
}

// UpdateSQL returns update sql and binds
func UpdateSQL(sql string, data interface{}, args ...interface{}) (string, []interface{}) {
	v := reflect.Indirect(reflect.ValueOf(data))

	_sql := ""
	binds := []interface{}{}

	switch v.Kind() {
	case reflect.Map:
		if x, ok := data.(X); ok {
			_sql, binds = updateByMap(sql, x, args...)
		}
	case reflect.Struct:
		_sql, binds = updateByStruct(sql, v, args...)
	}

	return _sql, binds
}

// Expr returns expression, eg: yiigo.Expr("price * ? + ?", 2, 100)
func Expr(expression string, args ...interface{}) *expr {
	return &expr{expr: expression, args: args}
}

func singleInsert(table string, v reflect.Value) (string, []interface{}) {
	fieldNum := v.NumField()

	columns := make([]string, 0, fieldNum)
	placeholders := make([]string, 0, fieldNum)
	binds := make([]interface{}, 0, fieldNum)

	t := v.Type()

	for i := 0; i < fieldNum; i++ {
		columns = append(columns, t.Field(i).Tag.Get("db"))
		placeholders = append(placeholders, "?")
		binds = append(binds, v.Field(i))
	}

	sql := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)", table, strings.Join(columns, ","), strings.Join(placeholders, ","))

	return sql, binds
}

func batchInsert(table string, v reflect.Value) (string, []interface{}) {
	count := v.Len()
	fieldNum := reflect.Indirect(v.Index(0)).NumField()

	columns := make([]string, 0, fieldNum)
	placeholders := make([]string, 0, fieldNum)
	binds := make([]interface{}, 0, fieldNum*count)

	t := reflect.Indirect(v.Index(0)).Type()

	for i := 0; i < fieldNum; i++ {
		columns = append(columns, t.Field(i).Tag.Get("db"))
	}

	for i := 0; i < count; i++ {
		phrs := make([]string, 0, fieldNum)

		for j := 0; j < fieldNum; j++ {
			phrs = append(phrs, "?")
			binds = append(binds, reflect.Indirect(v.Index(i)).Field(j))
		}

		placeholders = append(placeholders, fmt.Sprintf("(%s)", strings.Join(phrs, ",")))
	}

	sql := fmt.Sprintf("INSERT INTO %s (%s) VALUES %s", table, strings.Join(columns, ","), strings.Join(placeholders, ","))

	return sql, binds
}

func updateByMap(sql string, data X, args ...interface{}) (string, []interface{}) {
	sets := []string{}
	binds := []interface{}{}

	for k, v := range data {
		if e, ok := v.(*expr); ok {
			sets = append(sets, fmt.Sprintf("%s = %s", k, e.expr))
			binds = append(binds, e.args...)
		} else {
			sets = append(sets, fmt.Sprintf("%s = ?", k))
			binds = append(binds, v)
		}
	}

	sql = strings.Replace(sql, "?", strings.Join(sets, ", "), 1)
	binds = append(binds, args...)

	return sql, binds
}

func updateByStruct(sql string, v reflect.Value, args ...interface{}) (string, []interface{}) {
	fieldNum := v.NumField()

	sets := make([]string, 0, fieldNum)
	binds := make([]interface{}, 0, fieldNum+len(args))

	t := v.Type()

	for i := 0; i < fieldNum; i++ {
		sets = append(sets, fmt.Sprintf("%s = ?", t.Field(i).Tag.Get("db")))
		binds = append(binds, v.Field(i))
	}

	sql = strings.Replace(sql, "?", strings.Join(sets, ", "), 1)
	binds = append(binds, args...)

	return sql, binds
}
