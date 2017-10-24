package yiigo

import (
	"fmt"
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

// InsertSQL returns insert sql and binds
func InsertSQL(table string, data X) (string, []interface{}) {
	length := len(data)

	columns := make([]string, length)
	placeholders := make([]string, length)
	binds := make([]interface{}, length)

	i := 0

	for k, v := range data {
		columns[i] = k
		placeholders[i] = "?"
		binds[i] = v

		i++
	}

	sql := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)", table, strings.Join(columns[1:], ","), strings.Join(placeholders[1:], ","))

	return sql, binds
}

// BatchInsertSQL returns batch insert sql and binds
func BatchInsertSQL(table string, data []X) (string, []interface{}) {
	length := len(data)

	if length == 0 {
		return "", nil
	}

	keyNum := len(data[0])

	columns := make([]string, keyNum)
	placeholders := []string{}
	binds := []interface{}{}

	i := 0
	phrs := make([]string, keyNum)
	args := make([]interface{}, keyNum)

	for k, v := range data[0] {
		columns[i] = k
		phrs[i] = "?"
		args[i] = v

		i++
	}

	placeholders = append(placeholders, fmt.Sprintf("(%s)", strings.Join(phrs, ",")))
	binds = append(binds, args...)

	for i := 1; i < length; i++ {
		phrs := make([]string, keyNum)
		args := make([]interface{}, keyNum)

		for j, v := range columns {
			phrs[j] = "?"
			args[j] = data[i][v]
		}

		placeholders = append(placeholders, fmt.Sprintf("(%s)", strings.Join(phrs, ",")))
		binds = append(binds, args...)
	}

	sql := fmt.Sprintf("INSERT INTO %s (%s) VALUES %s", table, strings.Join(columns, ","), strings.Join(placeholders, ","))

	return sql, binds
}

// UpdateSQL returns update sql and binds
func UpdateSQL(sql string, data X, args ...interface{}) (string, []interface{}) {
	sets := []string{}
	binds := []interface{}{}

	for k, v := range data {
		if expr, ok := v.(*expr); ok {
			sets = append(sets, fmt.Sprintf("%s = %s", k, expr.expr))
			binds = append(binds, expr.args...)
		} else {
			sets = append(sets, fmt.Sprintf("%s = ?", k))
			binds = append(binds, v)
		}
	}

	sql = strings.Replace(sql, "?", strings.Join(sets, ", "), 1)
	binds = append(binds, args...)

	return sql, binds
}

// Expr returns expression, eg: yiigo.Expr("price * ? + ?", 2, 100)
func Expr(expression string, args ...interface{}) *expr {
	return &expr{expr: expression, args: args}
}
