package yiigo

import (
	"errors"
	"fmt"
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

// DBConn get db connection
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
