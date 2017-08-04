package yiigo

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

type MySQL struct {
	Connection  string
	MasterSlave bool
	Table       string
}

var dbmap map[string]*sqlx.DB

// SQL expression
type expr struct {
	expr string
	args []interface{}
}

/**
 * 初始化DB
 */
func initMySQL() error {
	dbmap = make(map[string]*sqlx.DB)
	sections := childSections("mysql")

	for _, v := range sections {
		database := v.Key("database").MustString("test")
		charset := v.Key("charset").MustString("utf8mb4")
		collection := v.Key("collection").MustString("utf8mb4_general_ci")

		maxOpenConns := v.Key("maxOpenConns").MustInt(20)
		maxIdleConns := v.Key("maxIdleConns").MustInt(10)

		// 是否配置主从
		childs := v.ChildSections()

		if len(childs) == 0 {
			host := v.Key("host").MustString("localhost")
			port := v.Key("post").MustInt(3306)
			username := v.Key("username").MustString("root")
			password := v.Key("password").MustString("")

			dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&collation=%s&parseTime=True&loc=Local", username, password, host, port, database, charset, collection)
			db, err := sqlx.Open("mysql", dsn)

			if err != nil {
				db.Close()
				return err
			}

			db.SetMaxOpenConns(maxOpenConns)
			db.SetMaxIdleConns(maxIdleConns)

			err = db.Ping()

			if err != nil {
				db.Close()
				return err
			}

			dbmap[v.Name()] = db
		} else {
			for _, c := range childs {
				host := c.Key("host").MustString("localhost")
				port := c.Key("post").MustInt(3306)
				username := c.Key("username").MustString("root")
				password := c.Key("password").MustString("")

				dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&collation=%s&parseTime=True&loc=Local", username, password, host, port, database, charset, collection)
				db, err := sqlx.Open("mysql", dsn)

				if err != nil {
					db.Close()
					return err
				}

				db.SetMaxOpenConns(maxOpenConns)
				db.SetMaxIdleConns(maxIdleConns)

				err = db.Ping()

				if err != nil {
					db.Close()
					return err
				}

				dbmap[c.Name()] = db
			}
		}
	}

	return nil
}

/**
 * 获取db
 * @return *sqlx.DB
 */
func (m *MySQL) getDB(read bool) (*sqlx.DB, error) {
	connection := "default"

	if m.Connection != "" {
		connection = m.Connection
	}

	schema := fmt.Sprintf("mysql.%s", connection)

	if m.MasterSlave {
		if read {
			schema = fmt.Sprintf("mysql.%s.read", schema)
		} else {
			schema = fmt.Sprintf("mysql.%s.write", schema)
		}
	}

	db, ok := dbmap[schema]

	if !ok {
		return nil, fmt.Errorf("database %s is not connected", schema)
	}

	return db, nil
}

/**
 * 获取表前缀
 * @return string
 */
func (m *MySQL) getPrefix() string {
	connection := "default"

	if m.Connection != "" {
		connection = m.Connection
	}

	prefix := GetEnvString(fmt.Sprintf("mysql.%s", connection), "prefix", "")

	return prefix
}

/**
 * Insert 插入
 * @param data X 插入数据
 * @return int64, error 新增记录ID
 */
func (m *MySQL) Insert(data X) (int64, error) {
	db, err := m.getDB(false)

	if err != nil {
		return 0, err
	}

	sql, binds := m.buildInsert(data)
	result, err := db.Exec(sql, binds...)

	if err != nil {
		return 0, fmt.Errorf("%v, SQL: %s, args: %v", err, sql, binds)
	}

	id, err := result.LastInsertId()

	if err != nil {
		return 0, fmt.Errorf("%v, SQL: %s, args: %v", err, sql, binds)
	}

	return id, nil
}

/**
 * BatchInsert 批量插入
 * @param columns []string 插入的字段
 * @param data []X 插入数据
 * @return int64, error 影响的行数
 */
func (m *MySQL) BatchInsert(columns []string, data []X) (int64, error) {
	db, err := m.getDB(false)

	if err != nil {
		return 0, err
	}

	sql, binds := m.buildBatchInsert(columns, data)
	result, err := db.Exec(sql, binds...)

	if err != nil {
		return 0, fmt.Errorf("%v, SQL: %s, args: %v", err, sql, binds)
	}

	rows, err := result.RowsAffected()

	if err != nil {
		return 0, fmt.Errorf("%v, SQL: %s, args: %v", err, sql, binds)
	}

	return rows, nil
}

/**
 * Update 更新
 * @param query X 查询条件
 * yiigo.X{
 *     where string WHERE条件语句
 *     binds []interface{} WHERE语句中 "?" 的绑定值
 * }
 * @param data X 更新字段
 * @return int64, error 影响的行数
 */
func (m *MySQL) Update(query X, data X) (int64, error) {
	db, err := m.getDB(false)

	if err != nil {
		return 0, err
	}

	sql, binds := m.buildUpdate(query, data)
	_sql, args, err := sqlx.In(sql, binds...)

	if err != nil {
		return 0, fmt.Errorf("%v, SQL: %s, args: %v", err, sql, binds)
	}

	result, err := db.Exec(_sql, args...)

	if err != nil {
		return 0, fmt.Errorf("%v, SQL: %s, args: %v", err, _sql, args)
	}

	rows, err := result.RowsAffected()

	if err != nil {
		return 0, fmt.Errorf("%v, SQL: %s, args: %v", err, _sql, args)
	}

	return rows, nil
}

/**
 * Count 获取记录数
 * @param query X 查询条件
 * yiigo.X{
 *     where string WHERE语句
 *     binds []interface{} WHERE语句中 "?" 的绑定值
 * }
 * @param columns ...string 聚合字段，默认为：*
 * @return error
 */
func (m *MySQL) Count(query X, columns ...string) (int, error) {
	db, err := m.getDB(true)

	if err != nil {
		return 0, err
	}

	if len(columns) > 0 {
		query["select"] = fmt.Sprintf("COUNT(%s)", columns[0])
	} else {
		query["select"] = "COUNT(*)"
	}

	sql, binds := m.buildQuery(query)
	_sql, args, err := sqlx.In(sql, binds...)

	if err != nil {
		return 0, fmt.Errorf("%v, SQL: %s, args: %v", err, sql, binds)
	}

	count := 0
	err = db.Get(&count, _sql, args...)

	if err != nil {
		return 0, fmt.Errorf("%v, SQL: %s, args: %v", err, sql, binds)
	}

	return count, nil
}

/**
 * FindOne 查询单条记录
 * @param query X 查询条件
 * yiigo.X{
 *     select string SELECT语句
 *     join string JOIN语句
 *     where string WHERE语句
 *     binds []interface{} WHERE语句中 "?" 的绑定值
 * }
 * @param dest interface{} 查询数据 (struct指针)
 * @return error
 */
func (m *MySQL) FindOne(query X, dest interface{}) error {
	db, err := m.getDB(true)

	if err != nil {
		return err
	}

	query["limit"] = 1

	sql, binds := m.buildQuery(query)
	_sql, args, err := sqlx.In(sql, binds...)

	if err != nil {
		return fmt.Errorf("%v, SQL: %s, args: %v", err, sql, binds)
	}

	err = db.Get(dest, _sql, args...)

	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			return errors.New("not found")
		}

		return fmt.Errorf("%v, SQL: %s, args: %v", err, _sql, args)
	}

	return nil
}

/**
 * Find 查询多条记录
 * @param query X 查询条件
 * yiigo.X{
 *     select string SELECT语句
 *     join []string JOIN语句
 *     where string WHERE语句
 *     group string GROUP BY语句
 *     order string ORDER BY语句
 *     limit int LIMIT语句
 *     offset int OFFSET语句
 *     binds []interface{} WHERE语句中 "?" 的绑定值
 * }
 * @param dest interface{} 查询数据 (struct切片指针)
 * @return error
 */
func (m *MySQL) Find(query X, dest interface{}) error {
	db, err := m.getDB(true)

	if err != nil {
		return err
	}

	sql, binds := m.buildQuery(query)
	_sql, args, err := sqlx.In(sql, binds...)

	if err != nil {
		return fmt.Errorf("%v, SQL: %s, args: %v", err, sql, binds)
	}

	err = db.Select(dest, _sql, args...)

	if err != nil {
		return fmt.Errorf("%v, SQL: %s, args: %v", err, _sql, args)
	}

	return nil
}

/**
 * FindAll 查询所有记录
 * @param dest interface{} 查询数据 (struct切片指针)
 * @param columns ...string 查询字段
 * @return error
 */
func (m *MySQL) FindAll(dest interface{}, columns ...string) error {
	db, err := m.getDB(true)

	if err != nil {
		return err
	}

	query := X{}

	if len(columns) > 0 {
		query["select"] = strings.Join(columns, ",")
	}

	sql, binds := m.buildQuery(query)
	err = db.Select(dest, sql, binds...)

	if err != nil {
		return fmt.Errorf("%v, SQL: %s, args: %v", err, sql, binds)
	}

	return nil
}

/**
 * FindBySQL SQL查询
 * @param sql string SQL查询语句
 * @parms binds []interface{} SQL绑定值
 * @param dest interface{} 查询数据 (struct指针或struct切片指针)
 * @return error
 */
func (m *MySQL) FindBySQL(sql string, binds []interface{}, dest interface{}) error {
	db, err := m.getDB(true)

	if err != nil {
		return err
	}

	_sql, args, err := sqlx.In(sql, binds...)

	if err != nil {
		return fmt.Errorf("%v, SQL: %s, args: %v", err, sql, binds)
	}

	v := reflect.Indirect(reflect.ValueOf(dest))

	if v.Kind() == reflect.Slice {
		err = db.Select(dest, _sql, args...)
	} else {
		err = db.Get(dest, _sql, args...)
	}

	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			return errors.New("not found")
		}

		return fmt.Errorf("%v, SQL: %s, args: %v", err, _sql, args)
	}

	return nil
}

/**
 * Delete 删除
 * @param query X 查询条件
 * yiigo.X{
 *     where string WHERE语句
 *     binds []interface{} WHERE语句中 "?" 的绑定值
 * }
 * @return int64, error 影响的行数
 */
func (m *MySQL) Delete(query X) (int64, error) {
	db, err := m.getDB(false)

	if err != nil {
		return 0, err
	}

	sql, binds := m.buildDelete(query)
	_sql, args, err := sqlx.In(sql, binds...)

	if err != nil {
		return 0, fmt.Errorf("%v, SQL: %s, args: %v", err, sql, binds)
	}

	result, err := db.Exec(_sql, args...)

	if err != nil {
		return 0, fmt.Errorf("%v, SQL: %s, args: %v", err, _sql, args)
	}

	rows, err := result.RowsAffected()

	if err != nil {
		return 0, fmt.Errorf("%v, SQL: %s, args: %v", err, _sql, args)
	}

	return rows, nil
}

/**
 * DoTransactions 事务处理
 * @param operations X 操作集合
 * [
 *     yiigo.X{
 *		   "type": "insert"
 *	 	   "table": string,
 *		   "data": yiigo.X,
 *     },
 *     yiigo.X{
 *		   "type": "batchInsert"
 *	 	   "table": string,
 *		   "columns": []string
 *		   "data": []yiigo.X,
 *     },
 *     yiigo.X{
 *	       "type": "update"
 *	       "query": yiigo.X{
 *	           "table": string,
 * 			   "where": string,
 *			   "binds": []interface{},
 *         },
 *		   "data": yiigo.X,
 *     },
 *	   yiigo.X{
 *		   "type": "delete"
 *	 	   "query": yiigo.X{
 *	 	   	   "table": string,
 * 			   "where": string,
 *			   "binds": []interface{},
 *         },
 *     },
 * ]
 * @return error
 */
func (m *MySQL) DoTransactions(operations []X) error {
	db, err := m.getDB(false)

	if err != nil {
		return err
	}

	tx, err := db.Begin()

	if err != nil {
		return err
	}

	errSQL := ""
	errArgs := []interface{}{}

	for _, opt := range operations {
		optType := ""

		if v, ok := opt["type"]; ok {
			optType = v.(string)
		} else {
			continue
		}

		switch optType {
		case "insert":
			table := []string{}
			data := X{}

			if v, ok := opt["table"]; ok {
				table = append(table, v.(string))
			}

			if v, ok := opt["data"]; ok {
				data = v.(X)
			}

			sql, binds := m.buildInsert(data, table...)
			_, err = tx.Exec(sql, binds...)

			if err != nil {
				errSQL = sql
				errArgs = binds

				break
			}
		case "batchInsert":
			table := []string{}
			columns := []string{}
			data := []X{}

			if v, ok := opt["table"]; ok {
				table = append(table, v.(string))
			}

			if v, ok := opt["columns"]; ok {
				columns = v.([]string)
			}

			if v, ok := opt["data"]; ok {
				data = v.([]X)
			}

			sql, binds := m.buildBatchInsert(columns, data, table...)
			_, err = tx.Exec(sql, binds...)

			if err != nil {
				errSQL = sql
				errArgs = binds

				break
			}
		case "update":
			query := X{}
			data := X{}

			if v, ok := opt["query"]; ok {
				query = v.(X)
			}

			if v, ok := opt["data"]; ok {
				data = v.(X)
			}

			sql, binds := m.buildUpdate(query, data)
			_sql, args, _ := sqlx.In(sql, binds...)
			_, err = tx.Exec(_sql, args...)

			if err != nil {
				errSQL = _sql
				errArgs = args

				break
			}
		case "delete":
			query := X{}

			if v, ok := opt["query"]; ok {
				query = v.(X)
			}

			sql, binds := m.buildDelete(query)
			_sql, args, _ := sqlx.In(sql, binds...)
			_, err = tx.Exec(_sql, args...)

			if err != nil {
				errSQL = _sql
				errArgs = args

				break
			}
		}

		if err != nil {
			break
		}
	}

	if err != nil {
		tx.Rollback()

		return fmt.Errorf("%v, SQL: %s, args: %v", err, errSQL, errArgs)
	}

	tx.Commit()

	return nil
}

/**
 * Expr SQL表达式
 * @param expression string 表达式，如：yiigo.Expr("price * ? + ?", 2, 100)
 * @param args ...interface{} 表达式中 "?" 的绑定值
 * @return *expr
 */
func Expr(expression string, args ...interface{}) *expr {
	return &expr{expr: expression, args: args}
}

/**
 * buildInsert 构建插入SQL
 * @param data X 插入数据
 * @param tables ...string 插入表
 * @return string, []interface{}
 */
func (m *MySQL) buildInsert(data X, tables ...string) (string, []interface{}) {
	if len(tables) == 0 {
		tables = append(tables, m.Table)
	}

	prefix := m.getPrefix()

	columns := []string{}
	placeholders := []string{}
	binds := []interface{}{}

	for k, v := range data {
		columns = append(columns, k)
		placeholders = append(placeholders, "?")
		binds = append(binds, v)
	}

	sql := fmt.Sprintf("INSERT INTO %s%s (%s) VALUES (%s)", prefix, tables[0], strings.Join(columns, ","), strings.Join(placeholders, ","))

	return sql, binds
}

/**
 * buildBatchInsert 构建批量插入SQL
 * @param columns []string 插入字段
 * @param data []X 插入数据
 * @param tables ...string 插入表
 * @return string, []interface{}
 */
func (m *MySQL) buildBatchInsert(columns []string, data []X, tables ...string) (string, []interface{}) {
	if len(tables) == 0 {
		tables = append(tables, m.Table)
	}

	prefix := m.getPrefix()

	placeholders := []string{}
	binds := []interface{}{}

	for _, v := range data {
		bindvars := []string{}

		for _, column := range columns {
			binds = append(binds, v[column])
			bindvars = append(bindvars, "?")
		}

		placeholders = append(placeholders, fmt.Sprintf("(%s)", strings.Join(bindvars, ",")))
	}

	sql := fmt.Sprintf("INSERT INTO %s%s (%s) VALUES %s", prefix, tables[0], strings.Join(columns, ","), strings.Join(placeholders, ","))

	return sql, binds
}

/**
 * buildUpdate 构建更新SQL
 * @param query X 查询条件
 * @param data X 更新数据
 * @return string, []interface{}
 */
func (m *MySQL) buildUpdate(query X, data X) (string, []interface{}) {
	table := m.Table
	prefix := m.getPrefix()

	clauses := []string{}
	sets := []string{}
	binds := []interface{}{}

	if v, ok := query["table"]; ok {
		table = v.(string)
	}

	clauses = append(clauses, fmt.Sprintf("UPDATE %s%s", prefix, table))

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

	if v, ok := query["where"]; ok {
		clauses = append(clauses, fmt.Sprintf("WHERE %s", v.(string)))
	}

	if v, ok := query["binds"]; ok {
		binds = append(binds, v.([]interface{})...)
	}

	sql := strings.Join(clauses, " ")

	return sql, binds
}

/**
 * buildQuery 构建查询SQL
 * @param query X 查询条件
 * @return string, []interface{}
 */
func (m *MySQL) buildQuery(query X) (string, []interface{}) {
	table := m.Table
	prefix := m.getPrefix()

	clauses := []string{}
	binds := []interface{}{}

	if v, ok := query["select"]; ok {
		clauses = append(clauses, fmt.Sprintf("SELECT %s", v.(string)))
	} else {
		clauses = append(clauses, "SELECT *")
	}

	if v, ok := query["table"]; ok {
		table = v.(string)
	}

	if v, ok := query["join"]; ok {
		clauses = append(clauses, fmt.Sprintf("FROM %s%s AS a", prefix, table))

		for _, join := range v.([]string) {
			clauses = append(clauses, join)
		}
	} else {
		clauses = append(clauses, fmt.Sprintf("FROM %s%s", prefix, table))
	}

	if v, ok := query["where"]; ok {
		clauses = append(clauses, fmt.Sprintf("WHERE %s", v.(string)))
	}

	if v, ok := query["group"]; ok {
		clauses = append(clauses, fmt.Sprintf("GROUP BY %s", v.(string)))
	}

	if v, ok := query["order"]; ok {
		clauses = append(clauses, fmt.Sprintf("ORDER BY %s", v.(string)))
	}

	if v, ok := query["limit"]; ok {
		clauses = append(clauses, fmt.Sprintf("LIMIT %d", v.(int)))
	}

	if v, ok := query["offset"]; ok {
		clauses = append(clauses, fmt.Sprintf("OFFSET %d", v.(int)))
	}

	if v, ok := query["binds"]; ok {
		binds = append(binds, v.([]interface{})...)
	}

	sql := strings.Join(clauses, " ")

	return sql, binds
}

/**
 * buildDelete 构建删除SQL
 * @param query X 查询条件
 * @return string, []interface{}
 */
func (m *MySQL) buildDelete(query X) (string, []interface{}) {
	table := m.Table
	prefix := m.getPrefix()

	clauses := []string{}
	binds := []interface{}{}

	if v, ok := query["table"]; ok {
		table = v.(string)
	}

	clauses = append(clauses, fmt.Sprintf("DELETE FROM %s%s", prefix, table))

	if v, ok := query["where"]; ok {
		clauses = append(clauses, fmt.Sprintf("WHERE %s", v.(string)))
	}

	if v, ok := query["binds"]; ok {
		binds = append(binds, v.([]interface{})...)
	}

	sql := strings.Join(clauses, " ")

	return sql, binds
}
