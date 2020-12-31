package yiigo

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

// SQLBuilder is the interface that wrap query options
type SQLBuilder interface {
	Wrap(options ...QueryOption) SQLWrapper
}

// SQLWrapper is the interface that build sql statement
type SQLWrapper interface {
	ToQuery() (string, []interface{})
	ToInsert(data interface{}) (string, []interface{})
	ToBatchInsert(data interface{}) (string, []interface{})
	ToUpdate(data interface{}) (string, []interface{})
	ToDelete() (string, []interface{})
}

// QueryBuilder is a SQLBuilder implementation
type QueryBuilder struct {
	driver DBDriver
}

// Wrap wrap query options
func (b *QueryBuilder) Wrap(options ...QueryOption) SQLWrapper {
	wrapper := &QueryWrapper{
		driver:  b.driver,
		columns: []string{"*"},
	}

	for _, f := range options {
		f(wrapper)
	}

	return wrapper
}

// NewSQLBuilder returns new SQLBuilder
func NewSQLBuilder(driver DBDriver) SQLBuilder {
	return &QueryBuilder{driver: driver}
}

// SQLClause SQL clause
type SQLClause struct {
	query string
	binds []interface{}
}

// Clause returns sql clause, eg: yiigo.Clause("price * ? + ?", 2, 100).
func Clause(query string, binds ...interface{}) *SQLClause {
	return &SQLClause{
		query: query,
		binds: binds,
	}
}

// QueryWrapper is a SQLWrapper implementation
type QueryWrapper struct {
	driver  DBDriver
	table   string
	columns []string
	where   *SQLClause
	joins   []string
	group   []string
	having  *SQLClause
	order   []string
	offset  *SQLClause
	limit   *SQLClause
	unions  []*SQLClause
	distinct bool
	whereIn  bool
}

// ToQuery returns query statement and binds.
func (w *QueryWrapper) ToQuery() (string, []interface{}) {
	clause := w.subquery()

	// unions
	if l := len(w.unions); l != 0 {
		statements := make([]string, 0, l+1)

		statements = append(statements, fmt.Sprintf("(%s)", clause.query))

		for _, v := range w.unions {
			statements = append(statements, v.query)
			clause.binds = append(clause.binds, v.binds...)
		}

		clause.query = strings.Join(statements, " ")
	}

	// where in
	if w.whereIn {
		var err error

		clause.query, clause.binds, err = sqlx.In(clause.query, clause.binds...)

		if err != nil {
			logger.Error("yiigo: build 'IN' query error", zap.Error(err))

			return "", nil
		}
	}

	clause.query = sqlx.Rebind(sqlx.BindType(string(w.driver)), clause.query)

	if debug {
		logger.Info(clause.query, zap.Any("binds", clause.binds))
	}

	return clause.query, clause.binds
}

func (w *QueryWrapper) subquery() *SQLClause {
	statements := make([]string, 0)
	binds := make([]interface{}, 0)

	statements = append(statements, "SELECT")

	if w.distinct {
		statements = append(statements, "DISTINCT")
	}

	statements = append(statements, strings.Join(w.columns, ", "), "FROM", w.table)

	if len(w.joins) != 0 {
		statements = append(statements, w.joins...)
	}

	if w.where != nil {
		statements = append(statements, "WHERE", w.where.query)
		binds = append(binds, w.where.binds...)
	}

	if len(w.group) != 0 {
		statements = append(statements, "GROUP BY", strings.Join(w.group, ", "))
	}

	if w.having != nil {
		statements = append(statements, "HAVING", w.having.query)
		binds = append(binds, w.having.binds...)
	}

	if len(w.order) != 0 {
		statements = append(statements, "ORDER BY", strings.Join(w.order, ", "))
	}

	if w.offset != nil {
		statements = append(statements, w.offset.query)
		binds = append(binds, w.offset.binds...)
	}

	if w.limit != nil {
		statements = append(statements, w.limit.query)
		binds = append(binds, w.limit.binds...)
	}

	return Clause(strings.Join(statements, " "), binds...)
}

// ToInsert returns insert statement and binds.
// data expects `struct`, `*struct`, `yiigo.X`.
func (w *QueryWrapper) ToInsert(data interface{}) (string, []interface{}) {
	var clause *SQLClause

	v := reflect.Indirect(reflect.ValueOf(data))

	switch v.Kind() {
	case reflect.Map:
		x, ok := data.(X)

		if !ok {
			logger.Error("yiigo: invalid data type for insert, expects struct, *struct, yiigo.X")

			return "", nil
		}

		clause = w.insertWithMap(x)
	case reflect.Struct:
		clause = w.insertWithStruct(v)
	default:
		logger.Error("yiigo: invalid data type for insert, expects struct, *struct, yiigo.X")

		return "", nil
	}

	statements := make([]string, 0, 8)

	statements = append(statements, "INSERT INTO", w.table, clause.query)

	if w.driver == Postgres {
		statements = append(statements, "RETURNING id")
	}

	query := sqlx.Rebind(sqlx.BindType(string(w.driver)), strings.Join(statements, " "))

	if debug {
		logger.Info(query, zap.Any("binds", clause.binds))
	}

	return query, clause.binds
}

func (w *QueryWrapper) insertWithMap(data X) *SQLClause {
	fieldNum := len(data)

	columns := make([]string, 0, fieldNum)
	values := make([]string, 0, fieldNum)
	binds := make([]interface{}, 0, fieldNum)

	for k, v := range data {
		columns = append(columns, k)
		values = append(values, "?")
		binds = append(binds, v)
	}

	return Clause(strings.Join([]string{fmt.Sprintf("(%s)", strings.Join(columns, ", ")), "VALUES", fmt.Sprintf("(%s)", strings.Join(values, ", "))}, " "), binds...)
}

func (w *QueryWrapper) insertWithStruct(v reflect.Value) *SQLClause {
	fieldNum := v.NumField()

	columns := make([]string, 0, fieldNum)
	values := make([]string, 0, fieldNum)
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

		columns = append(columns, column)
		values = append(values, "?")
		binds = append(binds, v.Field(i).Interface())
	}

	return Clause(strings.Join([]string{fmt.Sprintf("(%s)", strings.Join(columns, ", ")), "VALUES", fmt.Sprintf("(%s)", strings.Join(values, ", "))}, " "), binds...)
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

	var clause *SQLClause

	e := v.Type().Elem()

	switch e.Kind() {
	case reflect.Map:
		x, ok := data.([]X)

		if !ok {
			logger.Error("yiigo: invalid data type for batch insert, expects []struct, []*struct, []yiigo.X")

			return "", nil
		}

		clause = w.batchInsertWithMap(x)
	case reflect.Struct:
		clause = w.batchInsertWithStruct(v)
	case reflect.Ptr:
		if e.Elem().Kind() != reflect.Struct {
			logger.Error("yiigo: invalid data type for batch insert, expects []struct, []*struct, []yiigo.X")

			return "", nil
		}

		clause = w.batchInsertWithStruct(v)
	default:
		logger.Error("yiigo: invalid data type for batch insert, expects []struct, []*struct, []yiigo.X")

		return "", nil
	}

	query := sqlx.Rebind(sqlx.BindType(string(w.driver)), clause.query)

	if debug {
		logger.Info(query, zap.Any("binds", clause.binds))
	}

	return query, clause.binds
}

func (w *QueryWrapper) batchInsertWithMap(data []X) *SQLClause {
	dataLen := len(data)
	fieldNum := len(data[0])

	columns := make([]string, 0, fieldNum)
	values := make([]string, 0, dataLen)
	binds := make([]interface{}, 0, fieldNum*dataLen)

	for k := range data[0] {
		columns = append(columns, k)
	}

	for _, x := range data {
		phrs := make([]string, 0, fieldNum)

		for _, v := range columns {
			phrs = append(phrs, "?")
			binds = append(binds, x[v])
		}

		values = append(values, fmt.Sprintf("(%s)", strings.Join(phrs, ", ")))
	}

	return Clause(strings.Join([]string{"INSERT INTO", w.table, fmt.Sprintf("(%s)", strings.Join(columns, ", ")), "VALUES", strings.Join(values, ", ")}, " "), binds...)
}

func (w *QueryWrapper) batchInsertWithStruct(v reflect.Value) *SQLClause {
	first := reflect.Indirect(v.Index(0))

	dataLen := v.Len()
	fieldNum := first.NumField()

	columns := make([]string, 0, fieldNum)
	values := make([]string, 0, dataLen)
	binds := make([]interface{}, 0, fieldNum*dataLen)

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

				columns = append(columns, column)
			}

			phrs = append(phrs, "?")
			binds = append(binds, reflect.Indirect(v.Index(i)).Field(j).Interface())
		}

		values = append(values, fmt.Sprintf("(%s)", strings.Join(phrs, ", ")))
	}

	return Clause(strings.Join([]string{"INSERT INTO", w.table, fmt.Sprintf("(%s)", strings.Join(columns, ", ")), "VALUES", strings.Join(values, ", ")}, " "), binds...)
}

// ToUpdate returns update statement and binds.
// data expects `struct`, `*struct`, `yiigo.X`.
func (w *QueryWrapper) ToUpdate(data interface{}) (string, []interface{}) {
	var clause *SQLClause

	v := reflect.Indirect(reflect.ValueOf(data))

	switch v.Kind() {
	case reflect.Map:
		x, ok := data.(X)

		if !ok {
			logger.Error("yiigo: invalid data type for update, expects struct, *struct, yiigo.X")

			return "", nil
		}

		clause = w.updateWithMap(x)
	case reflect.Struct:
		clause = w.updateWithStruct(v)
	default:
		logger.Error("yiigo: invalid data type for update, expects struct, *struct, yiigo.X")

		return "", nil
	}

	statements := make([]string, 0, 6)

	statements = append(statements, "UPDATE", w.table, "SET", clause.query)

	if w.where != nil {
		statements = append(statements, "WHERE", w.where.query)
		clause.binds = append(clause.binds, w.where.binds...)
	}

	clause.query = strings.Join(statements, " ")

	if w.whereIn {
		var err error

		clause.query, clause.binds, err = sqlx.In(clause.query, clause.binds...)

		if err != nil {
			logger.Error("yiigo: build 'IN' query error", zap.Error(err))

			return "", nil
		}
	}

	clause.query = sqlx.Rebind(sqlx.BindType(string(w.driver)), clause.query)

	if debug {
		logger.Info(clause.query, zap.Any("binds", clause.binds))
	}

	return clause.query, clause.binds
}

func (w *QueryWrapper) updateWithMap(data X) *SQLClause {
	fieldNum := len(data)

	sets := make([]string, 0, fieldNum)
	binds := make([]interface{}, 0, fieldNum)

	for k, v := range data {
		if clause, ok := v.(*SQLClause); ok {
			sets = append(sets, strings.Join([]string{k, "=", clause.query}, " "))
			binds = append(binds, clause.binds...)

			continue
		}

		sets = append(sets, strings.Join([]string{k, "=", "?"}, " "))
		binds = append(binds, v)
	}

	return Clause(strings.Join(sets, ", "), binds...)
}

func (w *QueryWrapper) updateWithStruct(v reflect.Value) *SQLClause {
	fieldNum := v.NumField()

	sets := make([]string, 0, fieldNum)
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

		sets = append(sets, strings.Join([]string{column, "=", "?"}, " "))
		binds = append(binds, v.Field(i).Interface())
	}

	return Clause(strings.Join(sets, ", "), binds...)
}

// ToDelete returns delete clause and binds.
func (w *QueryWrapper) ToDelete() (string, []interface{}) {
	statements := make([]string, 0, 5)
	binds := make([]interface{}, 0)

	statements = append(statements, "DELETE", "FROM", w.table)

	if w.where != nil {
		statements = append(statements, "WHERE", w.where.query)
		binds = append(binds, w.where.binds...)
	}

	query := strings.Join(statements, " ")

	if w.whereIn {
		var err error

		query, binds, err = sqlx.In(query, binds...)

		if err != nil {
			logger.Error("yiigo: build 'IN' query error", zap.Error(err))

			return "", nil
		}
	}

	query = sqlx.Rebind(sqlx.BindType(string(w.driver)), query)

	if debug {
		logger.Info(query, zap.Any("binds", binds))
	}

	return query, binds
}

// QueryOption configures how we set up the SQL query statement
type QueryOption func(w *QueryWrapper)

// Table specifies the query table.
func Table(name string) QueryOption {
	return func(w *QueryWrapper) {
		w.table = name
	}
}

// Select specifies the query columns.
func Select(columns ...string) QueryOption {
	return func(w *QueryWrapper) {
		w.columns = columns
	}
}

// Distinct specifies the `distinct` clause.
func Distinct(columns ...string) QueryOption {
	return func(w *QueryWrapper) {
		w.columns = columns
		w.distinct = true
	}
}

// Join specifies the `inner join` clause.
func Join(table, on string) QueryOption {
	return func(w *QueryWrapper) {
		w.joins = append(w.joins, strings.Join([]string{"INNER", "JOIN", table, "ON", on}, " "))
	}
}

// LeftJoin specifies the `left join` clause.
func LeftJoin(table, on string) QueryOption {
	return func(w *QueryWrapper) {
		w.joins = append(w.joins, strings.Join([]string{"LEFT", "JOIN", table, "ON", on}, " "))
	}
}

// RightJoin specifies the `right join` clause.
func RightJoin(table, on string) QueryOption {
	return func(w *QueryWrapper) {
		w.joins = append(w.joins, strings.Join([]string{"RIGHT", "JOIN", table, "ON", on}, " "))
	}
}

// FullJoin specifies the `full join` clause.
func FullJoin(table, on string) QueryOption {
	return func(w *QueryWrapper) {
		w.joins = append(w.joins, strings.Join([]string{"FULL", "JOIN", table, "ON", on}, " "))
	}
}

// Where specifies the `where` clause.
func Where(query string, binds ...interface{}) QueryOption {
	return func(w *QueryWrapper) {
		w.where = Clause(query, binds...)
	}
}

// WhereIn specifies the `where in` clause.
func WhereIn(query string, binds ...interface{}) QueryOption {
	return func(w *QueryWrapper) {
		w.where = Clause(query, binds...)
		w.whereIn = true
	}
}

// GroupBy specifies the `group by` clause.
func GroupBy(columns ...string) QueryOption {
	return func(w *QueryWrapper) {
		w.group = columns
	}
}

// Having specifies the `having` clause.
func Having(query string, binds ...interface{}) QueryOption {
	return func(w *QueryWrapper) {
		w.having = Clause(query, binds...)
	}
}

// OrderBy specifies the `order by` clause.
func OrderBy(columns ...string) QueryOption {
	return func(w *QueryWrapper) {
		w.order = columns
	}
}

// Offset specifies the `offset` clause.
func Offset(n int) QueryOption {
	return func(w *QueryWrapper) {
		w.offset = Clause("OFFSET ?", n)
	}
}

// Limit specifies the `limit` clause.
func Limit(n int) QueryOption {
	return func(w *QueryWrapper) {
		w.limit = Clause("LIMIT ?", n)
	}
}

// Union specifies the `union` clause.
func Union(wrappers ...SQLWrapper) QueryOption {
	return func(w *QueryWrapper) {
		for _, wrapper := range wrappers {
			v, ok := wrapper.(*QueryWrapper)

			if !ok {
				continue
			}

			if v.whereIn {
				w.whereIn = true
			}

			clause := v.subquery()

			w.unions = append(w.unions, Clause(strings.Join([]string{"UNION", fmt.Sprintf("(%s)", v.subquery().query)}, " "), clause.binds...))
		}
	}
}

// UnionAll specifies the `union all` clause.
func UnionAll(wrappers ...SQLWrapper) QueryOption {
	return func(w *QueryWrapper) {
		for _, wrapper := range wrappers {
			v, ok := wrapper.(*QueryWrapper)

			if !ok {
				continue
			}

			if v.whereIn {
				w.whereIn = true
			}

			clause := v.subquery()

			w.unions = append(w.unions, Clause(strings.Join([]string{"UNION", "ALL", fmt.Sprintf("(%s)", clause.query)}, " "), clause.binds...))
		}
	}
}
