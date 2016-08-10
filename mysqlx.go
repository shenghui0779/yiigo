package yiigo

import (
	"errors"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
	"strings"
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
 * data 插入数据 (interface{} 指针)
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
 * query 查询条件 (map[string]interface{})
 * data 更新字段 (map[string]interface{})
 */
func (m *MysqlBaseX) Update(query map[string]interface{}, data map[string]interface{}) error {

	db, err := m.GetWriteDb()

	if err != nil {
		return err
	}

	defer db.Close()

	db = formatQueryX(db, query)

	updateErr := db.Updates(data).Error

	if updateErr != nil {
		LogErrorf("mysql table %s update error: %s", m.TableName, updateErr.Error())

		return updateErr
	}

	return nil
}

/**
 * increment 自增
 * query 查询条件 (map[string]interface{})
 * column 自增字段 (string)
 * inc 增量 (int)
 */
func (m *MysqlBaseX) Increment(query map[string]interface{}, column string, inc int) error {

	db, err := m.GetWriteDb()

	if err != nil {
		return err
	}

	defer db.Close()

	db = formatQueryX(db, query)

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
 * query 查询条件 (map[string]interface{})
 * column 自减字段 (string)
 * dec 减量 (int)
 */
func (m *MysqlBaseX) Decrement(query map[string]interface{}, column string, dec int) error {

	db, err := m.GetWriteDb()

	if err != nil {
		return err
	}

	defer db.Close()

	db = formatQueryX(db, query)

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
 * query 查询条件 (map[string]interface{})
 * data 查询数据 (interface{})
 * fields 查询的字段 ([]string)
 */
func (m *MysqlBaseX) FindOne(query map[string]interface{}, data interface{}, fields ...[]string) error {

	db, err := m.GetReadDb()

	if err != nil {
		return err
	}

	defer db.Close()

	if len(fields) > 0 {
		db = db.Select(fields[0])
	}

	db = formatQueryX(db, query)

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
 * query 查询条件 (map[string]interface{})
 * data 查询数据 (interface{})
 * options (map[string]interface{}) [
 *      fields 查询的字段 ([]string)
 *      count (*int)
 *      order (string)
 *      offset (int)
 *      limit (int)
 * ]
 */
func (m *MysqlBaseX) Find(query map[string]interface{}, data interface{}, options ...map[string]interface{}) error {

	db, err := m.GetReadDb()

	if err != nil {
		return err
	}

	defer db.Close()

	if len(options) > 0 {
		if fields, ok := options[0]["fields"]; ok {
			db = db.Select(fields)
		}

		db = formatQueryX(db, query)

		if count, ok := options[0]["count"]; ok {
			db = db.Count(count)
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
		db = formatQueryX(db, query)
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
 * query 查询条件 (map[string]interface{}) [
 *      sql SQL查询语句 (string)
 *      fields 查询的字段 ([]string)
 * ]
 * data 查询数据 (interface{})
 * bindParams SQL语句中 "?" 绑定的值
 */
func (m *MysqlBaseX) FindOneBySql(query map[string]interface{}, data interface{}, bindParams ...interface{}) error {
	db, err := m.GetReadDb()

	if err != nil {
		return err
	}

	defer db.Close()

	if fields, ok := query["fields"]; ok {
		db = db.Select(fields)
	}

	if sql, ok := query["sql"]; ok {
		db = db.Where(sql, bindParams...)
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
 * query 查询条件 (map[string]interface{}) [
 *      sql SQL查询语句 (string)
 *      fields 查询的字段 ([]string)
 *      count (*int)
 *      order (string)
 *      offset (int)
 *      limit (int)
 * ]
 * data 查询数据 (interface{})
 * bindParams SQL语句中 "?" 绑定的值
 */
func (m *MysqlBaseX) FindBySql(query map[string]interface{}, data interface{}, bindParams ...interface{}) error {
	db, err := m.GetReadDb()

	if err != nil {
		return err
	}

	defer db.Close()

	if fields, ok := query["fields"]; ok {
		db = db.Select(fields)
	}

	if sql, ok := query["sql"]; ok {
		db = db.Where(sql, bindParams...)
	}

	if count, ok := query["count"]; ok {
		db = db.Count(count)
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

func formatQueryX(db *gorm.DB, query map[string]interface{}) *gorm.DB {
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
					query := fmt.Sprintf("%s in (?)", tmp[0])
					db = db.Where(query, value)
				case "ni":
					db = db.Not(tmp[0], value)
				case "fi":
					query := fmt.Sprintf("find_in_set(?, %s)", tmp[0])
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
