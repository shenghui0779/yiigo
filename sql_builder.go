package yiigo

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

// SQLBuilder build SQL statement
type SQLBuilder struct {
	driver DBDriver
}

// Wrap wrap query clauses
func (b *SQLBuilder) Wrap(options ...QueryOption) *QueryWrapper {
	wrapper := &QueryWrapper{driver: b.driver}

	for _, option := range options {
		option.apply(wrapper)
	}

	return wrapper
}

// NewSQLBuilder returns new SQL builder
func NewSQLBuilder(driver DBDriver) *SQLBuilder {
	return &SQLBuilder{driver: driver}
}

// SQLClause SQL clause
type SQLClause struct {
	query string
	args  []interface{}
}

// Clause returns sql clause, eg: yiigo.Clause("price * ? + ?", 2, 100).
func Clause(query string, args ...interface{}) *SQLClause {
	return &SQLClause{
		query: query,
		args:  args,
	}
}

// QueryWrapper query clauses wrapper
type QueryWrapper struct {
	driver   DBDriver
	table    string
	columns  []string
	distinct []string
	where    *SQLClause
	joins    []string
	group    string
	having   *SQLClause
	order    string
	offset   int
	limit    int

	queryLen int
	bindsLen int

	values []string
	sets   []string
	binds  []interface{}
}

// ToToQuery returns query statement and binds.
func (w *QueryWrapper) ToQuery() (string, []interface{}) {
	clauses := make([]string, 0, w.queryLen+2)
	w.binds = make([]interface{}, 0, w.bindsLen)

	clauses = append(clauses, "SELECT")

	if len(w.distinct) != 0 {
		clauses = append(clauses, "DISTINCT", strings.Join(w.distinct, ", "))
	} else if len(w.columns) != 0 {
		clauses = append(clauses, strings.Join(w.columns, ", "))
	} else {
		clauses = append(clauses, "*")
	}

	clauses = append(clauses, "FROM", w.table)

	if len(w.joins) != 0 {
		clauses = append(clauses, w.joins...)
	}

	if w.where != nil {
		clauses = append(clauses, "WHERE", w.where.query)
		w.binds = append(w.binds, w.where.args...)
	}

	if w.group != "" {
		clauses = append(clauses, "GROUP BY", w.group)
	}

	if w.having != nil {
		clauses = append(clauses, "HAVING", w.having.query)
		w.binds = append(w.binds, w.having.args...)
	}

	if w.order != "" {
		clauses = append(clauses, "ORDER BY", w.order)
	}

	if w.offset != 0 {
		clauses = append(clauses, "OFFSET", strconv.Itoa(w.offset))
	}

	if w.limit != 0 {
		clauses = append(clauses, "LIMIT", strconv.Itoa(w.limit))
	}

	query, binds, err := sqlx.In(strings.Join(clauses, " "), w.binds...)

	if err != nil {
		logger.Error("yiigo: build 'IN' query error", zap.Error(err))

		return "", nil
	}

	query = sqlx.Rebind(sqlx.BindType(string(w.driver)), query)

	if debug {
		logger.Info(query, zap.Any("binds", binds))
	}

	return query, binds
}

// ToInsert returns insert statement and binds.
// data expects `struct`, `*struct`, `yiigo.X`.
func (w *QueryWrapper) ToInsert(data interface{}) (string, []interface{}) {
	v := reflect.Indirect(reflect.ValueOf(data))

	switch v.Kind() {
	case reflect.Map:
		x, ok := data.(X)

		if !ok {
			logger.Error("yiigo: invalid data type for insert, expects struct, *struct, yiigo.X")

			return "", nil
		}

		w.insertWithMap(x)
	case reflect.Struct:
		w.insertWithStruct(v)
	default:
		logger.Error("yiigo: invalid data type for insert, expects struct, *struct, yiigo.X")

		return "", nil
	}

	clauses := make([]string, 0, 12)

	clauses = append(clauses, "INSERT", "INTO", w.table, fmt.Sprintf("(%s)", strings.Join(w.columns, ", ")), "VALUES", fmt.Sprintf("(%s)", strings.Join(w.values, ", ")))

	if w.driver == Postgres {
		clauses = append(clauses, "RETURNING", "id")
	}

	query := sqlx.Rebind(sqlx.BindType(string(w.driver)), strings.Join(clauses, " "))

	if debug {
		logger.Info(query, zap.Any("binds", w.binds))
	}

	return query, w.binds
}

func (w *QueryWrapper) insertWithMap(data X) {
	fieldNum := len(data)

	w.columns = make([]string, 0, fieldNum)
	w.values = make([]string, 0, fieldNum)
	w.binds = make([]interface{}, 0, fieldNum)

	for k, v := range data {
		w.columns = append(w.columns, k)
		w.values = append(w.values, "?")
		w.binds = append(w.binds, v)
	}
}

func (w *QueryWrapper) insertWithStruct(v reflect.Value) {
	fieldNum := v.NumField()

	w.columns = make([]string, 0, fieldNum)
	w.values = make([]string, 0, fieldNum)
	w.binds = make([]interface{}, 0, fieldNum)

	t := v.Type()

	for i := 0; i < fieldNum; i++ {
		column := t.Field(i).Tag.Get("db")

		if column == "-" {
			continue
		}

		if column == "" {
			column = t.Field(i).Name
		}

		w.columns = append(w.columns, column)
		w.values = append(w.values, "?")
		w.binds = append(w.binds, v.Field(i).Interface())
	}
}

// ToBatchInsert returns batch insert statement and binds.
// data expects `[]struct`, `[]*struct`, `[]yiigo.X`.
func (w *QueryWrapper) ToBatchInsert(data interface{}) (string, []interface{}) {
	v := reflect.Indirect(reflect.ValueOf(data))

	if v.Kind() != reflect.Slice {
		logger.Error("yiigo: invalid data type for batch insert, expects []struct, []*struct, []yiigo.X")

		return "", nil
	}

	if v.Len() == 0 {
		logger.Error("yiigo: empty data for batch insert")

		return "", nil
	}

	e := v.Type().Elem()

	switch e.Kind() {
	case reflect.Map:
		x, ok := data.([]X)

		if !ok {
			logger.Error("yiigo: invalid data type for batch insert, expects []struct, []*struct, []yiigo.X")

			return "", nil
		}

		w.batchInsertWithMap(x)
	case reflect.Struct:
		w.batchInsertWithStruct(v)
	case reflect.Ptr:
		if e.Elem().Kind() != reflect.Struct {
			logger.Error("yiigo: invalid data type for batch insert, expects []struct, []*struct, []yiigo.X")

			return "", nil
		}

		w.batchInsertWithStruct(v)
	default:
		logger.Error("yiigo: invalid data type for batch insert, expects []struct, []*struct, []yiigo.X")

		return "", nil
	}

	clauses := []string{"INSERT", "INTO", w.table, fmt.Sprintf("(%s)", strings.Join(w.columns, ", ")), "VALUES", strings.Join(w.values, ", ")}

	query := sqlx.Rebind(sqlx.BindType(string(w.driver)), strings.Join(clauses, " "))

	if debug {
		logger.Info(query, zap.Any("binds", w.binds))
	}

	return query, w.binds
}

func (w *QueryWrapper) batchInsertWithMap(data []X) {
	dataLen := len(data)
	fieldNum := len(data[0])

	w.columns = make([]string, 0, fieldNum)
	w.values = make([]string, 0, dataLen)
	w.binds = make([]interface{}, 0, fieldNum*dataLen)

	for k := range data[0] {
		w.columns = append(w.columns, k)
	}

	for _, x := range data {
		phrs := make([]string, 0, fieldNum)

		for _, v := range w.columns {
			phrs = append(phrs, "?")
			w.binds = append(w.binds, x[v])
		}

		w.values = append(w.values, fmt.Sprintf("(%s)", strings.Join(phrs, ", ")))
	}
}

func (w *QueryWrapper) batchInsertWithStruct(v reflect.Value) {
	first := reflect.Indirect(v.Index(0))

	dataLen := v.Len()
	fieldNum := first.NumField()

	w.columns = make([]string, 0, fieldNum)
	w.values = make([]string, 0, dataLen)
	w.binds = make([]interface{}, 0, fieldNum*dataLen)

	t := first.Type()

	for i := 0; i < dataLen; i++ {
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

				w.columns = append(w.columns, column)
			}

			phrs = append(phrs, "?")
			w.binds = append(w.binds, reflect.Indirect(v.Index(i)).Field(j).Interface())
		}

		w.values = append(w.values, fmt.Sprintf("(%s)", strings.Join(phrs, ", ")))
	}
}

// ToUpdate returns update statement and binds.
// data expects `struct`, `*struct`, `yiigo.X`.
func (w *QueryWrapper) ToUpdate(data interface{}) (string, []interface{}) {
	v := reflect.Indirect(reflect.ValueOf(data))

	switch v.Kind() {
	case reflect.Map:
		x, ok := data.(X)

		if !ok {
			logger.Error("yiigo: invalid data type for update, expects struct, *struct, yiigo.X")

			return "", nil
		}

		w.updateWithMap(x)
	case reflect.Struct:
		w.updateWithStruct(v)
	default:
		logger.Error("yiigo: invalid data type for update, expects struct, *struct, yiigo.X")

		return "", nil
	}

	clauses := make([]string, 0, 6)

	clauses = append(clauses, "UPDATE", w.table, "SET", strings.Join(w.sets, ", "))

	if w.where != nil {
		clauses = append(clauses, "WHERE", w.where.query)
		w.binds = append(w.binds, w.where.args...)
	}

	query, binds, err := sqlx.In(strings.Join(clauses, " "), w.binds...)

	if err != nil {
		logger.Error("yiigo: build 'IN' query error", zap.Error(err))

		return "", nil
	}

	query = sqlx.Rebind(sqlx.BindType(string(w.driver)), query)

	if debug {
		logger.Info(query, zap.Any("binds", binds))
	}

	return query, binds
}

func (w *QueryWrapper) updateWithMap(data X) {
	fieldNum := len(data)

	w.sets = make([]string, 0, fieldNum)
	w.binds = make([]interface{}, 0, fieldNum+w.bindsLen)

	for k, v := range data {
		if clause, ok := v.(*SQLClause); ok {
			w.sets = append(w.sets, fmt.Sprintf("%s = %s", k, clause.query))
			w.binds = append(w.binds, clause.args...)

			continue
		}

		w.sets = append(w.sets, fmt.Sprintf("%s = ?", k))
		w.binds = append(w.binds, v)
	}
}

func (w *QueryWrapper) updateWithStruct(v reflect.Value) {
	fieldNum := v.NumField()

	w.sets = make([]string, 0, fieldNum)
	w.binds = make([]interface{}, 0, fieldNum+w.bindsLen)

	t := v.Type()

	for i := 0; i < fieldNum; i++ {
		column := t.Field(i).Tag.Get("db")

		if column == "-" {
			continue
		}

		if column == "" {
			column = t.Field(i).Name
		}

		w.sets = append(w.sets, fmt.Sprintf("%s = ?", column))
		w.binds = append(w.binds, v.Field(i).Interface())
	}
}

// ToDelete returns delete clause and binds.
func (w *QueryWrapper) ToDelete() (string, []interface{}) {
	clauses := make([]string, 0, w.queryLen)
	w.binds = make([]interface{}, 0, w.bindsLen)

	clauses = append(clauses, "DELETE", "FROM", w.table)

	if w.where != nil {
		clauses = append(clauses, "WHERE", w.where.query)
		w.binds = append(w.binds, w.where.args...)
	}

	query, binds, err := sqlx.In(strings.Join(clauses, " "), w.binds...)

	if err != nil {
		logger.Error("yiigo: build 'IN' query error", zap.Error(err))

		return "", nil
	}

	query = sqlx.Rebind(sqlx.BindType(string(w.driver)), query)

	if debug {
		logger.Info(query, zap.Any("binds", binds))
	}

	return query, binds
}

// QueryOption configures how we set up the SQL query statement
type QueryOption interface {
	apply(*QueryWrapper)
}

// funcQueryOption implements query option
type funcQueryOption struct {
	f func(*QueryWrapper)
}

func (fo *funcQueryOption) apply(wrapper *QueryWrapper) {
	fo.f(wrapper)
}

func newFuncQueryOption(f func(*QueryWrapper)) *funcQueryOption {
	return &funcQueryOption{f: f}
}

// Table specifies the query table.
func Table(name string) QueryOption {
	return newFuncQueryOption(func(wrapper *QueryWrapper) {
		wrapper.table = name
		wrapper.queryLen += 2
	})
}

// Select specifies the query columns.
func Select(columns ...string) QueryOption {
	return newFuncQueryOption(func(wrapper *QueryWrapper) {
		wrapper.columns = columns
		wrapper.queryLen += 2
	})
}

// Distinct specifies the `distinct` clause.
func Distinct(columns ...string) QueryOption {
	return newFuncQueryOption(func(wrapper *QueryWrapper) {
		wrapper.distinct = columns
		wrapper.queryLen += 2
	})
}

// InnerJoin specifies the `inner join` clause.
func InnerJoin(table, on string) QueryOption {
	return newFuncQueryOption(func(wrapper *QueryWrapper) {
		wrapper.joins = append(wrapper.joins, "INNER", "JOIN", table, "ON", on)
		wrapper.queryLen += 5
	})
}

// LeftJoin specifies the `left join` clause.
func LeftJoin(table, on string) QueryOption {
	return newFuncQueryOption(func(wrapper *QueryWrapper) {
		wrapper.joins = append(wrapper.joins, "LEFT", "JOIN", table, "ON", on)
		wrapper.queryLen += 5
	})
}

// RightJoin specifies the `right join` clause.
func RightJoin(table, on string) QueryOption {
	return newFuncQueryOption(func(wrapper *QueryWrapper) {
		wrapper.joins = append(wrapper.joins, "RIGHT", "JOIN", table, "ON", on)
		wrapper.queryLen += 5
	})
}

// FullJoin specifies the `full join` clause.
func FullJoin(table, on string) QueryOption {
	return newFuncQueryOption(func(wrapper *QueryWrapper) {
		wrapper.joins = append(wrapper.joins, "FULL", "JOIN", table, "ON", on)
		wrapper.queryLen += 5
	})
}

// Where specifies the `where` clause.
func Where(query string, args ...interface{}) QueryOption {
	return newFuncQueryOption(func(wrapper *QueryWrapper) {
		wrapper.where = Clause(query, args...)

		wrapper.queryLen += 2
		wrapper.bindsLen += len(args)
	})
}

// GroupBy specifies the `group by` clause.
func GroupBy(column string) QueryOption {
	return newFuncQueryOption(func(wrapper *QueryWrapper) {
		wrapper.group = column
		wrapper.queryLen += 2
	})
}

// Having specifies the `having` clause.
func Having(query string, args ...interface{}) QueryOption {
	return newFuncQueryOption(func(wrapper *QueryWrapper) {
		wrapper.having = Clause(query, args...)

		wrapper.queryLen += 2
		wrapper.bindsLen += len(args)
	})
}

// OrderBy specifies the `order by` clause.
func OrderBy(query string) QueryOption {
	return newFuncQueryOption(func(wrapper *QueryWrapper) {
		wrapper.order = query
		wrapper.queryLen += 2
	})
}

// Offset specifies the `offset` clause.
func Offset(n int) QueryOption {
	return newFuncQueryOption(func(wrapper *QueryWrapper) {
		wrapper.offset = n
		wrapper.queryLen += 2
	})
}

// Limit specifies the `limit` clause.
func Limit(n int) QueryOption {
	return newFuncQueryOption(func(wrapper *QueryWrapper) {
		wrapper.limit = n
		wrapper.queryLen += 2
	})
}
