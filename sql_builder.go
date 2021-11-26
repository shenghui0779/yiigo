package yiigo

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

var (
	// ErrInvalidUpsertData invalid insert or update data.
	ErrInvalidUpsertData = errors.New("invaild data, expects struct, *struct, yiigo.X")
	// ErrInvalidBatchInsertData invalid batch insert data.
	ErrInvalidBatchInsertData = errors.New("invaild data, expects []struct, []*struct, []yiigo.X")
)

// SQLBuilder is the interface for wrapping query options.
type SQLBuilder interface {
	// Wrap wrapping query options
	Wrap(options ...QueryOption) SQLWrapper
}

// SQLWrapper is the interface for building sql statement.
type SQLWrapper interface {
	// ToQuery returns query statement and binds.
	ToQuery(ctx context.Context) (string, []interface{})

	// ToInsert returns insert statement and binds.
	// data expects `struct`, `*struct`, `yiigo.X`.
	ToInsert(ctx context.Context, data interface{}) (string, []interface{})

	// ToBatchInsert returns batch insert statement and binds.
	// data expects `[]struct`, `[]*struct`, `[]yiigo.X`.
	ToBatchInsert(ctx context.Context, data interface{}) (string, []interface{})

	// ToUpdate returns update statement and binds.
	// data expects `struct`, `*struct`, `yiigo.X`.
	ToUpdate(ctx context.Context, data interface{}) (string, []interface{})

	// ToDelete returns delete statement and binds.
	ToDelete(ctx context.Context) (string, []interface{})

	// ToTruncate returns truncate statement
	ToTruncate(ctx context.Context) string
}

// SQLLogger is the log interface for sql builder.
type SQLLogger interface {
	// Info logs sql statement in debug mode for sql builder.
	Info(ctx context.Context, query string, args ...interface{})

	// Error logs an error for sql builder.
	Error(ctx context.Context, err error)
}

type builderLog struct{}

func (l *builderLog) Info(ctx context.Context, query string, args ...interface{}) {
	logger.Info(fmt.Sprintf("[yiigo] [SQL] %s", query), zap.Any("args", args))
}

func (l *builderLog) Error(ctx context.Context, err error) {
	logger.Error("[yiigo] SQL Builder Error", zap.Error(err))
}

type queryBuilder struct {
	driver DBDriver
	logger SQLLogger
	debug  bool
}

func (b *queryBuilder) Wrap(options ...QueryOption) SQLWrapper {
	wrapper := &queryWrapper{
		builder: b,
		columns: []string{"*"},
	}

	for _, f := range options {
		f(wrapper)
	}

	return wrapper
}

// NewSQLBuilder returns new SQLBuilder
func NewSQLBuilder(driver DBDriver, options ...BuilderOption) SQLBuilder {
	builder := &queryBuilder{
		driver: driver,
		logger: new(builderLog),
	}

	for _, f := range options {
		f(builder)
	}

	return builder
}

// NewMySQLBuilder returns new SQLBuilder for MySQL
func NewMySQLBuilder(options ...BuilderOption) SQLBuilder {
	return NewSQLBuilder(MySQL, options...)
}

// NewPGSQLBuilder returns new SQLBuilder for Postgres
func NewPGSQLBuilder(options ...BuilderOption) SQLBuilder {
	return NewSQLBuilder(Postgres, options...)
}

// NewSQLiteBuilder returns new SQLBuilder for SQLite
func NewSQLiteBuilder(options ...BuilderOption) SQLBuilder {
	return NewSQLBuilder(SQLite, options...)
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
	builder  *queryBuilder
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

func (w *queryWrapper) ToQuery(ctx context.Context) (string, []interface{}) {
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
			w.builder.logger.Error(ctx, errors.Wrap(err, "error build 'IN' query"))

			return "", nil
		}
	}

	query = sqlx.Rebind(sqlx.BindType(string(w.builder.driver)), query)

	if w.builder.debug {
		w.builder.logger.Info(ctx, query, binds...)
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

func (w *queryWrapper) ToInsert(ctx context.Context, data interface{}) (string, []interface{}) {
	var (
		columns []string
		binds   []interface{}
	)

	v := reflect.Indirect(reflect.ValueOf(data))

	switch v.Kind() {
	case reflect.Map:
		x, ok := data.(X)

		if !ok {
			w.builder.logger.Error(ctx, ErrInvalidUpsertData)

			return "", nil
		}

		columns, binds = w.insertWithMap(x)
	case reflect.Struct:
		columns, binds = w.insertWithStruct(v)
	default:
		w.builder.logger.Error(ctx, ErrInvalidUpsertData)

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

	if w.builder.driver == Postgres {
		builder.WriteString(" RETURNING id")
	}

	query := sqlx.Rebind(sqlx.BindType(string(w.builder.driver)), builder.String())

	if w.builder.debug {
		w.builder.logger.Info(ctx, query, binds...)
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

		if len(tag) != 0 {
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

func (w *queryWrapper) ToBatchInsert(ctx context.Context, data interface{}) (string, []interface{}) {
	v := reflect.Indirect(reflect.ValueOf(data))

	if v.Kind() != reflect.Slice {
		w.builder.logger.Error(ctx, ErrInvalidBatchInsertData)

		return "", nil
	}

	if v.Len() == 0 {
		w.builder.logger.Error(ctx, errors.New("error empty data"))

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
			w.builder.logger.Error(ctx, ErrInvalidBatchInsertData)

			return "", nil
		}

		columns, binds = w.batchInsertWithMap(x)
	case reflect.Struct:
		columns, binds = w.batchInsertWithStruct(v)
	case reflect.Ptr:
		if e.Elem().Kind() != reflect.Struct {
			w.builder.logger.Error(ctx, ErrInvalidBatchInsertData)

			return "", nil
		}

		columns, binds = w.batchInsertWithStruct(v)
	default:
		w.builder.logger.Error(ctx, ErrInvalidBatchInsertData)

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

	query := sqlx.Rebind(sqlx.BindType(string(w.builder.driver)), builder.String())

	if w.builder.debug {
		w.builder.logger.Info(ctx, query, binds...)
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

			if len(tag) != 0 {
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

func (w *queryWrapper) ToUpdate(ctx context.Context, data interface{}) (string, []interface{}) {
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
			w.builder.logger.Error(ctx, ErrInvalidUpsertData)

			return "", nil
		}

		columns, exprs, binds = w.updateWithMap(x)
	case reflect.Struct:
		columns, binds = w.updateWithStruct(v)
	default:
		w.builder.logger.Error(ctx, ErrInvalidUpsertData)

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
			w.builder.logger.Error(ctx, errors.Wrap(err, "error build 'IN' query"))

			return "", nil
		}
	}

	query = sqlx.Rebind(sqlx.BindType(string(w.builder.driver)), query)

	if w.builder.debug {
		w.builder.logger.Info(ctx, query, binds...)
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

		if len(tag) != 0 {
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

func (w *queryWrapper) ToDelete(ctx context.Context) (string, []interface{}) {
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
			w.builder.logger.Error(ctx, errors.Wrap(err, "error build 'IN' query"))

			return "", nil
		}
	}

	query = sqlx.Rebind(sqlx.BindType(string(w.builder.driver)), query)

	if w.builder.debug {
		w.builder.logger.Info(ctx, query, binds...)
	}

	return query, binds
}

func (w *queryWrapper) ToTruncate(ctx context.Context) string {
	var builder strings.Builder

	builder.WriteString("TRUNCATE ")
	builder.WriteString(w.table)

	query := builder.String()

	if w.builder.debug {
		w.builder.logger.Info(ctx, query)
	}

	return query
}

// BuilderOption configures how we set up the SQL builder.
type BuilderOption func(builder *queryBuilder)

// WithSQLLogger sets logger for SQL builder.
func WithSQLLogger(l SQLLogger) BuilderOption {
	return func(builder *queryBuilder) {
		builder.logger = l
	}
}

// WithSQLDebug sets debug mode for SQL builder.
func WithSQLDebug() BuilderOption {
	return func(builder *queryBuilder) {
		builder.debug = true
	}
}

// QueryOption configures how we set up the SQL query statement.
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

	for len(s) != 0 {
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
