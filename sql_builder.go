package yiigo

import (
	"reflect"
	"strings"

	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

// SQLBuilder is the interface for wrapping query options
type SQLBuilder interface {
	// Wrap wrapping query options
	Wrap(options ...QueryOption) SQLWrapper
}

// SQLWrapper is the interface for building sql statement
type SQLWrapper interface {
	// ToQuery returns query statement and binds.
	ToQuery() (string, []interface{})

	// ToInsert returns insert statement and binds.
	// data expects `struct`, `*struct`, `yiigo.X`.
	ToInsert(data interface{}) (string, []interface{})

	// ToBatchInsert returns batch insert statement and binds.
	// data expects `[]struct`, `[]*struct`, `[]yiigo.X`.
	ToBatchInsert(data interface{}) (string, []interface{})

	// ToUpdate returns update statement and binds.
	// data expects `struct`, `*struct`, `yiigo.X`.
	ToUpdate(data interface{}) (string, []interface{})

	// ToDelete returns delete statement and binds.
	ToDelete() (string, []interface{})

	// ToTruncate returns truncate statement
	ToTruncate() string
}

type queryBuilder struct {
	driver DBDriver
}

func (b *queryBuilder) Wrap(options ...QueryOption) SQLWrapper {
	wrapper := &queryWrapper{
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
	return &queryBuilder{driver: driver}
}

// NewMySQLBuilder returns new SQLBuilder for MySQL
func NewMySQLBuilder() SQLBuilder {
	return NewSQLBuilder(MySQL)
}

// NewPGSQLBuilder returns new SQLBuilder for Postgres
func NewPGSQLBuilder() SQLBuilder {
	return NewSQLBuilder(Postgres)
}

// NewSQLiteBuilder returns new SQLBuilder for SQLite
func NewSQLiteBuilder() SQLBuilder {
	return NewSQLBuilder(SQLite)
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

type queryWrapper struct {
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

func (w *queryWrapper) ToQuery() (string, []interface{}) {
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

	if Debug {
		logger.Info(query, zap.Any("binds", binds))
	}

	return query, binds
}

func (w *queryWrapper) subquery() (string, []interface{}) {
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

			if len(join.query) != 0 {
				builder.WriteString(" ON ")
				builder.WriteString(join.query)
			}
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

	if w.limit != 0 {
		builder.WriteString(" LIMIT ?")
		binds = append(binds, w.limit)
	}

	if w.offset != 0 {
		builder.WriteString(" OFFSET ?")
		binds = append(binds, w.offset)
	}

	return builder.String(), binds
}

func (w *queryWrapper) ToInsert(data interface{}) (string, []interface{}) {
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

	if Debug {
		logger.Info(query, zap.Any("binds", binds))
	}

	return query, binds
}

func (w *queryWrapper) insertWithMap(data X) (columns []string, binds []interface{}) {
	fieldNum := len(data)

	columns = make([]string, 0, fieldNum)
	binds = make([]interface{}, 0, fieldNum)

	for k, v := range data {
		columns = append(columns, k)
		binds = append(binds, v)
	}

	return
}

func (w *queryWrapper) insertWithStruct(v reflect.Value) (columns []string, binds []interface{}) {
	fieldNum := v.NumField()

	columns = make([]string, 0, fieldNum)
	binds = make([]interface{}, 0, fieldNum)

	t := v.Type()

	for i := 0; i < fieldNum; i++ {
		fieldT := t.Field(i)
		tag := fieldT.Tag.Get("db")

		if tag == "-" {
			continue
		}

		fieldV := v.Field(i)
		column := fieldT.Name

		if tag != "" {
			name, opts := parseTag(tag)

			if opts.Contains("omitempty") && isEmptyValue(fieldV) {
				continue
			}

			column = name
		}

		columns = append(columns, column)
		binds = append(binds, fieldV.Interface())
	}

	return
}

func (w *queryWrapper) ToBatchInsert(data interface{}) (string, []interface{}) {
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

	if Debug {
		logger.Info(query, zap.Any("binds", binds))
	}

	return query, binds
}

func (w *queryWrapper) batchInsertWithMap(data []X) (columns []string, binds []interface{}) {
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

func (w *queryWrapper) batchInsertWithStruct(v reflect.Value) (columns []string, binds []interface{}) {
	first := reflect.Indirect(v.Index(0))

	dataLen := v.Len()
	fieldNum := first.NumField()

	columns = make([]string, 0, fieldNum)
	binds = make([]interface{}, 0, fieldNum*dataLen)

	t := first.Type()

	for i := 0; i < dataLen; i++ {
		for j := 0; j < fieldNum; j++ {
			fieldT := t.Field(j)

			tag := fieldT.Tag.Get("db")

			if tag == "-" {
				continue
			}

			fieldV := reflect.Indirect(v.Index(i)).Field(j)
			column := fieldT.Name

			if tag != "" {
				name, opts := parseTag(tag)

				if opts.Contains("omitempty") && isEmptyValue(fieldV) {
					continue
				}

				column = name
			}

			if i == 0 {
				columns = append(columns, column)
			}

			binds = append(binds, fieldV.Interface())
		}
	}

	return
}

func (w *queryWrapper) ToUpdate(data interface{}) (string, []interface{}) {
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

	if Debug {
		logger.Info(query, zap.Any("binds", binds))
	}

	return query, binds
}

func (w *queryWrapper) updateWithMap(data X) (columns []string, exprs map[string]string, binds []interface{}) {
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

func (w *queryWrapper) updateWithStruct(v reflect.Value) (columns []string, binds []interface{}) {
	fieldNum := v.NumField()

	columns = make([]string, 0, fieldNum)
	binds = make([]interface{}, 0, fieldNum)

	t := v.Type()

	for i := 0; i < fieldNum; i++ {
		fieldT := t.Field(i)
		tag := fieldT.Tag.Get("db")

		if tag == "-" {
			continue
		}

		fieldV := v.Field(i)
		column := fieldT.Name

		if tag != "" {
			name, opts := parseTag(tag)

			if opts.Contains("omitempty") && isEmptyValue(fieldV) {
				continue
			}

			column = name
		}

		columns = append(columns, column)
		binds = append(binds, fieldV.Interface())
	}

	return
}

func (w *queryWrapper) ToDelete() (string, []interface{}) {
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

	if Debug {
		logger.Info(query, zap.Any("binds", binds))
	}

	return query, binds
}

func (w *queryWrapper) ToTruncate() string {
	var builder strings.Builder

	builder.WriteString("TRUNCATE ")
	builder.WriteString(w.table)

	query := builder.String()

	if Debug {
		logger.Info(query)
	}

	return query
}

// QueryOption configures how we set up the SQL query statement
type QueryOption func(w *queryWrapper)

// Table specifies the query table.
func Table(name string) QueryOption {
	return func(w *queryWrapper) {
		w.table = name
	}
}

// Select specifies the query columns.
func Select(columns ...string) QueryOption {
	return func(w *queryWrapper) {
		w.columns = columns
	}
}

// Distinct specifies the `distinct` clause.
func Distinct(columns ...string) QueryOption {
	return func(w *queryWrapper) {
		w.columns = columns
		w.distinct = true
	}
}

// Join specifies the `inner join` clause.
func Join(table, on string) QueryOption {
	return func(w *queryWrapper) {
		w.joins = append(w.joins, &SQLClause{
			table:   table,
			keyword: "INNER",
			query:   on,
		})
	}
}

// LeftJoin specifies the `left join` clause.
func LeftJoin(table, on string) QueryOption {
	return func(w *queryWrapper) {
		w.joins = append(w.joins, &SQLClause{
			table:   table,
			keyword: "LEFT",
			query:   on,
		})
	}
}

// RightJoin specifies the `right join` clause.
func RightJoin(table, on string) QueryOption {
	return func(w *queryWrapper) {
		w.joins = append(w.joins, &SQLClause{
			table:   table,
			keyword: "RIGHT",
			query:   on,
		})
	}
}

// FullJoin specifies the `full join` clause.
func FullJoin(table, on string) QueryOption {
	return func(w *queryWrapper) {
		w.joins = append(w.joins, &SQLClause{
			table:   table,
			keyword: "FULL",
			query:   on,
		})
	}
}

// CrossJoin specifies the `cross join` clause.
func CrossJoin(table string) QueryOption {
	return func(w *queryWrapper) {
		w.joins = append(w.joins, &SQLClause{
			table:   table,
			keyword: "CROSS",
		})
	}
}

// Where specifies the `where` clause.
func Where(query string, binds ...interface{}) QueryOption {
	return func(w *queryWrapper) {
		w.where = &SQLClause{
			query: query,
			binds: binds,
		}
	}
}

// WhereIn specifies the `where in` clause.
func WhereIn(query string, binds ...interface{}) QueryOption {
	return func(w *queryWrapper) {
		w.where = &SQLClause{
			query: query,
			binds: binds,
		}

		w.whereIn = true
	}
}

// GroupBy specifies the `group by` clause.
func GroupBy(columns ...string) QueryOption {
	return func(w *queryWrapper) {
		w.groups = columns
	}
}

// Having specifies the `having` clause.
func Having(query string, binds ...interface{}) QueryOption {
	return func(w *queryWrapper) {
		w.having = &SQLClause{
			query: query,
			binds: binds,
		}
	}
}

// OrderBy specifies the `order by` clause.
func OrderBy(columns ...string) QueryOption {
	return func(w *queryWrapper) {
		w.orders = columns
	}
}

// Offset specifies the `offset` clause.
func Offset(n int) QueryOption {
	return func(w *queryWrapper) {
		w.offset = n
	}
}

// Limit specifies the `limit` clause.
func Limit(n int) QueryOption {
	return func(w *queryWrapper) {
		w.limit = n
	}
}

// Union specifies the `union` clause.
func Union(wrappers ...SQLWrapper) QueryOption {
	return func(w *queryWrapper) {
		for _, wrapper := range wrappers {
			v, ok := wrapper.(*queryWrapper)

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
	return func(w *queryWrapper) {
		for _, wrapper := range wrappers {
			v, ok := wrapper.(*queryWrapper)

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

// tagOptions is the string following a comma in a struct field's "json"
// tag, or the empty string. It does not include the leading comma.
type tagOptions string

// Contains reports whether a comma-separated list of options
// contains a particular substr flag. substr must be surrounded by a
// string boundary or commas.
func (o tagOptions) Contains(optionName string) bool {
	if len(o) == 0 {
		return false
	}

	s := string(o)

	for s != "" {
		var next string

		i := strings.Index(s, ",")

		if i >= 0 {
			s, next = s[:i], s[i+1:]
		}

		if s == optionName {
			return true
		}

		s = next
	}

	return false
}

// parseTag splits a struct field's json tag into its name and
// comma-separated options.
func parseTag(tag string) (string, tagOptions) {
	if idx := strings.Index(tag, ","); idx != -1 {
		return tag[:idx], tagOptions(tag[idx+1:])
	}

	return tag, tagOptions("")
}

func isEmptyValue(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Array, reflect.Map, reflect.Slice, reflect.String:
		return v.Len() == 0
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Interface, reflect.Ptr:
		return v.IsNil()
	}

	return false
}
