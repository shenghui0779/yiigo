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
	dbmap = make(map[string]*sqlx.DB)

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

// InsertSQL returns insert sql and binds, the first field of data is defaulted as the primary key
func InsertSQL(table string, data interface{}) (string, []interface{}) {
	v := reflect.Indirect(reflect.ValueOf(data))

	if v.Kind() == reflect.Slice {
		if v.Len() == 0 {
			return "", nil
		}

		return batchInsert(table, v)
	}

	return singleInsert(table, v)
}

// UpdateSQL returns update sql and binds
func UpdateSQL(sql string, args ...interface{}) (string, []interface{}) {
	sets := []string{}
	binds := []interface{}{}

	for k, v := range args[0].(X) {
		if expr, ok := v.(*expr); ok {
			sets = append(sets, fmt.Sprintf("%s = %s", k, expr.expr))
			binds = append(binds, expr.args...)
		} else {
			sets = append(sets, fmt.Sprintf("%s = ?", k))
			binds = append(binds, v)
		}
	}

	sql = strings.Replace(sql, "?", strings.Join(sets, ", "), 1)
	binds = append(binds, args[1:]...)

	return sql, binds
}

// Expr returns expression, eg: yiigo.Expr("price * ? + ?", 2, 100)
func Expr(expression string, args ...interface{}) *expr {
	return &expr{expr: expression, args: args}
}

func singleInsert(table string, v reflect.Value) (string, []interface{}) {
	t := v.Type()
	fieldNum := v.NumField()

	columns := make([]string, fieldNum)
	placeholders := make([]string, fieldNum)
	binds := make([]interface{}, fieldNum)

	for i := 0; i < fieldNum; i++ {
		columns[i] = t.Field(i).Tag.Get("db")
		placeholders[i] = "?"
		binds[i] = v.Field(i)
	}

	sql := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)", table, strings.Join(columns[1:len(columns)], ","), strings.Join(placeholders[1:len(placeholders)], ","))

	return sql, binds
}

func batchInsert(table string, v reflect.Value) (string, []interface{}) {
	fieldNum := v.Index(0).NumField()
	t := v.Index(0).Type()

	columns := make([]string, fieldNum)

	for i := 0; i < fieldNum; i++ {
		columns[i] = t.Field(i).Tag.Get("db")
	}

	placeholders := []string{}
	binds := []interface{}{}

	for i := 0; i < v.Len(); i++ {
		phrs := make([]string, fieldNum)
		args := make([]interface{}, fieldNum)

		for j := 0; j < fieldNum; j++ {
			phrs[j] = "?"
			args[j] = v.Index(i).Field(j)
		}

		placeholders = append(placeholders, fmt.Sprintf("(%s)", strings.Join(phrs[1:len(phrs)], ",")))
		binds = append(binds, args[1:len(args)]...)
	}

	sql := fmt.Sprintf("INSERT INTO %s (%s) VALUES %s", table, strings.Join(columns[1:len(columns)], ","), strings.Join(placeholders, ","))

	return sql, binds
}
