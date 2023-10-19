package yiigo

import (
	"context"
	"errors"
	"reflect"
	"strings"

	"github.com/jmoiron/sqlx"
)

var (
	// ErrUpsertData 不合法的插入或更新数据类型错误
	ErrUpsertData = errors.New("invaild data, expects struct, *struct, yiigo.X")

	// ErrBatchInsertData 不合法的批量插入数据类型错误
	ErrBatchInsertData = errors.New("invaild data, expects []struct, []*struct, []yiigo.X")
)

// SQLBuilder SQL构造器
type SQLBuilder interface {
	// Wrap 包装查询选项
	Wrap(options ...QueryOption) SQLWrapper
}

// SQLWrapper SQL包装器
type SQLWrapper interface {
	// ToQuery 生成SELECT语句
	ToQuery(ctx context.Context) (sql string, args []any, err error)

	// ToInsert 生成INSERT语句
	// 数据类型：`struct`, `*struct`, `yiigo.X`.
	ToInsert(ctx context.Context, data any) (sql string, args []any, err error)

	// ToBatchInsert 生成批量INSERT语句
	// 数据类型：`[]struct`, `[]*struct`, `[]yiigo.X`.
	ToBatchInsert(ctx context.Context, data any) (sql string, args []any, err error)

	// ToUpdate 生成UPDATE语句
	// 数据类型：`struct`, `*struct`, `yiigo.X`.
	ToUpdate(ctx context.Context, data any) (sql string, args []any, err error)

	// ToDelete 生成DELETE语句
	ToDelete(ctx context.Context) (sql string, args []any, err error)

	// ToTruncate 生成TRUNCATE语句
	ToTruncate(ctx context.Context) string
}

type queryBuilder struct {
	driver DBDriver
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

// NewSQLBuilder 生成一个指定驱动类型的SQL构造器
func NewSQLBuilder(driver DBDriver) SQLBuilder {
	return &queryBuilder{
		driver: driver,
	}
}

// NewMySQLBuilder 生成一个MySQL构造器
func NewMySQLBuilder() SQLBuilder {
	return NewSQLBuilder(MySQL)
}

// NewPGSQLBuilder 生成一个Postgres构造器
func NewPGSQLBuilder() SQLBuilder {
	return NewSQLBuilder(Postgres)
}

// NewSQLiteBuilder 生成一个SQLite构造器
func NewSQLiteBuilder() SQLBuilder {
	return NewSQLBuilder(SQLite)
}

// SQLClause SQL语句
type SQLClause struct {
	table   string
	keyword string
	query   string
	binds   []any
}

// SQLExpr 生成一个语句表达式，例如：yiigo.SQLExpr("price * ? + ?", 2, 100)
func SQLExpr(query string, binds ...any) *SQLClause {
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

func (w *queryWrapper) ToQuery(ctx context.Context) (sql string, args []any, err error) {
	sql, args = w.subquery()

	// unions
	if l := len(w.unions); l != 0 {
		var builder strings.Builder

		builder.WriteString("(")
		builder.WriteString(sql)
		builder.WriteString(")")

		for _, v := range w.unions {
			builder.WriteString(" ")
			builder.WriteString(v.keyword)
			builder.WriteString(" (")
			builder.WriteString(v.query)
			builder.WriteString(")")

			args = append(args, v.binds...)
		}

		sql = builder.String()
	}

	// where in
	if w.whereIn {
		sql, args, err = sqlx.In(sql, args...)
		if err != nil {
			return
		}
	}

	sql = sqlx.Rebind(sqlx.BindType(string(w.builder.driver)), sql)

	return
}

func (w *queryWrapper) subquery() (string, []any) {
	binds := make([]any, 0)

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

func (w *queryWrapper) ToInsert(ctx context.Context, data any) (sql string, args []any, err error) {
	var columns []string

	v := reflect.Indirect(reflect.ValueOf(data))

	switch v.Kind() {
	case reflect.Map:
		x, ok := data.(X)
		if !ok {
			err = ErrUpsertData
			return
		}

		columns, args = w.insertWithMap(x)
	case reflect.Struct:
		columns, args = w.insertWithStruct(v)
	default:
		err = ErrUpsertData
		return
	}

	var builder strings.Builder

	builder.WriteString("INSERT INTO ")
	builder.WriteString(w.table)

	if l := len(columns); l != 0 {
		builder.WriteString(" (")
		builder.WriteString(columns[0])

		for _, column := range columns[1:] {
			builder.WriteString(", ")
			builder.WriteString(column)
		}

		builder.WriteString(") VALUES (?")
		for i := 1; i < l; i++ {
			builder.WriteString(", ?")
		}
		builder.WriteString(")")
	}

	if w.builder.driver == Postgres {
		builder.WriteString(" RETURNING id")
	}

	sql = sqlx.Rebind(sqlx.BindType(string(w.builder.driver)), builder.String())

	return
}

func (w *queryWrapper) insertWithMap(data X) (columns []string, binds []any) {
	fieldNum := len(data)

	columns = make([]string, 0, fieldNum)
	binds = make([]any, 0, fieldNum)

	for k, v := range data {
		columns = append(columns, k)
		binds = append(binds, v)
	}

	return
}

func (w *queryWrapper) insertWithStruct(v reflect.Value) (columns []string, binds []any) {
	fieldNum := v.NumField()

	columns = make([]string, 0, fieldNum)
	binds = make([]any, 0, fieldNum)

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

func (w *queryWrapper) ToBatchInsert(ctx context.Context, data any) (sql string, args []any, err error) {
	v := reflect.Indirect(reflect.ValueOf(data))

	if v.Kind() != reflect.Slice {
		err = ErrBatchInsertData
		return
	}

	if v.Len() == 0 {
		err = errors.New("err empty data")
		return
	}

	var columns []string

	e := v.Type().Elem()

	switch e.Kind() {
	case reflect.Map:
		x, ok := data.([]X)
		if !ok {
			err = ErrBatchInsertData
			return
		}

		columns, args = w.batchInsertWithMap(x)
	case reflect.Struct:
		columns, args = w.batchInsertWithStruct(v)
	case reflect.Ptr:
		if e.Elem().Kind() != reflect.Struct {
			err = ErrBatchInsertData
			return
		}

		columns, args = w.batchInsertWithStruct(v)
	default:
		err = ErrBatchInsertData
		return
	}

	var builder strings.Builder

	builder.WriteString("INSERT INTO ")
	builder.WriteString(w.table)

	if l := len(columns); l != 0 {
		builder.WriteString(" (")
		builder.WriteString(columns[0])

		for _, column := range columns[1:] {
			builder.WriteString(", ")
			builder.WriteString(column)
		}

		// 首行
		builder.WriteString(") VALUES (?")
		for i := 1; i < l; i++ {
			builder.WriteString(", ?")
		}
		builder.WriteString(")")

		rows := len(args) / l

		// 其余行
		for i := 1; i < rows; i++ {
			builder.WriteString(", (?")
			for j := 1; j < l; j++ {
				builder.WriteString(", ?")
			}
			builder.WriteString(")")
		}
	}

	sql = sqlx.Rebind(sqlx.BindType(string(w.builder.driver)), builder.String())

	return
}

func (w *queryWrapper) batchInsertWithMap(data []X) (columns []string, binds []any) {
	dataLen := len(data)
	fieldNum := len(data[0])

	columns = make([]string, 0, fieldNum)
	binds = make([]any, 0, fieldNum*dataLen)

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

func (w *queryWrapper) batchInsertWithStruct(v reflect.Value) (columns []string, binds []any) {
	first := reflect.Indirect(v.Index(0))

	dataLen := v.Len()
	fieldNum := first.NumField()

	columns = make([]string, 0, fieldNum)
	binds = make([]any, 0, fieldNum*dataLen)

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

func (w *queryWrapper) ToUpdate(ctx context.Context, data any) (sql string, args []any, err error) {
	var (
		columns []string
		exprs   map[string]string
	)

	v := reflect.Indirect(reflect.ValueOf(data))

	switch v.Kind() {
	case reflect.Map:
		x, ok := data.(X)
		if !ok {
			err = ErrUpsertData
			return
		}

		columns, exprs, args = w.updateWithMap(x)
	case reflect.Struct:
		columns, args = w.updateWithStruct(v)
	default:
		err = ErrUpsertData
		return
	}

	var builder strings.Builder

	builder.WriteString("UPDATE ")
	builder.WriteString(w.table)

	if len(columns) != 0 {
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
	}

	if w.where != nil {
		builder.WriteString(" WHERE ")
		builder.WriteString(w.where.query)

		args = append(args, w.where.binds...)
	}

	sql = builder.String()

	if w.whereIn {
		sql, args, err = sqlx.In(sql, args...)
		if err != nil {
			return
		}
	}

	sql = sqlx.Rebind(sqlx.BindType(string(w.builder.driver)), sql)

	return
}

func (w *queryWrapper) updateWithMap(data X) (columns []string, exprs map[string]string, binds []any) {
	fieldNum := len(data)

	columns = make([]string, 0, fieldNum)
	exprs = make(map[string]string)
	binds = make([]any, 0, fieldNum)

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

func (w *queryWrapper) updateWithStruct(v reflect.Value) (columns []string, binds []any) {
	fieldNum := v.NumField()

	columns = make([]string, 0, fieldNum)
	binds = make([]any, 0, fieldNum)

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

func (w *queryWrapper) ToDelete(ctx context.Context) (sql string, args []any, err error) {
	var builder strings.Builder

	builder.WriteString("DELETE FROM ")
	builder.WriteString(w.table)

	if w.where != nil {
		builder.WriteString(" WHERE ")
		builder.WriteString(w.where.query)

		args = append(args, w.where.binds...)
	}

	sql = builder.String()

	if w.whereIn {
		sql, args, err = sqlx.In(sql, args...)
		if err != nil {
			return
		}
	}

	sql = sqlx.Rebind(sqlx.BindType(string(w.builder.driver)), sql)

	return
}

func (w *queryWrapper) ToTruncate(ctx context.Context) string {
	var builder strings.Builder

	builder.WriteString("TRUNCATE ")
	builder.WriteString(w.table)

	return builder.String()
}

// QueryOption SQL查询选项
type QueryOption func(w *queryWrapper)

// Table 指定查询表名称
func Table(name string) QueryOption {
	return func(w *queryWrapper) {
		w.table = name
	}
}

// Select 指定查询字段名
func Select(columns ...string) QueryOption {
	return func(w *queryWrapper) {
		w.columns = columns
	}
}

// Distinct 指定 `DISTINCT` 语句
func Distinct(columns ...string) QueryOption {
	return func(w *queryWrapper) {
		w.columns = columns
		w.distinct = true
	}
}

// Join 指定 `INNER JOIN` 语句
func Join(table, on string) QueryOption {
	return func(w *queryWrapper) {
		w.joins = append(w.joins, &SQLClause{
			table:   table,
			keyword: "INNER",
			query:   on,
		})
	}
}

// LeftJoin 指定 `LEFT JOIN` 语句
func LeftJoin(table, on string) QueryOption {
	return func(w *queryWrapper) {
		w.joins = append(w.joins, &SQLClause{
			table:   table,
			keyword: "LEFT",
			query:   on,
		})
	}
}

// RightJoin 指定 `RIGHT JOIN` 语句
func RightJoin(table, on string) QueryOption {
	return func(w *queryWrapper) {
		w.joins = append(w.joins, &SQLClause{
			table:   table,
			keyword: "RIGHT",
			query:   on,
		})
	}
}

// FullJoin 指定 `FULL JOIN` 语句
func FullJoin(table, on string) QueryOption {
	return func(w *queryWrapper) {
		w.joins = append(w.joins, &SQLClause{
			table:   table,
			keyword: "FULL",
			query:   on,
		})
	}
}

// CrossJoin 指定 `CROSS JOIN` 语句
func CrossJoin(table string) QueryOption {
	return func(w *queryWrapper) {
		w.joins = append(w.joins, &SQLClause{
			table:   table,
			keyword: "CROSS",
		})
	}
}

// Where 指定 `WHERE` 语句
func Where(query string, binds ...any) QueryOption {
	return func(w *queryWrapper) {
		w.where = &SQLClause{
			query: query,
			binds: binds,
		}
	}
}

// WhereIn 指定 `WHERE IN` 语句
func WhereIn(query string, binds ...any) QueryOption {
	return func(w *queryWrapper) {
		w.where = &SQLClause{
			query: query,
			binds: binds,
		}

		w.whereIn = true
	}
}

// GroupBy 指定 `GROUP BY` 语句
func GroupBy(columns ...string) QueryOption {
	return func(w *queryWrapper) {
		w.groups = columns
	}
}

// Having 指定 `HAVING` 语句
func Having(query string, binds ...any) QueryOption {
	return func(w *queryWrapper) {
		w.having = &SQLClause{
			query: query,
			binds: binds,
		}
	}
}

// OrderBy 指定 `ORDER BY` 语句
func OrderBy(columns ...string) QueryOption {
	return func(w *queryWrapper) {
		w.orders = columns
	}
}

// Offset 指定 `OFFSET` 语句
func Offset(n int) QueryOption {
	return func(w *queryWrapper) {
		w.offset = n
	}
}

// Limit 指定 `LIMIT` 语句
func Limit(n int) QueryOption {
	return func(w *queryWrapper) {
		w.limit = n
	}
}

// Union 指定 `UNION` 语句
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

// UnionAll 指定 `UNION ALL` 语句
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
