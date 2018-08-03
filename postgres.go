package yiigo

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	toml "github.com/pelletier/go-toml"
)

type postgresConf struct {
	Name            string `toml:"name"`
	Host            string `toml:"host"`
	Port            int    `toml:"port"`
	Username        string `toml:"username"`
	Password        string `toml:"password"`
	Database        string `toml:"database"`
	MaxOpenConns    int    `toml:"maxOpenConns"`
	MaxIdleConns    int    `toml:"maxIdleConns"`
	ConnMaxLifetime int    `toml:"connMaxLifetime"`
}

var (
	// PG default postgres connection
	PG    *sqlx.DB
	pgmap sync.Map
)

func initPostgres() error {
	var err error

	result := Env.Get("postgres")

	switch node := result.(type) {
	case *toml.Tree:
		conf := &postgresConf{}
		err = node.Unmarshal(conf)

		if err != nil {
			break
		}

		err = initSinglePostgres(conf)
	case []*toml.Tree:
		conf := make([]*postgresConf, 0, len(node))

		for _, v := range node {
			c := &postgresConf{}
			err = v.Unmarshal(c)

			if err != nil {
				break
			}

			conf = append(conf, c)
		}

		err = initMultiPostgres(conf)
	default:
		return errors.New("postgres error config")
	}

	if err != nil {
		return fmt.Errorf("postgres error: %s", err.Error())
	}

	return nil
}

func initSinglePostgres(conf *postgresConf) error {
	var err error

	PG, err = postgresDial(conf)

	if err != nil {
		return err
	}

	pgmap.Store("default", PG)

	return nil
}

func initMultiPostgres(conf []*postgresConf) error {
	for _, v := range conf {
		db, err := postgresDial(v)

		if err != nil {
			return err
		}

		pgmap.Store(v.Name, db)
	}

	if v, ok := pgmap.Load("default"); ok {
		PG = v.(*sqlx.DB)
	}

	return nil
}

func postgresDial(conf *postgresConf) (*sqlx.DB, error) {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", conf.Host, conf.Port, conf.Username, conf.Password, conf.Database)

	db, err := sqlx.Connect("postgres", dsn)

	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(conf.MaxOpenConns)
	db.SetMaxIdleConns(conf.MaxIdleConns)
	db.SetConnMaxLifetime(time.Duration(conf.ConnMaxLifetime) * time.Second)

	return db, nil
}

// PGConn returns a postgres connection.
func PGConn(conn ...string) (*sqlx.DB, error) {
	schema := "default"

	if len(conn) > 0 {
		schema = conn[0]
	}

	v, ok := pgmap.Load(schema)

	if !ok {
		return nil, fmt.Errorf("postgres %s is not connected", schema)
	}

	return v.(*sqlx.DB), nil
}

// PGInsertSQL returns postgres insert sql and binds.
// param data expects struct, []struct, yiigo.X, []yiigo.X.
func PGInsertSQL(table string, data interface{}) (string, []interface{}) {
	var (
		sql   string
		binds []interface{}
	)

	v := reflect.Indirect(reflect.ValueOf(data))

	switch v.Kind() {
	case reflect.Map:
		if x, ok := data.(X); ok {
			sql, binds = singlePGInsertWithMap(table, x)
		}
	case reflect.Struct:
		sql, binds = singlePGInsertWithStruct(table, v)
	case reflect.Slice:
		if count := v.Len(); count > 0 {
			elemKind := v.Type().Elem().Kind()

			if elemKind == reflect.Map {
				if x, ok := data.([]X); ok {
					sql, binds = batchPGInsertWithMap(table, x, count)
				}

				break
			}

			if elemKind == reflect.Struct {
				sql, binds = batchPGInsertWithStruct(table, v, count)

				break
			}
		}
	}

	return sql, binds
}

// PGUpdateSQL returns postgres update sql and binds.
// param data expects struct, yiigo.X.
func PGUpdateSQL(query string, data interface{}, args ...interface{}) (string, []interface{}) {
	var (
		sql   string
		binds []interface{}
	)

	v := reflect.Indirect(reflect.ValueOf(data))

	switch v.Kind() {
	case reflect.Map:
		if x, ok := data.(X); ok {
			sql, binds = pgUpdateWithMap(query, x, args...)
		}
	case reflect.Struct:
		sql, binds = pgUpdateWithStruct(query, v, args...)
	}

	return sql, binds
}

func singlePGInsertWithMap(table string, data X) (string, []interface{}) {
	fieldNum := len(data)

	columns := make([]string, 0, fieldNum)
	placeholders := make([]string, 0, fieldNum)
	binds := make([]interface{}, 0, fieldNum)

	bindIndex := 0

	for k, v := range data {
		bindIndex++
		columns = append(columns, fmt.Sprintf("`%s`", k))
		placeholders = append(placeholders, fmt.Sprintf("$%d", bindIndex))
		binds = append(binds, v)
	}

	sql := fmt.Sprintf("INSERT INTO `%s` (%s) VALUES (%s)", table, strings.Join(columns, ", "), strings.Join(placeholders, ", "))

	return sql, binds
}

func singlePGInsertWithStruct(table string, v reflect.Value) (string, []interface{}) {
	fieldNum := v.NumField()

	columns := make([]string, 0, fieldNum)
	placeholders := make([]string, 0, fieldNum)
	binds := make([]interface{}, 0, fieldNum)

	t := v.Type()
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

		columns = append(columns, fmt.Sprintf("`%s`", column))
		placeholders = append(placeholders, fmt.Sprintf("$%d", bindIndex))
		binds = append(binds, v.Field(i).Interface())
	}

	sql := fmt.Sprintf("INSERT INTO `%s` (%s) VALUES (%s)", table, strings.Join(columns, ", "), strings.Join(placeholders, ", "))

	return sql, binds
}

func batchPGInsertWithMap(table string, data []X, count int) (string, []interface{}) {
	fieldNum := len(data[0])

	fields := make([]string, 0, fieldNum)
	columns := make([]string, 0, fieldNum)
	placeholders := make([]string, 0, fieldNum)
	binds := make([]interface{}, 0, fieldNum*count)

	for k := range data[0] {
		fields = append(fields, k)
		columns = append(columns, fmt.Sprintf("`%s`", k))
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

	sql := fmt.Sprintf("INSERT INTO `%s` (%s) VALUES %s", table, strings.Join(columns, ", "), strings.Join(placeholders, ", "))

	return sql, binds
}

func batchPGInsertWithStruct(table string, v reflect.Value, count int) (string, []interface{}) {
	first := v.Index(0)

	if first.Kind() != reflect.Struct {
		panic("the data must be a slice to struct")
	}

	fieldNum := first.NumField()

	columns := make([]string, 0, fieldNum)
	placeholders := make([]string, 0, fieldNum)
	binds := make([]interface{}, 0, fieldNum*count)

	t := first.Type()
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

				columns = append(columns, fmt.Sprintf("`%s`", column))
			}

			phrs = append(phrs, fmt.Sprintf("$%d", bindIndex))
			binds = append(binds, v.Index(i).Field(j).Interface())
		}

		placeholders = append(placeholders, fmt.Sprintf("(%s)", strings.Join(phrs, ", ")))
	}

	sql := fmt.Sprintf("INSERT INTO `%s` (%s) VALUES %s", table, strings.Join(columns, ", "), strings.Join(placeholders, ", "))

	return sql, binds
}

func pgUpdateWithMap(query string, data X, args ...interface{}) (string, []interface{}) {
	var sets []string
	var binds []interface{}

	bindIndex := 0

	for k, v := range data {
		if e, ok := v.(*SQLExpr); ok {
			for i := 0; i < len(e.Args); i++ {
				bindIndex++
				e.Expr = strings.Replace(e.Expr, "?", fmt.Sprintf("$%d", bindIndex), 1)
			}

			sets = append(sets, fmt.Sprintf("`%s` = %s", k, e.Expr))
			binds = append(binds, e.Args...)
		} else {
			bindIndex++
			sets = append(sets, fmt.Sprintf("`%s` = $%d", k, bindIndex))
			binds = append(binds, v)
		}
	}

	sql := strings.Replace(query, "?", strings.Join(sets, ", "), 1)

	for i := 0; i < strings.Count(sql, "?"); i++ {
		bindIndex++
		sql = strings.Replace(sql, "?", fmt.Sprintf("$%d", bindIndex), 1)
	}

	binds = append(binds, args...)

	return sql, binds
}

func pgUpdateWithStruct(query string, v reflect.Value, args ...interface{}) (string, []interface{}) {
	fieldNum := v.NumField()

	sets := make([]string, 0, fieldNum)
	binds := make([]interface{}, 0, fieldNum+len(args))

	t := v.Type()
	bindIndex := 0

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
			for i := 0; i < len(e.Args); i++ {
				bindIndex++
				e.Expr = strings.Replace(e.Expr, "?", fmt.Sprintf("$%d", bindIndex), 1)
			}

			sets = append(sets, fmt.Sprintf("`%s` = %s", column, e.Expr))
			binds = append(binds, e.Args...)
		} else {
			bindIndex++
			sets = append(sets, fmt.Sprintf("`%s` = $%d", column, bindIndex))
			binds = append(binds, field)
		}
	}

	sql := strings.Replace(query, "?", strings.Join(sets, ", "), 1)

	for i := 0; i < strings.Count(sql, "?"); i++ {
		bindIndex++
		sql = strings.Replace(sql, "?", fmt.Sprintf("$%d", bindIndex), 1)
	}

	binds = append(binds, args...)

	return sql, binds
}
