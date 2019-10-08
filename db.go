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
)

// DBDriver indicates the db drivers.
type DBDriver int

const (
	MySQL    DBDriver = 1
	Postgres DBDriver = 2
)

// dbOptions db options
type dbOptions struct {
	maxOpenConns    int
	maxIdleConns    int
	connMaxLifetime time.Duration
	debug           bool
}

// DBOption configures how we set up the db
type DBOption interface {
	apply(*dbOptions)
}

// funcDBOption implements db option
type funcDBOption struct {
	f func(*dbOptions)
}

func (fo *funcDBOption) apply(o *dbOptions) {
	fo.f(o)
}

func newFuncDBOption(f func(*dbOptions)) *funcDBOption {
	return &funcDBOption{f: f}
}

// WithDBMaxOpenConns specifies the `MaxOpenConns` to db.
// MaxOpenConns sets the maximum number of open connections to the database.
//
// If MaxIdleConns is greater than 0 and the new MaxOpenConns is less than
// MaxIdleConns, then MaxIdleConns will be reduced to match the new
// MaxOpenConns limit.
//
// If n <= 0, then there is no limit on the number of open connections.
// The default is 0 (unlimited).
func WithDBMaxOpenConns(n int) DBOption {
	return newFuncDBOption(func(o *dbOptions) {
		o.maxOpenConns = n
	})
}

// WithDBMaxIdleConns specifies the `MaxIdleConns` to db.
// MaxIdleConns sets the maximum number of connections in the idle
// connection pool.
//
// If MaxOpenConns is greater than 0 but less than the new MaxIdleConns,
// then the new MaxIdleConns will be reduced to match the MaxOpenConns limit.
//
// If n <= 0, no idle connections are retained.
//
// The default max idle connections is currently 2. This may change in
// a future release.
func WithDBMaxIdleConns(n int) DBOption {
	return newFuncDBOption(func(o *dbOptions) {
		o.maxIdleConns = n
	})
}

// WithDBConnMaxLifetime specifies the `ConnMaxLifetime` to db.
// ConnMaxLifetime sets the maximum amount of time a connection may be reused.
//
// Expired connections may be closed lazily before reuse.
//
// If d <= 0, connections are reused forever.
func WithDBConnMaxLifetime(d time.Duration) DBOption {
	return newFuncDBOption(func(o *dbOptions) {
		o.connMaxLifetime = d
	})
}

// WithDBDebug specifies the `LogMod` to orm.
func WithDBDebug(b bool) DBOption {
	return newFuncDBOption(func(o *dbOptions) {
		o.debug = b
	})
}

var (
	// DB default db connection
	DB    *sqlx.DB
	dbmap sync.Map
	Orm   *gorm.DB
	ormap sync.Map
)

func dbDial(driverName, dsn string, options ...DBOption) (*gorm.DB, error) {
	o := &dbOptions{
		maxOpenConns:    20,
		maxIdleConns:    10,
		connMaxLifetime: 60 * time.Second,
	}

	if len(options) > 0 {
		for _, option := range options {
			option.apply(o)
		}
	}

	orm, err := gorm.Open(driverName, dsn)

	if err != nil {
		return nil, err
	}

	if o.debug {
		orm.LogMode(true)
	}

	orm.DB().SetMaxOpenConns(o.maxOpenConns)
	orm.DB().SetMaxIdleConns(o.maxIdleConns)
	orm.DB().SetConnMaxLifetime(o.connMaxLifetime)

	return orm, nil
}

// RegisterDB register a db, the param `dsn` eg:
//
// MySQL: `username:password@tcp(localhost:3306)/dbname?timeout=10s&charset=utf8mb4&collation=utf8mb4_general_ci&parseTime=True&loc=Local`;
//
// Postgres: `host=localhost port=5432 user=root password=secret dbname=test connect_timeout=10 sslmode=disable`.
//
// The default `MaxOpenConns` is 20;
// The default `MaxIdleConns` is 10;
// The default `ConnMaxLifetime` is 60s.
func RegisterDB(name string, driver DBDriver, dsn string, options ...DBOption) error {
	driverName := ""

	switch driver {
	case MySQL:
		driverName = "mysql"
	case Postgres:
		driverName = "postgres"
	default:
		return errors.New("yiigo: invalid db driver")
	}

	orm, err := dbDial(driverName, dsn, options...)

	if err != nil {
		return err
	}

	db := sqlx.NewDb(orm.DB(), driverName)

	dbmap.Store(name, db)
	ormap.Store(name, orm)

	if name == AsDefault {
		DB = db
		Orm = orm
	}

	return nil
}

// UseDB returns a db.
func UseDB(name ...string) *sqlx.DB {
	k := AsDefault

	if len(name) != 0 {
		k = name[0]
	}

	v, ok := dbmap.Load(k)

	if !ok {
		panic(fmt.Errorf("yiigo: db.%s is not registered", name))
	}

	return v.(*sqlx.DB)
}

// UseOrm returns an orm.
func UseOrm(name ...string) *gorm.DB {
	k := AsDefault

	if len(name) != 0 {
		k = name[0]
	}

	v, ok := ormap.Load(k)

	if !ok {
		panic(fmt.Errorf("yiigo: db.%s is not registered", name))
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
	sql := ""
	binds := make([]interface{}, 0)

	v := reflect.Indirect(reflect.ValueOf(data))

	switch v.Kind() {
	case reflect.Map:
		if x, ok := data.(X); ok {
			sql, binds = singleInsertWithMap(MySQL, table, x)
		}
	case reflect.Struct:
		sql, binds = singleInsertWithStruct(MySQL, table, v)
	case reflect.Slice:
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

			sql, binds = batchInsertWithMap(MySQL, table, x, count)
		case reflect.Struct:
			sql, binds = batchInsertWithStruct(MySQL, table, v, count)
		case reflect.Ptr:
			if e.Elem().Kind() != reflect.Struct {
				panic(errInsertInvalidType)
			}

			sql, binds = batchInsertWithStruct(MySQL, table, v, count)
		default:
			panic(errInsertInvalidType)
		}
	default:
		panic(errInsertInvalidType)
	}

	return sql, binds
}

// UpdateSQL returns mysql update sql and binds.
// param query expects eg: "UPDATE `table` SET ? WHERE `id` = ?".
// param data expects: `struct`, `*struct`, `yiigo.X`.
func UpdateSQL(query string, data interface{}, args ...interface{}) (string, []interface{}) {
	sql := ""
	binds := make([]interface{}, 0)

	v := reflect.Indirect(reflect.ValueOf(data))

	switch v.Kind() {
	case reflect.Map:
		x, ok := data.(X)

		if !ok {
			panic(errUpdateInvalidType)
		}

		sql, binds = updateWithMap(MySQL, query, x, args...)
	case reflect.Struct:
		sql, binds = updateWithStruct(MySQL, query, v, args...)
	default:
		panic(errUpdateInvalidType)
	}

	return sql, binds
}

// PGInsertSQL returns postgres insert sql and binds.
// param data expects: `struct`, `*struct`, `[]struct`, `[]*struct`, `yiigo.X`, `[]yiigo.X`.
func PGInsertSQL(table string, data interface{}) (string, []interface{}) {
	sql := ""
	binds := make([]interface{}, 0)

	v := reflect.Indirect(reflect.ValueOf(data))

	switch v.Kind() {
	case reflect.Map:
		if x, ok := data.(X); ok {
			sql, binds = singleInsertWithMap(Postgres, table, x)
		}
	case reflect.Struct:
		sql, binds = singleInsertWithStruct(Postgres, table, v)
	case reflect.Slice:
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

			sql, binds = batchInsertWithMap(Postgres, table, x, count)
		case reflect.Struct:
			sql, binds = batchInsertWithStruct(Postgres, table, v, count)
		case reflect.Ptr:
			if e.Elem().Kind() != reflect.Struct {
				panic(errInsertInvalidType)
			}

			sql, binds = batchInsertWithStruct(Postgres, table, v, count)
		default:
			panic(errInsertInvalidType)
		}
	default:
		panic(errInsertInvalidType)
	}

	return sql, binds
}

// PGUpdateSQL returns postgres update sql and binds.
// param query expects eg: "UPDATE `table` SET $1 WHERE `id` = $2".
// param data expects: `struct`, `*struct`, `yiigo.X`.
func PGUpdateSQL(query string, data interface{}, args ...interface{}) (string, []interface{}) {
	sql := ""
	binds := make([]interface{}, 0)

	v := reflect.Indirect(reflect.ValueOf(data))

	switch v.Kind() {
	case reflect.Map:
		x, ok := data.(X)

		if !ok {
			panic(errUpdateInvalidType)
		}

		sql, binds = updateWithMap(Postgres, query, x, args...)
	case reflect.Struct:
		sql, binds = updateWithStruct(Postgres, query, v, args...)
	}

	return sql, binds
}

func singleInsertWithMap(driver DBDriver, table string, data X) (string, []interface{}) {
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
	}

	return sql, binds
}

func singleInsertWithStruct(driver DBDriver, table string, v reflect.Value) (string, []interface{}) {
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
	}

	return sql, binds
}

func batchInsertWithMap(driver DBDriver, table string, data []X, count int) (string, []interface{}) {
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

func batchInsertWithStruct(driver DBDriver, table string, v reflect.Value, count int) (string, []interface{}) {
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

func updateWithMap(driver DBDriver, query string, data X, args ...interface{}) (string, []interface{}) {
	dataLen := len(data)
	argsLen := len(args)

	sql := ""
	sets := make([]string, 0, dataLen)
	binds := make([]interface{}, 0, dataLen+argsLen)

	switch driver {
	case MySQL:
		for k, v := range data {
			sets = append(sets, fmt.Sprintf("`%s` = ?", k))
			binds = append(binds, v)
		}

		sql = strings.Replace(query, "?", strings.Join(sets, ", "), 1)
		binds = append(binds, args...)
	case Postgres:
		bindIndex := 0

		for k, v := range data {
			bindIndex++

			sets = append(sets, fmt.Sprintf(`"%s" = $%d`, k, bindIndex))
			binds = append(binds, v)
		}

		oldnew := make([]string, 0, argsLen*2)

		for i := 1; i <= argsLen; i++ {
			oldnew = append(oldnew, fmt.Sprintf("$%d", i+1), fmt.Sprintf("$%d", dataLen+i))
		}

		r := strings.NewReplacer(oldnew...)
		query = r.Replace(query)

		sql = strings.Replace(query, "$1", strings.Join(sets, ", "), 1)
		binds = append(binds, args...)
	}

	return sql, binds
}

func updateWithStruct(driver DBDriver, query string, v reflect.Value, args ...interface{}) (string, []interface{}) {
	fieldNum := v.NumField()
	argsLen := len(args)

	sql := ""
	sets := make([]string, 0, fieldNum)
	binds := make([]interface{}, 0, fieldNum+argsLen)

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
		binds = append(binds, args...)
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

		oldnew := make([]string, 0, argsLen*2)

		for i := 1; i <= argsLen; i++ {
			oldnew = append(oldnew, fmt.Sprintf("$%d", i+1), fmt.Sprintf("$%d", bindIndex+i))
		}

		r := strings.NewReplacer(oldnew...)
		query = r.Replace(query)

		sql = strings.Replace(query, "$1", strings.Join(sets, ", "), 1)
		binds = append(binds, args...)
	}

	return sql, binds
}
