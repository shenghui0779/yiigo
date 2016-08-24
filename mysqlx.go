package yiigo

import (
	"errors"
	"fmt"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
)

type MysqlBaseX struct {
	TableName string
}

func initMysql(isRead bool) (*gorm.DB, error) {
	var (
		host     string
		port     int
		username string
		password string
		dbname   string
		charset  string
	)

	debug := GetConfigBool("default", "debug", true)

	if isRead {
		host = GetConfigString("mysql", "host", "localhost")
		port = GetConfigInt("mysql", "port", 3306)
		username = GetConfigString("mysql", "username", "root")
		password = GetConfigString("mysql", "password", "root")
		dbname = GetConfigString("mysql", "dbname", "yiicms")
		charset = GetConfigString("mysql", "charset", "utf8mb4")
	} else {
		host = GetConfigString("mysql", "host", "localhost")
		port = GetConfigInt("mysql", "port", 3306)
		username = GetConfigString("mysql", "username", "root")
		password = GetConfigString("mysql", "password", "root")
		dbname = GetConfigString("mysql", "dbname", "yiicms")
		charset = GetConfigString("mysql", "charset", "utf8mb4")
	}

	address := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=Local", username, password, host, port, dbname, charset)

	db, err := gorm.Open("mysql", address)

	if err != nil {
		LogError("connect mysql error: ", err.Error())
	}

	db.SingularTable(true)

	if debug {
		db.LogMode(true)
	}

	return db, err
}

func (m *MysqlBaseX) GetReadDb() (*gorm.DB, error) {
	return m.getDb(true)
}

func (m *MysqlBaseX) GetWriteDb() (*gorm.DB, error) {
	return m.getDb(false)
}

func (m *MysqlBaseX) getDb(isRead bool) (*gorm.DB, error) {
	db, err := initMysql(isRead)

	if err != nil {
		LogError("init mysql error: ", err.Error())

		return db, err
	}

	if m.TableName == "" {
		tableErr := errors.New("tablename empty")
		LogError("init db error: tablename empty")

		return db, tableErr
	}

	var table string
	prefix := GetConfigString("mysql", "prefix", "")

	if prefix != "" {
		table = prefix + m.TableName
	} else {
		table = m.TableName
	}

	return db.Table(table), nil
}

/**
 * insert 插入
 * data 插入数据 interface{} (指针)
 */
func (m *MysqlBaseX) Insert(data interface{}) error {

	db, err := m.GetWriteDb()

	if err != nil {
		return err
	}

	defer db.Close()

	insertErr := db.Create(data).Error

	if insertErr != nil {
		LogErrorf("mysql table %s insert error: %s", m.TableName, insertErr.Error())

		return insertErr
	}

	return nil
}

/**
 * update 更新
 * query 查询条件 map[string]interface{}
 * data 更新字段 map[string]interface{}
 */
func (m *MysqlBaseX) Update(query map[string]interface{}, data map[string]interface{}) error {

	db, err := m.GetWriteDb()

	if err != nil {
		return err
	}

	defer db.Close()

	db = buildQueryX(db, query)

	updateErr := db.Updates(data).Error

	if updateErr != nil {
		LogErrorf("mysql table %s update error: %s", m.TableName, updateErr.Error())

		return updateErr
	}

	return nil
}

/**
 * increment 自增
 * query 查询条件 map[string]interface{}
 * column 自增字段 string
 * inc 增量 int
 */
func (m *MysqlBaseX) Increment(query map[string]interface{}, column string, inc int) error {

	db, err := m.GetWriteDb()

	if err != nil {
		return err
	}

	defer db.Close()

	db = buildQueryX(db, query)

	expr := fmt.Sprintf("%s + ?", column)
	incErr := db.Update(column, gorm.Expr(expr, inc)).Error

	if incErr != nil {
		LogErrorf("mysql table %s inc error: %s", m.TableName, incErr.Error())

		return incErr
	}

	return nil
}

/**
 * decrement 自减
 * query 查询条件 map[string]interface{}
 * column 自减字段 string
 * dec 减量 int
 */
func (m *MysqlBaseX) Decrement(query map[string]interface{}, column string, dec int) error {

	db, err := m.GetWriteDb()

	if err != nil {
		return err
	}

	defer db.Close()

	db = buildQueryX(db, query)

	expr := fmt.Sprintf("%s - ?", column)
	decErr := db.Update(column, gorm.Expr(expr, dec)).Error

	if decErr != nil {
		LogErrorf("mysql table %s dec error: %s", m.TableName, decErr.Error())

		return decErr
	}

	return nil
}

/**
 * findOne 查询
 * data 查询数据 interface{} (指针)
 * query 查询条件 map[string]interface{}
 * fields 查询的字段 []string
 */
func (m *MysqlBaseX) FindOne(data interface{}, query map[string]interface{}, fields ...[]string) error {

	db, err := m.GetReadDb()

	if err != nil {
		return err
	}

	defer db.Close()

	if len(fields) > 0 {
		db = db.Select(fields[0])
	}

	db = buildQueryX(db, query)

	findErr := db.First(data).Error

	if findErr != nil {
		errMsg := findErr.Error()

		if errMsg != "record not found" {
			LogErrorf("mysql table %s findone error: %s", m.TableName, errMsg)
		}

		return findErr
	}

	return nil
}

/**
 * find 查询
 * data 查询数据 interface{} (切片指针)
 * query 查询条件 map[string]interface{}
 * options map[string]interface{}
 * [
 *      fields 查询的字段 []string
 *      count *int
 *      group string
 *      order string
 *      offset int
 *      limit int
 * ]
 */
func (m *MysqlBaseX) Find(data interface{}, query map[string]interface{}, options ...map[string]interface{}) error {

	db, err := m.GetReadDb()

	if err != nil {
		return err
	}

	defer db.Close()

	if len(options) > 0 {
		if fields, ok := options[0]["fields"]; ok {
			db = db.Select(fields)
		}

		db = buildQueryX(db, query)

		if count, ok := options[0]["count"]; ok {
			db = db.Count(count)
		}

		if group, ok := options[0]["group"]; ok {
			if gro, ok := group.(string); ok {
				db = db.Group(gro)
			}
		}

		if ord, ok := options[0]["order"]; ok {
			if order, ok := ord.(string); ok {
				db = db.Order(order)
			}
		}

		if off, ok := options[0]["offset"]; ok {
			if offset, ok := off.(int); ok {
				db = db.Offset(offset)
			}
		}

		if lmt, ok := options[0]["limit"]; ok {
			if limit, ok := lmt.(int); ok {
				db = db.Limit(limit)
			}
		}
	} else {
		db = buildQueryX(db, query)
	}

	findErr := db.Find(data).Error

	if findErr != nil {
		errMsg := findErr.Error()

		if errMsg != "record not found" {
			LogErrorf("mysql table %s find error: %s", m.TableName, errMsg)
		}

		return findErr
	}

	return nil
}

/**
 * findOneBySql 查询
 * data 查询数据 interface{}
 * query 查询条件 map[string]interface{}
 * [
 *      select SQL查询select语句 string
 *      join SQL查询join语句 string
 *      where SQL查询where语句 string
 * ]
 * bindParams where语句中 "?" 绑定的值
 */
func (m *MysqlBaseX) FindOneBySql(data interface{}, query map[string]interface{}, bindParams ...interface{}) error {
	db, err := m.GetReadDb()

	if err != nil {
		return err
	}

	defer db.Close()

	if sel, ok := query["select"]; ok {
		db = db.Select(sel)
	}

	if join, ok := query["join"]; ok {
		if jn, ok := join.(string); ok {
			db = db.Joins(jn)
		}
	}

	if where, ok := query["where"]; ok {
		db = db.Where(where, bindParams...)
	}

	findErr := db.First(data).Error

	if findErr != nil {
		errMsg := findErr.Error()

		if errMsg != "record not found" {
			LogErrorf("mysql table %s findone error: %s", m.TableName, errMsg)
		}

		return findErr
	}

	return nil
}

/**
 * findBySql 查询
 * data 查询数据 interface{} (切片指针)
 * query 查询条件 map[string]interface{}
 * [
 *      select SQL查询select语句 string
 *      join SQL查询join语句 string
 *      where SQL查询where语句 string
 *      count *int
 *      group string
 *      order string
 *      offset int
 *      limit int
 * ]
 * bindParams where语句中 "?" 绑定的值
 */
func (m *MysqlBaseX) FindBySql(data interface{}, query map[string]interface{}, bindParams ...interface{}) error {
	db, err := m.GetReadDb()

	if err != nil {
		return err
	}

	defer db.Close()

	if sel, ok := query["select"]; ok {
		db = db.Select(sel)
	}

	if join, ok := query["join"]; ok {
		if jn, ok := join.(string); ok {
			db = db.Joins(jn)
		}
	}

	if where, ok := query["where"]; ok {
		db = db.Where(where, bindParams...)
	}

	if count, ok := query["count"]; ok {
		db = db.Count(count)
	}

	if group, ok := query["group"]; ok {
		if gro, ok := group.(string); ok {
			db = db.Joins(gro)
		}
	}

	if ord, ok := query["order"]; ok {
		if order, ok := ord.(string); ok {
			db = db.Order(order)
		}
	}

	if off, ok := query["offset"]; ok {
		if offset, ok := off.(int); ok {
			db = db.Offset(offset)
		}
	}

	if lmt, ok := query["limit"]; ok {
		if limit, ok := lmt.(int); ok {
			db = db.Limit(limit)
		}
	}

	findErr := db.Find(data).Error

	if findErr != nil {
		errMsg := findErr.Error()

		if errMsg != "record not found" {
			LogErrorf("mysql table %s find error: %s", m.TableName, errMsg)
		}

		return findErr
	}

	return nil
}

func buildQueryX(db *gorm.DB, query map[string]interface{}) *gorm.DB {
	if len(query) > 0 {
		for key, value := range query {
			tmp := strings.Split(key, ":")

			if len(tmp) == 2 {
				switch tmp[1] {
				case "eq":
					query := fmt.Sprintf("%s = ?", tmp[0])
					db = db.Where(query, value)
				case "ne":
					query := fmt.Sprintf("%s <> ?", tmp[0])
					db = db.Where(query, value)
				case "ge":
					query := fmt.Sprintf("%s >= ?", tmp[0])
					db = db.Where(query, value)
				case "gt":
					query := fmt.Sprintf("%s > ?", tmp[0])
					db = db.Where(query, value)
				case "le":
					query := fmt.Sprintf("%s <= ?", tmp[0])
					db = db.Where(query, value)
				case "lt":
					query := fmt.Sprintf("%s < ?", tmp[0])
					db = db.Where(query, value)
				case "lk":
					if str, ok := value.(string); ok {
						value = fmt.Sprintf("%%%s%%", str)
						query := fmt.Sprintf("%s LIKE ?", tmp[0])
						db = db.Where(query, value)
					}
				case "in":
					query := fmt.Sprintf("%s IN (?)", tmp[0])
					db = db.Where(query, value)
				case "ni":
					query := fmt.Sprintf("%s NOT IN (?)", tmp[0])
					db = db.Where(query, value)
				case "fi":
					query := fmt.Sprintf("FIND_IN_SET(?, %s)", tmp[0])
					db = db.Where(query, value)
				}
			} else {
				query := fmt.Sprintf("%s = ?", tmp[0])
				db = db.Where(query, value)
			}
		}
	}

	return db
}
