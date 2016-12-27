package yiigo

import (
	"fmt"
	"reflect"
	"sync"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
)

type MysqlxBase struct {
	TableName string
}

var (
	readDb     *gorm.DB
	writeDb    *gorm.DB
	readDbMux  sync.Mutex
	writeDbMux sync.Mutex
)

/**
 * 初始化readDb
 * @return error
 */
func initReadDb() error {
	readDbMux.Lock()
	defer readDbMux.Unlock()

	if readDb == nil {
		var err error

		host := GetEnvString("sdb", "host", "localhost")
		port := GetEnvInt("sdb", "port", 3306)
		username := GetEnvString("sdb", "username", "root")
		password := GetEnvString("sdb", "password", "")
		maxOpenConns := GetEnvInt("sdb", "maxOpenConns", 10)
		maxIdleConns := GetEnvInt("sdb", "maxIdleConns", 10)
		database := GetEnvString("sdb", "database", "test")
		charset := GetEnvString("sdb", "charset", "utf8mb4")

		address := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=Local", username, password, host, port, database, charset)
		readDb, err = gorm.Open("mysql", address)

		if err != nil {
			LogErrorf("connect mysql server %s:%d error: %s", host, port, err.Error())
			return err
		}

		readDb.SingularTable(true)

		readDb.DB().SetMaxOpenConns(maxOpenConns)
		readDb.DB().SetMaxIdleConns(maxIdleConns)

		debug := GetEnvBool("app", "debug", true)

		if debug {
			readDb.LogMode(true)
		}
	}

	return nil
}

/**
 * 初始化writeDb
 * @return error
 */
func initWriteDb() error {
	writeDbMux.Lock()
	defer writeDbMux.Unlock()

	if writeDb == nil {
		var err error

		host := GetEnvString("mdb", "host", "localhost")
		port := GetEnvInt("mdb", "port", 3306)
		username := GetEnvString("mdb", "username", "root")
		password := GetEnvString("mdb", "password", "")
		maxOpenConns := GetEnvInt("mdb", "maxOpenConns", 10)
		maxIdleConns := GetEnvInt("mdb", "maxIdleConns", 10)
		database := GetEnvString("mdb", "database", "test")
		charset := GetEnvString("mdb", "charset", "utf8mb4")

		address := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=Local", username, password, host, port, database, charset)
		writeDb, err = gorm.Open("mysql", address)

		if err != nil {
			LogErrorf("connect mysql server %s:%d error: %s", host, port, err.Error())
			return err
		}

		writeDb.SingularTable(true)

		writeDb.DB().SetMaxOpenConns(maxOpenConns)
		writeDb.DB().SetMaxIdleConns(maxIdleConns)

		debug := GetEnvBool("app", "debug", true)

		if debug {
			writeDb.LogMode(true)
		}
	}

	return nil
}

/**
 * 获取readDb
 * @return *gorm.DB, error
 */
func (m *MysqlxBase) getReadDb() (*gorm.DB, error) {
	if readDb == nil {
		err := initReadDb()

		if err != nil {
			return nil, err
		}
	}

	var table string
	prefix := GetEnvString("sdb", "prefix", "")

	if prefix != "" {
		table = prefix + m.TableName
	} else {
		table = m.TableName
	}

	return readDb.Table(table), nil
}

/**
 * 获取writeDb
 * @return *gorm.DB, error
 */
func (m *MysqlxBase) getWriteDb() (*gorm.DB, error) {
	if writeDb == nil {
		err := initWriteDb()

		if err != nil {
			return nil, err
		}
	}

	var table string
	prefix := GetEnvString("mdb", "prefix", "")

	if prefix != "" {
		table = prefix + m.TableName
	} else {
		table = m.TableName
	}

	return writeDb.Table(table), nil
}

/**
 * Insert 插入
 * @param data interface{} 插入数据 (struct指针)
 * @return error
 */
func (m *MysqlxBase) Insert(data interface{}) error {
	writeDb, err := m.getWriteDb()

	if err != nil {
		return err
	}

	insertErr := writeDb.Create(data).Error

	if insertErr != nil {
		LogErrorf("mysql table %s insert error: %s", m.TableName, insertErr.Error())
		return insertErr
	}

	return nil
}

/**
 * BatchInsert 批量插入 (支持事务)
 * @param data interface{} 插入数据 (struct指针切片)
 * @return error
 */
func (m *MysqlxBase) BatchInsert(data interface{}) error {
	refVal := reflect.ValueOf(data)

	if refVal.Kind() != reflect.Slice {
		panic("data must be a pointer slice")
	}

	length := refVal.Len()

	if length == 0 {
		return nil
	}

	writeDb, err := m.getWriteDb()

	if err != nil {
		return err
	}

	tx := writeDb.Begin()

	var insertErr error

	for i := 0; i < length; i++ {
		insertErr = tx.Create(refVal.Index(i).Interface()).Error

		if insertErr != nil {
			break
		}
	}

	if insertErr != nil {
		tx.Rollback()
		LogErrorf("mysql table %s insert error: %s", m.TableName, insertErr.Error())

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
 * 		type string 操作类型 (delete 或 update)
 * 		query string SQL查询where语句
 * 		bind []interface{} SQL语句中 "?" 的绑定值
 * 		data interface{} 删除的struct指针或更新的字段
 * ]
 * @return error
 */
func (m *MysqlxBase) BatchInsertWithAction(data interface{}, action map[string]interface{}) error {
	refVal := reflect.ValueOf(data)

	if refVal.Kind() != reflect.Slice {
		panic("data must be a pointer slice")
	}

	length := refVal.Len()

	if length == 0 {
		return nil
	}

	writeDb, err := m.getWriteDb()

	if err != nil {
		return err
	}

	tx := writeDb.Begin()

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
		LogErrorf("mysql table %s %s error: %s", m.TableName, actionType, dbErr.Error())

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
		LogErrorf("mysql table %s insert error: %s", m.TableName, dbErr.Error())

		return dbErr
	}

	tx.Commit()

	return nil
}

/**
 * Update 更新
 * @param query map[string]interface{} 查询条件
 * [
 * 		where string SQL查询where语句
 * 		bind []interface{} SQL语句中 "?" 的绑定值
 * ]
 * @param data map[string]interface{} 更新字段
 * @return error
 */
func (m *MysqlxBase) Update(query map[string]interface{}, data map[string]interface{}) error {
	writeDb, err := m.getWriteDb()

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

	updateErr := writeDb.Where(where, bind...).Updates(data).Error

	if updateErr != nil {
		LogErrorf("mysql table %s update error: %s", m.TableName, updateErr.Error())
		return updateErr
	}

	return nil
}

/**
 * Increment 自增
 * @param query map[string]interface{} 查询条件
 * [
 * 		where string SQL查询where语句
 * 		bind []interface{} SQL语句中 "?" 的绑定值
 * ]
 * @param column string 自增字段
 * @param inc int 增量
 * @return error
 */
func (m *MysqlxBase) Increment(query map[string]interface{}, column string, inc int) error {
	writeDb, err := m.getWriteDb()

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
	incErr := writeDb.Where(where, bind...).Update(column, gorm.Expr(expr, inc)).Error

	if incErr != nil {
		LogErrorf("mysql table %s inc error: %s", m.TableName, incErr.Error())
		return incErr
	}

	return nil
}

/**
 * Decrement 自减
 * @param query map[string]interface{} 查询条件
 * [
 * 		where string SQL查询where语句
 * 		bind []interface{} SQL语句中 "?" 的绑定值
 * ]
 * @param column string 自减字段
 * @param dec int 减量
 * @return error
 */
func (m *MysqlxBase) Decrement(query map[string]interface{}, column string, dec int) error {
	writeDb, err := m.getWriteDb()

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
	decErr := writeDb.Where(where, bind...).Update(column, gorm.Expr(expr, dec)).Error

	if decErr != nil {
		LogErrorf("mysql table %s dec error: %s", m.TableName, decErr.Error())
		return decErr
	}

	return nil
}

/**
 * FindOne 查询
 * @param query map[string]interface{} 查询条件
 * [
 *      select string SQL查询select语句
 *      join string SQL查询join语句
 *      where string SQL查询where语句
 *      bind []interface{} SQL语句中 "?" 的绑定值
 * ]
 * @param data interface{} 查询数据 (struct指针)
 * @return error
 */
func (m *MysqlxBase) FindOne(query map[string]interface{}, data interface{}) error {
	readDb, err := m.getReadDb()

	if err != nil {
		return err
	}

	if sel, ok := query["select"]; ok {
		readDb = readDb.Select(sel)
	}

	if join, ok := query["join"]; ok {
		if jn, ok := join.(string); ok {
			readDb = readDb.Joins(jn)
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

	readDb = readDb.Where(where, bind...)

	findErr := readDb.First(data).Error

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
 * Find 查询
 * query map[string]interface{} 查询条件
 * [
 *      select string SQL查询select语句
 *      join string SQL查询join语句
 *      where string SQL查询where语句
 *      bind []interface{} SQL语句中 "?" 的绑定值
 *      count *int
 *      group string
 *      order string
 *      offset int
 *      limit int
 * ]
 * data interface{} 查询数据 (struct切片指针)
 * @return error
 */
func (m *MysqlxBase) Find(query map[string]interface{}, data interface{}) error {
	readDb, err := m.getReadDb()

	if err != nil {
		return err
	}

	if sel, ok := query["select"]; ok {
		readDb = readDb.Select(sel)
	}

	if join, ok := query["join"]; ok {
		if jn, ok := join.(string); ok {
			readDb = readDb.Joins(jn)
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

	readDb = readDb.Where(where, bind...)

	if count, ok := query["count"]; ok {
		readDb = readDb.Count(count)
	}

	if group, ok := query["group"]; ok {
		if gro, ok := group.(string); ok {
			readDb = readDb.Group(gro)
		}
	}

	if order, ok := query["order"]; ok {
		readDb = readDb.Order(order)
	}

	if offset, ok := query["offset"]; ok {
		readDb = readDb.Offset(offset)
	}

	if limit, ok := query["limit"]; ok {
		readDb = readDb.Limit(limit)
	}

	findErr := readDb.Find(data).Error

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
 * Delete 删除
 * @param query map[string]interface{} 查询条件
 * [
 * 		where string SQL查询where语句
 * 		bind []interface{} SQL语句中 "?" 的绑定值
 * ]
 * @param data interface{} (struct指针)
 * @return error
 */
func (m *MysqlxBase) Delete(query map[string]interface{}, data interface{}) error {
	readDb, err := m.getReadDb()

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

	delErr := readDb.Where(where, bind...).Delete(data).Error

	if delErr != nil {
		errMsg := delErr.Error()

		if errMsg != "record not found" {
			LogErrorf("mysql table %s delete error: %s", m.TableName, errMsg)
		}

		return delErr
	}

	return nil
}
