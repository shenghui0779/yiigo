package yiigo

import (
	"fmt"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

var dbmap map[string]*sqlx.DB

// SQL expression
type expr struct {
	expr string
	args []interface{}
}

type grammar struct {
	table   string
	columns []string
	clauses []string
	binds   []interface{}
}

type MySQL struct {
	db      *sqlx.DB
	tx      *sqlx.Tx
	grammar *grammar
	sql     string
}

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

		db.SetMaxOpenConns(EnvInt(k, "maxOpenConns", 20))
		db.SetMaxIdleConns(EnvInt(k, "maxIdleConns", 10))

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
func DB(connection ...string) (*MySQL, error) {
	conn := "default"

	if len(connection) > 0 {
		conn = connection[0]
	}

	db, ok := dbmap[fmt.Sprintf("mysql.%s", conn)]

	if !ok {
		return nil, fmt.Errorf("database %s is not connected", conn)
	}

	return &MySQL{
		db:      db,
		grammar: &grammar{},
	}, nil
}

// Expr build sql expression, eg: yiigo.Expr("price * ? + ?", 2, 100)
func Expr(expression string, args ...interface{}) *expr {
	return &expr{expr: expression, args: args}
}

func (m *MySQL) Table(table string) *MySQL {
	m.grammar.table = table
	return m
}

func (m *MySQL) Select(columns ...string) *MySQL {
	m.grammar.columns = columns
	return m
}

func (m *MySQL) InnerJoin(table string, on string) *MySQL {
	clause := fmt.Sprintf("INNER JOIN %s ON %s", table, on)
	m.grammar.clauses = append(m.grammar.clauses, clause)

	return m
}

func (m *MySQL) LeftJoin(table string, on string) *MySQL {
	clause := fmt.Sprintf("LEFT JOIN %s ON %s", table, on)
	m.grammar.clauses = append(m.grammar.clauses, clause)

	return m
}

func (m *MySQL) RightJoin(table string, on string) *MySQL {
	clause := fmt.Sprintf("RIGHT JOIN %s ON %s", table, on)
	m.grammar.clauses = append(m.grammar.clauses, clause)

	return m
}

func (m *MySQL) Where(where string) *MySQL {
	clause := fmt.Sprintf("WHERE %s", where)
	m.grammar.clauses = append(m.grammar.clauses, clause)

	return m
}

func (m *MySQL) GroupBy(group string) *MySQL {
	clause := fmt.Sprintf("GROUP BY %s", group)
	m.grammar.clauses = append(m.grammar.clauses, clause)

	return m
}

func (m *MySQL) Having(having string) *MySQL {
	clause := fmt.Sprintf("HAVING %s", having)
	m.grammar.clauses = append(m.grammar.clauses, clause)

	return m
}

func (m *MySQL) OrderBy(order string) *MySQL {
	clause := fmt.Sprintf("ORDER BY %s", order)
	m.grammar.clauses = append(m.grammar.clauses, clause)

	return m
}

func (m *MySQL) Limit(limit int) *MySQL {
	clause := fmt.Sprintf("LIMIT %d", limit)
	m.grammar.clauses = append(m.grammar.clauses, clause)

	return m
}

func (m *MySQL) Offset(offset int) *MySQL {
	join := fmt.Sprintf("OFFSET %d", offset)
	m.grammar.clauses = append(m.grammar.clauses, join)

	return m
}

func (m *MySQL) Insert(data X) {
	defer m.reset(false)

	columns := []string{}
	placeholders := []string{}
	binds := []interface{}{}

	for k, v := range data {
		columns = append(columns, k)
		placeholders = append(placeholders, "?")
		binds = append(binds, v)
	}

	sql := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)", m.table, strings.Join(columns, ","), strings.Join(placeholders, ","))

	result, err := m.db.Exec(sql, binds...)

	if err != nil {
		m.Error = fmt.Errorf("%v, SQL: %s, Args: %v", err, sql, binds)
		return
	}

	id, err := result.LastInsertId()

	if err != nil {
		return 0, fmt.Errorf("%v, SQL: %s, args: %v", err, sql, binds)
	}

	return id, nil
}

// BuildInsert build insert sql
func BuildInsert(table string, data X) (string, []interface{}) {
	columns := []string{}
	placeholders := []string{}
	binds := []interface{}{}

	for k, v := range data {
		columns = append(columns, k)
		placeholders = append(placeholders, "?")
		binds = append(binds, v)
	}

	sql := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)", table, strings.Join(columns, ","), strings.Join(placeholders, ","))

	return sql, binds
}

// BuildBatchInsert build batch insert sql
func BuildBatchInsert(table string, columns []string, data X) (string, []interface{}) {
	placeholders := []string{}
	binds := []interface{}{}

	for _, v := range data {
		bindvars := []string{}

		for _, c := range columns {
			binds = append(binds, v[c])
			bindvars = append(bindvars, "?")
		}

		placeholders = append(placeholders, fmt.Sprintf("(%s)", strings.Join(bindvars, ",")))
	}

	sql := fmt.Sprintf("INSERT INTO %s (%s) VALUES %s", table, strings.Join(columns, ","), strings.Join(placeholders, ","))

	return sql, binds
}

// BuildUpdate build update sql
func BuildUpdate(table string, data X, condition ...interface{}) (string, []interface{}) {
	clauses := []string{}
	sets := []string{}
	binds := []interface{}{}

	clauses = append(clauses, fmt.Sprintf("UPDATE %s", table))

	for k, v := range data {
		if expr, ok := v.(*expr); ok {
			sets = append(sets, fmt.Sprintf("%s = %s", k, expr.expr))
			binds = append(binds, expr.args...)
		} else {
			sets = append(sets, fmt.Sprintf("%s = ?", k))
			binds = append(binds, v)
		}
	}

	clauses = append(clauses, fmt.Sprintf("SET %s", strings.Join(sets, ",")))

	length := len(condition)

	if length > 0 {
		clauses = append(clauses, fmt.Sprintf("WHERE %s", condition[0].(string)))
	}

	if length > 1 {
		binds = append(binds, condition[1:]...)
	}

	sql := strings.Join(clauses, " ")

	return sql, binds
}

func (m *MySQL) reset(tx bool) {
	if tx {
		m.tx = nil
	}

	m.grammar = &grammar{}
	m.sql = ""
}
