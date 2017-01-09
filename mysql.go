package yiigo

import (
	"fmt"
	"reflect"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
)

type MysqlBase struct {
	DB    string
	Table string
}

var dbmap map[string]*gorm.DB

/**
 * 初始化DB(不指定配置名称，则默认为：db)
 * @param sections ...string 数据库配置名称，如：db1,db2,db3
 */
func InitDB(sections ...string) {
	dbmap = make(map[string]*gorm.DB)

	debug := GetEnvBool("app", "debug", true)

	if len(sections) == 0 {
		sections = append(sections, "db")
	}

	for _, section := range sections {
		var err error

		host := GetEnvString(section, "host", "localhost")
		port := GetEnvInt(section, "port", 3306)
		username := GetEnvString(section, "username", "root")
		password := GetEnvString(section, "password", "")
		maxOpenConns := GetEnvInt(section, "maxOpenConns", 20)
		maxIdleConns := GetEnvInt(section, "maxIdleConns", 10)
		database := GetEnvString(section, "database", "test")
		charset := GetEnvString(section, "charset", "utf8mb4")

		dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=Local", username, password, host, port, database, charset)
		dbmap[section], err = gorm.Open("mysql", dsn)

		if err != nil {
			LogError("connect mysql faild, ", err.Error())
			panic(err)
		}

		dbmap[section].SingularTable(true)

		dbmap[section].DB().SetMaxOpenConns(maxOpenConns)
		dbmap[section].DB().SetMaxIdleConns(maxIdleConns)

		if debug {
			dbmap[section].LogMode(true)
		}
	}
}

/**
 * 获取db
 * @return *gorm.DB, error
 */
func (m *MysqlBase) getDB() (*gorm.DB, error) {
	var dbsection string
	var table string

	if m.DB == "" {
		dbsection = "db"
	}

	db, ok := dbmap[dbsection]

	if !ok {
		LogErrorf("mysql %s is not initialized.", dbsection)
		panic(fmt.Sprintf("mysql %s is not initialized.", dbsection))
	}

	prefix := GetEnvString(dbsection, "prefix", "")

	if prefix != "" {
		table = prefix + m.Table
	} else {
		table = m.Table
	}

	return db.Table(table), nil
}

/**
 * Insert 插入
 * @param data interface{} 插入数据 (struct指针)
 * @return error
 */
func (m *MysqlBase) Insert(data interface{}) error {
	db, err := m.getDB()

	if err != nil {
		return err
	}

	insertErr := db.Create(data).Error

	if insertErr != nil {
		LogErrorf("mysql table %s insert failed, %s", m.Table, insertErr.Error())
		return insertErr
	}

	return nil
}

/**
 * BatchInsert 批量插入 (支持事务)
 * @param data interface{} 插入数据 (struct指针切片)
 * @return error
 */
func (m *MysqlBase) BatchInsert(data interface{}) error {
	refVal := reflect.ValueOf(data)

	if refVal.Kind() != reflect.Slice {
		panic("data must be a pointer slice")
	}

	length := refVal.Len()

	if length == 0 {
		return nil
	}

	db, err := m.getDB()

	if err != nil {
		return err
	}

	tx := db.Begin()

	var insertErr error

	for i := 0; i < length; i++ {
		insertErr = tx.Create(refVal.Index(i).Interface()).Error

		if insertErr != nil {
			break
		}
	}

	if insertErr != nil {
		tx.Rollback()
		LogErrorf("mysql table %s insert failed, %s", m.Table, insertErr.Error())

		return insertErr
	}

	tx.Commit()

	return nil
}

/**
 * BatchInsertWithAction 带操作的批量插入 (支持事务)
 * @param data interface{} 插入数据 (struct指针切片)
 * @param action map[string]interface{} 执行数据插入前的操作 (支持更新和删除)
 * [
 *     type string 操作类型 (delete 或 update)
 *     query string SQL查询 where 语句
 *     bind []interface{} SQL语句中 "?" 的绑定值
 *     data interface{} 删除的 struct 指针或更新的字段
 * ]
 * @return error
 */
func (m *MysqlBase) BatchInsertWithAction(data interface{}, action map[string]interface{}) error {
	refVal := reflect.ValueOf(data)

	if refVal.Kind() != reflect.Slice {
		panic("data must be a pointer slice")
	}

	length := refVal.Len()

	if length == 0 {
		return nil
	}

	db, err := m.getDB()

	if err != nil {
		return err
	}

	tx := db.Begin()

	var (
		dbErr      error
		actionType string
		query      interface{}
		bind       []interface{}
		actionData interface{}
	)

	if v, ok := action["type"]; ok {
		actionType = v.(string)
	}

	if v, ok := action["query"]; ok {
		query = v
	}

	if v, ok := action["bind"]; ok {
		bind = v.([]interface{})
	}

	if v, ok := action["data"]; ok {
		actionData = v
	}

	switch actionType {
	case "update":
		dbErr = tx.Where(query, bind...).Updates(actionData).Error
	case "delete":
		dbErr = tx.Where(query, bind...).Delete(actionData).Error
	}

	if dbErr != nil {
		tx.Rollback()
		LogErrorf("mysql table %s %s failed, %s", m.Table, actionType, dbErr.Error())

		return dbErr
	}

	for i := 0; i < length; i++ {
		dbErr = tx.Create(refVal.Index(i).Interface()).Error

		if dbErr != nil {
			break
		}
	}

	if dbErr != nil {
		tx.Rollback()
		LogErrorf("mysql table %s insert failed, %s", m.Table, dbErr.Error())

		return dbErr
	}

	tx.Commit()

	return nil
}

/**
 * Update 更新
 * @param query map[string]interface{} 查询条件
 * [
 *     where string SQL查询 where 语句
 *     bind []interface{} SQL语句中 "?" 的绑定值
 * ]
 * @param data map[string]interface{} 更新字段
 * @return error
 */
func (m *MysqlBase) Update(query map[string]interface{}, data map[string]interface{}) error {
	db, err := m.getDB()

	if err != nil {
		return err
	}

	var (
		where interface{}
		bind  []interface{}
	)

	if v, ok := query["where"]; ok {
		where = v
	}

	if v, ok := query["bind"]; ok {
		bind = v.([]interface{})
	}

	updateErr := db.Where(where, bind...).Updates(data).Error

	if updateErr != nil {
		LogErrorf("mysql table %s update failed, %s", m.Table, updateErr.Error())
		return updateErr
	}

	return nil
}

/**
 * Increment 自增
 * @param query map[string]interface{} 查询条件
 * [
 *     where string SQL查询 where 语句
 *     bind []interface{} SQL语句中 "?" 的绑定值
 * ]
 * @param column string 自增字段
 * @param inc int 增量
 * @return error
 */
func (m *MysqlBase) Increment(query map[string]interface{}, column string, inc int) error {
	db, err := m.getDB()

	if err != nil {
		return err
	}

	var (
		where interface{}
		bind  []interface{}
	)

	if v, ok := query["where"]; ok {
		where = v
	}

	if v, ok := query["bind"]; ok {
		bind = v.([]interface{})
	}

	expr := fmt.Sprintf("%s + ?", column)
	incErr := db.Where(where, bind...).Update(column, gorm.Expr(expr, inc)).Error

	if incErr != nil {
		LogErrorf("mysql table %s increment failed, %s", m.Table, incErr.Error())
		return incErr
	}

	return nil
}

/**
 * Decrement 自减
 * @param query map[string]interface{} 查询条件
 * [
 *     where string SQL查询 where 语句
 *     bind []interface{} SQL语句中 "?" 的绑定值
 * ]
 * @param column string 自减字段
 * @param dec int 减量
 * @return error
 */
func (m *MysqlBase) Decrement(query map[string]interface{}, column string, dec int) error {
	db, err := m.getDB()

	if err != nil {
		return err
	}

	var (
		where interface{}
		bind  []interface{}
	)

	if v, ok := query["where"]; ok {
		where = v
	}

	if v, ok := query["bind"]; ok {
		bind = v.([]interface{})
	}

	expr := fmt.Sprintf("%s - ?", column)
	decErr := db.Where(where, bind...).Update(column, gorm.Expr(expr, dec)).Error

	if decErr != nil {
		LogErrorf("mysql table %s decrement failed, %s", m.Table, decErr.Error())
		return decErr
	}

	return nil
}

/**
 * FindOne 查询
 * @param query map[string]interface{} 查询条件
 * [
 *     select string SQL查询 select 语句
 *     join string SQL查询 join 语句
 *     where string SQL查询 where 语句
 *     bind []interface{} SQL语句中 "?" 的绑定值
 * ]
 * @param data interface{} 查询数据 (struct指针)
 * @return error
 */
func (m *MysqlBase) FindOne(query map[string]interface{}, data interface{}) error {
	db, err := m.getDB()

	if err != nil {
		return err
	}

	if sel, ok := query["select"]; ok {
		db = db.Select(sel)
	}

	if join, ok := query["join"]; ok {
		if jn, ok := join.(string); ok {
			db = db.Joins(jn)
		}
	}

	var (
		where interface{}
		bind  []interface{}
	)

	if v, ok := query["where"]; ok {
		where = v
	}

	if v, ok := query["bind"]; ok {
		bind = v.([]interface{})
	}

	db = db.Where(where, bind...)

	findErr := db.First(data).Error

	if findErr != nil {
		errMsg := findErr.Error()

		if errMsg != "record not found" {
			LogErrorf("mysql table %s findone failed, %s", m.Table, errMsg)
		}

		return findErr
	}

	return nil
}

/**
 * Find 查询
 * query map[string]interface{} 查询条件
 * [
 *     select string SQL查询 select 语句
 *     join string SQL查询 join 语句
 *     where string SQL查询 where 语句
 *     bind []interface{} SQL语句中 "?" 的绑定值
 *     count *int
 *     group string
 *     order string
 *     offset int
 *     limit int
 * ]
 * data interface{} 查询数据 (struct切片指针)
 * @return error
 */
func (m *MysqlBase) Find(query map[string]interface{}, data interface{}) error {
	db, err := m.getDB()

	if err != nil {
		return err
	}

	if sel, ok := query["select"]; ok {
		db = db.Select(sel)
	}

	if join, ok := query["join"]; ok {
		if jn, ok := join.(string); ok {
			db = db.Joins(jn)
		}
	}

	var (
		where interface{}
		bind  []interface{}
	)

	if v, ok := query["where"]; ok {
		where = v
	}

	if v, ok := query["bind"]; ok {
		bind = v.([]interface{})
	}

	db = db.Where(where, bind...)

	if count, ok := query["count"]; ok {
		db = db.Count(count)
	}

	if group, ok := query["group"]; ok {
		if gro, ok := group.(string); ok {
			db = db.Group(gro)
		}
	}

	if order, ok := query["order"]; ok {
		db = db.Order(order)
	}

	if offset, ok := query["offset"]; ok {
		db = db.Offset(offset)
	}

	if limit, ok := query["limit"]; ok {
		db = db.Limit(limit)
	}

	findErr := db.Find(data).Error

	if findErr != nil {
		errMsg := findErr.Error()

		if errMsg != "record not found" {
			LogErrorf("mysql table %s find failed, %s", m.Table, errMsg)
		}

		return findErr
	}

	return nil
}

/**
 * Delete 删除
 * @param query map[string]interface{} 查询条件
 * [
 *     where string SQL查询 where 语句
 *     bind []interface{} SQL语句中 "?" 的绑定值
 * ]
 * @param data interface{} (struct指针)
 * @return error
 */
func (m *MysqlBase) Delete(query map[string]interface{}, data interface{}) error {
	db, err := m.getDB()

	if err != nil {
		return err
	}

	var (
		where interface{}
		bind  []interface{}
	)

	if v, ok := query["where"]; ok {
		where = v
	}

	if v, ok := query["bind"]; ok {
		bind = v.([]interface{})
	}

	delErr := db.Where(where, bind...).Delete(data).Error

	if delErr != nil {
		errMsg := delErr.Error()

		if errMsg != "record not found" {
			LogErrorf("mysql table %s delete failed, %s", m.Table, errMsg)
		}

		return delErr
	}

	return nil
}
