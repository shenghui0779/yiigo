package yiigo

import (
	"fmt"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

var dbmap map[string]*sqlx.DB

// initMySQL init db connections
func initMySQL() error {
	dbmap = make(map[string]*sqlx.DB)

	sections := childSections("mysql")

	for _, v := range sections {
		host := v.Key("host").MustString("localhost")
		port := v.Key("post").MustInt(3306)
		username := v.Key("username").MustString("root")
		password := v.Key("password").MustString("")
		database := v.Key("database").MustString("test")
		charset := v.Key("charset").MustString("utf8mb4")
		collection := v.Key("collection").MustString("utf8mb4_general_ci")

		dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&collation=%s&parseTime=True&loc=Local", username, password, host, port, database, charset, collection)

		db, err := sqlx.Open("mysql", dsn)

		if err != nil {
			db.Close()
			return err
		}

		db.SetMaxOpenConns(v.Key("maxOpenConns").MustInt(20))
		db.SetMaxIdleConns(v.Key("maxIdleConns").MustInt(10))

		err = db.Ping()

		if err != nil {
			db.Close()
			return err
		}

		dbmap[v.Name()] = db
	}

	return nil
}

// DB get a db connection
func DB(connection ...string) (*sqlx.DB, error) {
	conn := "default"

	if len(connection) > 0 {
		conn = connection[0]
	}

	db, ok := dbmap[fmt.Sprintf("mysql.%s", conn)]

	if !ok {
		return nil, fmt.Errorf("database %s is not connected", conn)
	}

	return db, nil
}
