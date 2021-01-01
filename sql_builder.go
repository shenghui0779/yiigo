package yiigo

import (
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
	table   string
	keyword string
	query   string
	binds   []interface{}
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
	driver   DBDriver
	table    string
	columns  []string
	where    *SQLClause
	joins    []*SQLClause
	groups   []string
	having   *SQLClause
	orders   []string
	offset   int
	limit    int
	unions   []*SQLClause
	distinct bool
	whereIn  bool
}

// ToQuery returns query statement and binds.
func (w *QueryWrapper) ToQuery() (string, []interface{}) {
	query, binds := w.subquery()

	// unions
	if l := len(w.unions); l != 0 {
		var builder strings.Builder

		builder.WriteString("(")
		builder.WriteString(query)
		builder.WriteString(")")

		for _, v := range w.unions {
			builder.WriteString(" ")
			builder.WriteString(v.keyword)
			builder.WriteString(" (")
			builder.WriteString(v.query)
			builder.WriteString(")")

			binds = append(binds, v.binds...)
		}

		query = builder.String()
	}

	// where in
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

func (w *QueryWrapper) subquery() (string, []interface{}) {
	binds := make([]interface{}, 0)

	var builder strings.Builder

	builder.WriteString("SELECT ")

	if w.distinct {
		builder.WriteString("DISTINCT ")
	}

	builder.WriteString(w.columns[0])

	for _, column := range w.columns[1:] {
		builder.WriteString(", ")
		builder.WriteString(column)
	}

	builder.WriteString(" FROM ")
	builder.WriteString(w.table)

	if len(w.joins) != 0 {
		for _, join := range w.joins {
			builder.WriteString(" ")
			builder.WriteString(join.keyword)
			builder.WriteString(" JOIN ")
			builder.WriteString(join.table)
			builder.WriteString(" ON ")
			builder.WriteString(join.query)
		}
	}

	if w.where != nil {
		builder.WriteString(" WHERE ")
		builder.WriteString(w.where.query)

		binds = append(binds, w.where.binds...)
	}

	if len(w.groups) != 0 {
		builder.WriteString(" GROUP BY ")
		builder.WriteString(w.groups[0])

		for _, column := range w.groups[1:] {
			builder.WriteString(", ")
			builder.WriteString(column)
		}
	}

	if w.having != nil {
		builder.WriteString(" HAVING ")
		builder.WriteString(w.having.query)

		binds = append(binds, w.having.binds...)
	}

	if len(w.orders) != 0 {
		builder.WriteString(" ORDER BY ")
		builder.WriteString(w.orders[0])

		for _, column := range w.orders[1:] {
			builder.WriteString(", ")
			builder.WriteString(column)
		}
	}

	if w.offset != 0 {
		builder.WriteString(" OFFSET ?")
		binds = append(binds, w.offset)
	}

	if w.limit != 0 {
		builder.WriteString(" LIMIT ?")
		binds = append(binds, w.limit)
	}

	return builder.String(), binds
}

// ToInsert returns insert statement and binds.
// data expects `struct`, `*struct`, `yiigo.X`.
func (w *QueryWrapper) ToInsert(data interface{}) (string, []interface{}) {
	var (
		columns []string
		binds   []interface{}
	)

	v := reflect.Indirect(reflect.ValueOf(data))

	switch v.Kind() {
	case reflect.Map:
		x, ok := data.(X)

		if !ok {
			logger.Error("yiigo: invalid data type for insert, expects struct, *struct, yiigo.X")

			return "", nil
		}

		columns, binds = w.insertWithMap(x)
	case reflect.Struct:
		columns, binds = w.insertWithStruct(v)
	default:
		logger.Error("yiigo: invalid data type for insert, expects struct, *struct, yiigo.X")

		return "", nil
	}

	columnLen := len(columns)

	if columnLen == 0 {
		return "", nil
	}

	var builder strings.Builder

	builder.WriteString("INSERT INTO ")
	builder.WriteString(w.table)
	builder.WriteString(" (")
	builder.WriteString(columns[0])

	for _, column := range columns[1:] {
		builder.WriteString(", ")
		builder.WriteString(column)
	}

	builder.WriteString(") VALUES (?")

	for i := 1; i < columnLen; i++ {
		builder.WriteString(", ?")
	}

	builder.WriteString(")")

	if w.driver == Postgres {
		builder.WriteString(" RETURNING id")
	}

	query := sqlx.Rebind(sqlx.BindType(string(w.driver)), builder.String())

	if debug {
		logger.Info(query, zap.Any("binds", binds))
	}

	return query, binds
}

func (w *QueryWrapper) insertWithMap(data X) (columns []string, binds []interface{}) {
	fieldNum := len(data)

	columns = make([]string, 0, fieldNum)
	binds = make([]interface{}, 0, fieldNum)

	for k, v := range data {
		columns = append(columns, k)
		binds = append(binds, v)
	}

	return
}

func (w *QueryWrapper) insertWithStruct(v reflect.Value) (columns []string, binds []interface{}) {
	fieldNum := v.NumField()

	columns = make([]string, 0, fieldNum)
	binds = make([]interface{}, 0, fieldNum)

	t := v.Type()

	for i := 0; i < fieldNum; i++ {
		tag := t.Field(i).Tag.Get("db")

		if tag == "-" {
			continue
		}

		if tag == "" {
			tag = t.Field(i).Name
		}

		columns = append(columns, tag)
		binds = append(binds, v.Field(i).Interface())
	}

	return
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

	var (
		columns []string
		binds   []interface{}
	)

	e := v.Type().Elem()

	switch e.Kind() {
	case reflect.Map:
		x, ok := data.([]X)

		if !ok {
			logger.Error("yiigo: invalid data type for batch insert, expects []struct, []*struct, []yiigo.X")

			return "", nil
		}

		columns, binds = w.batchInsertWithMap(x)
	case reflect.Struct:
		columns, binds = w.batchInsertWithStruct(v)
	case reflect.Ptr:
		if e.Elem().Kind() != reflect.Struct {
			logger.Error("yiigo: invalid data type for batch insert, expects []struct, []*struct, []yiigo.X")

			return "", nil
		}

		columns, binds = w.batchInsertWithStruct(v)
	default:
		logger.Error("yiigo: invalid data type for batch insert, expects []struct, []*struct, []yiigo.X")

		return "", nil
	}

	columnLen := len(columns)

	if columnLen == 0 {
		return "", nil
	}

	var builder strings.Builder

	builder.WriteString("INSERT INTO ")
	builder.WriteString(w.table)
	builder.WriteString(" (")
	builder.WriteString(columns[0])

	for _, column := range columns[1:] {
		builder.WriteString(", ")
		builder.WriteString(column)
	}

	builder.WriteString(") VALUES (?")

	// 首行
	for i := 1; i < columnLen; i++ {
		builder.WriteString(", ?")
	}

	builder.WriteString(")")

	rows := len(binds) / columnLen

	// 其余行
	for i := 1; i < rows; i++ {
		builder.WriteString(", (?")

		for j := 1; j < columnLen; j++ {
			builder.WriteString(", ?")
		}

		builder.WriteString(")")
	}

	query := sqlx.Rebind(sqlx.BindType(string(w.driver)), builder.String())

	if debug {
		logger.Info(query, zap.Any("binds", binds))
	}

	return query, binds
}

func (w *QueryWrapper) batchInsertWithMap(data []X) (columns []string, binds []interface{}) {
	dataLen := len(data)
	fieldNum := len(data[0])

	columns = make([]string, 0, fieldNum)
	binds = make([]interface{}, 0, fieldNum*dataLen)

	for k := range data[0] {
		columns = append(columns, k)
	}

	for _, x := range data {
		for _, v := range columns {
			binds = append(binds, x[v])
		}
	}

	return
}

func (w *QueryWrapper) batchInsertWithStruct(v reflect.Value) (columns []string, binds []interface{}) {
	first := reflect.Indirect(v.Index(0))

	dataLen := v.Len()
	fieldNum := first.NumField()

	columns = make([]string, 0, fieldNum)
	binds = make([]interface{}, 0, fieldNum*dataLen)

	t := first.Type()

	for i := 0; i < dataLen; i++ {
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

			binds = append(binds, reflect.Indirect(v.Index(i)).Field(j).Interface())
		}
	}

	return
}

// ToUpdate returns update statement and binds.
// data expects `struct`, `*struct`, `yiigo.X`.
func (w *QueryWrapper) ToUpdate(data interface{}) (string, []interface{}) {
	var (
		columns []string
		exprs   map[string]string
		binds   []interface{}
	)

	v := reflect.Indirect(reflect.ValueOf(data))

	switch v.Kind() {
	case reflect.Map:
		x, ok := data.(X)

		if !ok {
			logger.Error("yiigo: invalid data type for update, expects struct, *struct, yiigo.X")

			return "", nil
		}

		columns, exprs, binds = w.updateWithMap(x)
	case reflect.Struct:
		columns, binds = w.updateWithStruct(v)
	default:
		logger.Error("yiigo: invalid data type for update, expects struct, *struct, yiigo.X")

		return "", nil
	}

	if len(columns) == 0 {
		return "", nil
	}

	var builder strings.Builder

	builder.WriteString("UPDATE ")
	builder.WriteString(w.table)
	builder.WriteString(" SET ")
	builder.WriteString(columns[0])

	if expr, ok := exprs[columns[0]]; ok {
		builder.WriteString(" = ")
		builder.WriteString(expr)
	} else {
		builder.WriteString(" = ?")
	}

	for _, column := range columns[1:] {
		builder.WriteString(", ")
		builder.WriteString(column)

		if expr, ok := exprs[column]; ok {
			builder.WriteString(" = ")
			builder.WriteString(expr)
		} else {
			builder.WriteString(" = ?")
		}
	}

	if w.where != nil {
		builder.WriteString(" WHERE ")
		builder.WriteString(w.where.query)

		binds = append(binds, w.where.binds...)
	}

	query := builder.String()

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

func (w *QueryWrapper) updateWithMap(data X) (columns []string, exprs map[string]string, binds []interface{}) {
	fieldNum := len(data)

	columns = make([]string, 0, fieldNum)
	exprs = make(map[string]string)
	binds = make([]interface{}, 0, fieldNum)

	for k, v := range data {
		columns = append(columns, k)

		if clause, ok := v.(*SQLClause); ok {
			exprs[k] = clause.query
			binds = append(binds, clause.binds...)

			continue
		}

		binds = append(binds, v)
	}

	return
}

func (w *QueryWrapper) updateWithStruct(v reflect.Value) (columns []string, binds []interface{}) {
	fieldNum := v.NumField()

	columns = make([]string, 0, fieldNum)
	binds = make([]interface{}, 0, fieldNum)

	t := v.Type()

	for i := 0; i < fieldNum; i++ {
		tag := t.Field(i).Tag.Get("db")

		if tag == "-" {
			continue
		}

		if tag == "" {
			tag = t.Field(i).Name
		}

		columns = append(columns, tag)
		binds = append(binds, v.Field(i).Interface())
	}

	return
}

// ToDelete returns delete clause and binds.
func (w *QueryWrapper) ToDelete() (string, []interface{}) {
	binds := make([]interface{}, 0)

	var builder strings.Builder

	builder.WriteString("DELETE FROM ")
	builder.WriteString(w.table)

	if w.where != nil {
		builder.WriteString(" WHERE ")
		builder.WriteString(w.where.query)

		binds = append(binds, w.where.binds...)
	}

	query := builder.String()

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
		w.joins = append(w.joins, &SQLClause{
			table:   table,
			keyword: "INNER",
			query:   on,
		})
	}
}

// LeftJoin specifies the `left join` clause.
func LeftJoin(table, on string) QueryOption {
	return func(w *QueryWrapper) {
		w.joins = append(w.joins, &SQLClause{
			table:   table,
			keyword: "LEFT",
			query:   on,
		})
	}
}

// RightJoin specifies the `right join` clause.
func RightJoin(table, on string) QueryOption {
	return func(w *QueryWrapper) {
		w.joins = append(w.joins, &SQLClause{
			table:   table,
			keyword: "RIGHT",
			query:   on,
		})
	}
}

// FullJoin specifies the `full join` clause.
func FullJoin(table, on string) QueryOption {
	return func(w *QueryWrapper) {
		w.joins = append(w.joins, &SQLClause{
			table:   table,
			keyword: "FULL",
			query:   on,
		})
	}
}

// Where specifies the `where` clause.
func Where(query string, binds ...interface{}) QueryOption {
	return func(w *QueryWrapper) {
		w.where = &SQLClause{
			query: query,
			binds: binds,
		}
	}
}

// WhereIn specifies the `where in` clause.
func WhereIn(query string, binds ...interface{}) QueryOption {
	return func(w *QueryWrapper) {
		w.where = &SQLClause{
			query: query,
			binds: binds,
		}

		w.whereIn = true
	}
}

// GroupBy specifies the `group by` clause.
func GroupBy(columns ...string) QueryOption {
	return func(w *QueryWrapper) {
		w.groups = columns
	}
}

// Having specifies the `having` clause.
func Having(query string, binds ...interface{}) QueryOption {
	return func(w *QueryWrapper) {
		w.having = &SQLClause{
			query: query,
			binds: binds,
		}
	}
}

// OrderBy specifies the `order by` clause.
func OrderBy(columns ...string) QueryOption {
	return func(w *QueryWrapper) {
		w.orders = columns
	}
}

// Offset specifies the `offset` clause.
func Offset(n int) QueryOption {
	return func(w *QueryWrapper) {
		w.offset = n
	}
}

// Limit specifies the `limit` clause.
func Limit(n int) QueryOption {
	return func(w *QueryWrapper) {
		w.limit = n
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

			query, binds := v.subquery()

			w.unions = append(w.unions, &SQLClause{
				keyword: "UNION",
				query:   query,
				binds:   binds,
			})
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

			query, binds := v.subquery()

			w.unions = append(w.unions, &SQLClause{
				keyword: "UNION ALL",
				query:   query,
				binds:   binds,
			})
		}
	}
}
