package yiigo

import (
	"errors"
	"fmt"
	"sync"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
	"github.com/youtube/vitess/go/pools"
	"golang.org/x/net/context"
)

type MysqlBase struct {
	TableName string
}

var (
	mysqlPool    *pools.ResourcePool
	mysqlPoolMux sync.Mutex
)

type ResourceDb struct {
	Db *gorm.DB
}

/**
 * 关闭连接
 */
func (r ResourceDb) Close() {
	err := r.Db.Close()

	if err != nil {
		LogError("mysql connection close error: ", err.Error())
	}
}

/**
 * 连接数据库
 * @return *gorm.DB, error
 */
func mysqlDial() (*gorm.DB, error) {
	var (
		host     string
		port     int
		username string
		password string
	)

	host = GetEnvString("mysql", "host", "localhost")
	port = GetEnvInt("mysql", "port", 3306)
	username = GetEnvString("mysql", "username", "root")
	password = GetEnvString("mysql", "password", "root")

	database := GetEnvString("db", "database", "yiicms")
	charset := GetEnvString("db", "charset", "utf8mb4")

	address := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=Local", username, password, host, port, database, charset)
	db, err := gorm.Open("mysql", address)

	if err != nil {
		LogErrorf("connect mysql server %s:%d error: %s", host, port, err.Error())
		return nil, err
	}

	db.SingularTable(true)

	debug := GetEnvBool("app", "debug", true)

	if debug {
		db.LogMode(true)
	}

	return db, err
}

/**
 * 初始化mysql连接池
 */
func initMysqlPool() {
	mysqlPoolMux.Lock()
	defer mysqlPoolMux.Unlock()

	var (
		poolMinActive   int
		poolMaxActive   int
		poolIdleTimeout int
	)

	if mysqlPool == nil {
		poolMinActive = GetEnvInt("mysql", "poolMinActive", 10)
		poolMaxActive = GetEnvInt("mysql", "poolMaxActive", 20)
		poolIdleTimeout = GetEnvInt("mysql", "poolIdleTimeout", 10000)

		mysqlPool = pools.NewResourcePool(func() (pools.Resource, error) {
			db, err := mysqlDial()
			return ResourceDb{Db: db}, err
		}, poolMinActive, poolMaxActive, time.Duration(poolIdleTimeout)*time.Millisecond)
	}
}

/**
 * 获取db资源
 * @return pools.Resource, error
 */
func poolGetDbResource() (pools.Resource, error) {
	if mysqlPool == nil || mysqlPool.IsClosed() {
		initMysqlPool()
	}

	if mysqlPool == nil {
		LogError("mysql pool is null")
		return nil, errors.New("mysql pool is null")
	}

	ctx := context.TODO()
	rd, err := mysqlPool.Get(ctx)

	if err != nil {
		return nil, err
	}

	if rd == nil {
		LogError("mysql pool resource is null")
		return nil, errors.New("mysql pool resource is null")
	}

	return rd, err
}

/**
 * 初始化db
 * @param db *gorm.DB
 * @return *gorm.DB
 */
func (m *MysqlBase) initDb(db *gorm.DB) *gorm.DB {
	var table string
	prefix := GetEnvString("db", "prefix", "")

	if prefix != "" {
		table = prefix + m.TableName
	} else {
		table = m.TableName
	}

	return db.Table(table)
}

/**
 * Insert 插入
 * @param data interface{} 插入数据 (struct指针)
 * @return error
 */
func (m *MysqlBase) Insert(data interface{}) error {
	rd, err := poolGetDbResource()

	if err != nil {
		return err
	}

	defer mysqlPool.Put(rd)

	db := m.initDb(rd.(ResourceDb).Db)

	insertErr := db.Create(data).Error

	if insertErr != nil {
		LogErrorf("mysql table %s insert error: %s", m.TableName, insertErr.Error())
		return insertErr
	}

	return nil
}

/**
 * BatchInsert 批量插入 (支持事务)
 * @param data []interface{} 插入数据 (struct指针切片)
 * @return error
 */
func (m *MysqlBase) BatchInsert(data []interface{}) error {
	rd, err := poolGetDbResource()

	if err != nil {
		return err
	}

	defer mysqlPool.Put(rd)

	db := m.initDb(rd.(ResourceDb).Db)

	tx := db.Begin()

	var insertErr error

	for _, model := range data {
		insertErr = tx.Create(model).Error

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
 * @param data []interface{} 插入数据 (struct指针切片)
 * @param action map[string]interface{} 执行数据插入前的操作 (支持更新和删除)
 * [
 * 		type string 操作类型 (delete 或 update)
 * 		query string SQL查询where语句
 * 		bind []interface{} SQL语句中 "?" 的绑定值
 * 		data interface{} 删除的struct指针或更新的字段
 * ]
 * @return error
 */
func (m *MysqlBase) BatchInsertWithAction(data []interface{}, action map[string]interface{}) error {
	rd, err := poolGetDbResource()

	if err != nil {
		return err
	}

	defer mysqlPool.Put(rd)

	db := m.initDb(rd.(ResourceDb).Db)

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
		LogErrorf("mysql table %s %s error: %s", m.TableName, actionType, dbErr.Error())

		return dbErr
	}

	for _, model := range data {
		dbErr = tx.Create(model).Error

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
func (m *MysqlBase) Update(query map[string]interface{}, data map[string]interface{}) error {
	rd, err := poolGetDbResource()

	if err != nil {
		return err
	}

	defer mysqlPool.Put(rd)

	db := m.initDb(rd.(ResourceDb).Db)

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
func (m *MysqlBase) Increment(query map[string]interface{}, column string, inc int) error {
	rd, err := poolGetDbResource()

	if err != nil {
		return err
	}

	defer mysqlPool.Put(rd)

	db := m.initDb(rd.(ResourceDb).Db)

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
func (m *MysqlBase) Decrement(query map[string]interface{}, column string, dec int) error {
	rd, err := poolGetDbResource()

	if err != nil {
		return err
	}

	defer mysqlPool.Put(rd)

	db := m.initDb(rd.(ResourceDb).Db)

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
func (m *MysqlBase) FindOne(query map[string]interface{}, data interface{}) error {
	rd, err := poolGetDbResource()

	if err != nil {
		return err
	}

	defer mysqlPool.Put(rd)

	db := m.initDb(rd.(ResourceDb).Db)

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
func (m *MysqlBase) Find(query map[string]interface{}, data interface{}) error {
	rd, err := poolGetDbResource()

	if err != nil {
		return err
	}

	defer mysqlPool.Put(rd)

	db := m.initDb(rd.(ResourceDb).Db)

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
func (m *MysqlBase) Delete(query map[string]interface{}, data interface{}) error {
	rd, err := poolGetDbResource()

	if err != nil {
		return err
	}

	defer mysqlPool.Put(rd)

	db := m.initDb(rd.(ResourceDb).Db)

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
			LogErrorf("mysql table %s delete error: %s", m.TableName, errMsg)
		}

		return delErr
	}

	return nil
}
