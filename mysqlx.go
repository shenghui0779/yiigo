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

type Mysqlx struct {
	TableName string
}

var (
	mysqlxReadPool  *pools.ResourcePool
	mysqlxWritePool *pools.ResourcePool
	mysqlxPoolMux   sync.Mutex
)

type ResourceDbx struct {
	Db *gorm.DB
}

func (r ResourceDbx) Close() {
	err := r.Db.Close()

	if err != nil {
		LogError("mysql connection close error: ", err.Error())
	}
}

/**
 * 连接数据库
 */
func mysqlxDial(isRead bool) (*gorm.DB, error) {
	var (
		host     string
		port     int
		username string
		password string
	)

	if isRead {
		host = GetConfigString("mysql-s", "host", "localhost")
		port = GetConfigInt("mysql-s", "port", 3306)
		username = GetConfigString("mysql-s", "username", "root")
		password = GetConfigString("mysql-s", "password", "root")
	} else {
		host = GetConfigString("mysql-this", "host", "localhost")
		port = GetConfigInt("mysql-this", "port", 3306)
		username = GetConfigString("mysql-this", "username", "root")
		password = GetConfigString("mysql-this", "password", "root")
	}

	database := GetConfigString("db", "database", "yiicms")
	charset := GetConfigString("db", "charset", "utf8mb4")

	address := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=Local", username, password, host, port, database, charset)
	db, err := gorm.Open("mysql", address)

	if err != nil {
		LogErrorf("connect mysql server %s:%d error: %s", host, port, err.Error())
		return nil, err
	}

	db.SingularTable(true)

	debug := GetConfigBool("default", "debug", true)

	if debug {
		db.LogMode(true)
	}

	return db, err
}

/**
 * 初始化mysql连接池
 */
func initMysqlxPool(isRead bool) {
	mysqlPoolMux.Lock()
	defer mysqlPoolMux.Unlock()

	var (
		poolMinActive   int
		poolMaxActive   int
		poolIdleTimeout int
	)

	if isRead {
		if mysqlxReadPool == nil {
			poolMinActive = GetConfigInt("mysql-s", "poolMinActive", 10)
			poolMaxActive = GetConfigInt("mysql-s", "poolMaxActive", 20)
			poolIdleTimeout = GetConfigInt("mysql-s", "poolIdleTimeout", 10000)

			mysqlxReadPool = pools.NewResourcePool(func() (pools.Resource, error) {
				db, err := mysqlxDial(true)
				return ResourceDbx{Db: db}, err
			}, poolMinActive, poolMaxActive, time.Duration(poolIdleTimeout)*time.Millisecond)
		}
	} else {
		if mysqlxWritePool == nil {
			poolMinActive = GetConfigInt("mysql-this", "poolMinActive", 10)
			poolMaxActive = GetConfigInt("mysql-this", "poolMaxActive", 20)
			poolIdleTimeout = GetConfigInt("mysql-this", "poolIdleTimeout", 10000)

			mysqlxWritePool = pools.NewResourcePool(func() (pools.Resource, error) {
				db, err := mysqlxDial(false)
				return ResourceDbx{Db: db}, err
			}, poolMinActive, poolMaxActive, time.Duration(poolIdleTimeout)*time.Millisecond)
		}
	}
}

/**
 * 获取db资源
 */
func poolGetDbxResource(isRead bool) (pools.Resource, error) {
	var (
		rd  pools.Resource
		err error
	)

	if isRead {
		if mysqlxReadPool == nil || mysqlxReadPool.IsClosed() {
			initMysqlxPool(true)
		}

		if mysqlxReadPool == nil {
			LogError("mysql write db pool is null")
			return nil, errors.New("mysql write db pool is null")
		}

		ctx := context.TODO()
		rd, err = mysqlxReadPool.Get(ctx)
	} else {
		if mysqlxWritePool == nil || mysqlxWritePool.IsClosed() {
			initMysqlxPool(false)
		}

		if mysqlxWritePool == nil {
			LogError("mysql write db pool is null")
			return nil, errors.New("mysql write db pool is null")
		}

		ctx := context.TODO()
		rd, err = mysqlxWritePool.Get(ctx)
	}

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
 */
func (this *Mysqlx) initDb(db *gorm.DB) *gorm.DB {
	var table string
	prefix := GetConfigString("db", "prefix", "")

	if prefix != "" {
		table = prefix + this.TableName
	} else {
		table = this.TableName
	}

	return db.Table(table)
}

/**
 * insert 插入
 * data 插入数据 interface{} (指针)
 */
func (this *Mysqlx) Insert(data interface{}) error {
	rd, err := poolGetDbxResource(false)

	if err != nil {
		return err
	}

	defer mysqlPool.Put(rd)

	db := this.initDb(rd.(ResourceDbx).Db)

	insertErr := db.Create(data).Error

	if insertErr != nil {
		LogErrorf("mysql table %s insert error: %s", this.TableName, insertErr.Error())

		return insertErr
	}

	return nil
}

/**
 * batch insert 批量插入(支持事务)
 * data 插入数据 []interface{} (指针切片)
 */
func (this *Mysqlx) BatchInsert(data []interface{}) error {
	rd, err := poolGetDbxResource(false)

	if err != nil {
		return err
	}

	defer mysqlPool.Put(rd)

	db := this.initDb(rd.(ResourceDbx).Db)

	tx := db.Begin()

	var insertErr error

	for _, model := range data {
		insertErr = tx.Create(model).Error

		if insertErr != nil {
			break
		}
	}

	if insertErr != nil {
		LogErrorf("mysql table %s insert error: %s", this.TableName, insertErr.Error())
		tx.Rollback()

		return insertErr
	}

	tx.Commit()

	return nil
}

/**
 * batchInsertWithAction 带操作的批量插入(支持事务)
 * data 插入数据 []interface{} (指针切片)
 * action 执行数据插入前的操作(支持更新和删除) map[string]interface{}
 * [
 * 		type 操作类型(delete或update)
 * 		query SQL查询where语句 string
 * 		bind SQL语句中 "?" 的绑定值 []interface{}
 * 		data 删除的表Model或需更新的字段 interface{}
 * ]
 */
func (this *Mysqlx) BatchInsertWithAction(data []interface{}, action map[string]interface{}) error {
	rd, err := poolGetDbxResource(false)

	if err != nil {
		return err
	}

	defer mysqlPool.Put(rd)

	db := this.initDb(rd.(ResourceDbx).Db)

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
		LogErrorf("mysql table %s %s error: %s", this.TableName, actionType, dbErr.Error())
		tx.Rollback()

		return dbErr
	}

	for _, model := range data {
		dbErr = tx.Create(model).Error

		if dbErr != nil {
			break
		}
	}

	if dbErr != nil {
		LogErrorf("mysql table %s insert error: %s", this.TableName, dbErr.Error())
		tx.Rollback()

		return dbErr
	}

	tx.Commit()

	return nil
}

/**
 * update 更新
 * query 查询条件 map[string]interface{}
 * [
 * 		where SQL查询where语句 string
 * 		bind SQL语句中 "?" 的绑定值 []interface{}
 * ]
 * data 更新字段 map[string]interface{}
 */
func (this *Mysqlx) Update(query map[string]interface{}, data map[string]interface{}) error {
	rd, err := poolGetDbxResource(false)

	if err != nil {
		return err
	}

	defer mysqlPool.Put(rd)

	db := this.initDb(rd.(ResourceDbx).Db)

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
		LogErrorf("mysql table %s update error: %s", this.TableName, updateErr.Error())

		return updateErr
	}

	return nil
}

/**
 * increment 自增
 * query 查询条件 map[string]interface{}
 * [
 * 		where SQL查询where语句 string
 * 		bind SQL语句中 "?" 的绑定值 []interface{}
 * ]
 * column 自增字段 string
 * inc 增量 int
 */
func (this *Mysqlx) Increment(query map[string]interface{}, column string, inc int) error {
	rd, err := poolGetDbxResource(false)

	if err != nil {
		return err
	}

	defer mysqlPool.Put(rd)

	db := this.initDb(rd.(ResourceDbx).Db)

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
		LogErrorf("mysql table %s inc error: %s", this.TableName, incErr.Error())

		return incErr
	}

	return nil
}

/**
 * decrement 自减
 * query 查询条件 map[string]interface{}
 * [
 * 		where SQL查询where语句 string
 * 		bind SQL语句中 "?" 的绑定值 []interface{}
 * ]
 * column 自减字段 string
 * dec 减量 int
 */
func (this *Mysqlx) Decrement(query map[string]interface{}, column string, dec int) error {
	rd, err := poolGetDbxResource(false)

	if err != nil {
		return err
	}

	defer mysqlPool.Put(rd)

	db := this.initDb(rd.(ResourceDbx).Db)

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
		LogErrorf("mysql table %s dec error: %s", this.TableName, decErr.Error())

		return decErr
	}

	return nil
}

/**
 * findOne 查询
 * data 查询数据 interface{}
 * query 查询条件 map[string]interface{}
 * [
 *      select SQL查询select语句 string
 *      join SQL查询join语句 string
 *      where SQL查询where语句 string
 *      bind SQL语句中 "?" 的绑定值 []interface{}
 * ]
 */
func (this *Mysqlx) FindOne(data interface{}, query map[string]interface{}) error {
	rd, err := poolGetDbxResource(true)

	if err != nil {
		return err
	}

	defer mysqlPool.Put(rd)

	db := this.initDb(rd.(ResourceDbx).Db)

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
			LogErrorf("mysql table %s findone error: %s", this.TableName, errMsg)
		}

		return findErr
	}

	return nil
}

/**
 * find 查询
 * data 查询数据 interface{} (切片指针)
 * query 查询条件 map[string]interface{}
 * [
 *      select SQL查询select语句 string
 *      join SQL查询join语句 string
 *      where SQL查询where语句 string
 *      bind SQL语句中 "?" 的绑定值 []interface{}
 *      count *int
 *      group string
 *      order string
 *      offset int
 *      limit int
 * ]
 */
func (this *Mysqlx) Find(data interface{}, query map[string]interface{}) error {
	rd, err := poolGetDbxResource(true)

	if err != nil {
		return err
	}

	defer mysqlPool.Put(rd)

	db := this.initDb(rd.(ResourceDbx).Db)

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
			LogErrorf("mysql table %s find error: %s", this.TableName, errMsg)
		}

		return findErr
	}

	return nil
}

/**
 * delete 删除
 * data 需要删除的表Model interface{} (struct)
 * query 查询条件 map[string]interface{}
 * [
 * 		where SQL查询where语句 string
 * 		bind SQL语句中 "?" 的绑定值 []interface{}
 * ]
 */
func (this *Mysqlx) Delete(data interface{}, query map[string]interface{}) error {
	rd, err := poolGetDbxResource(false)

	if err != nil {
		return err
	}

	defer mysqlPool.Put(rd)

	db := this.initDb(rd.(ResourceDbx).Db)

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
			LogErrorf("mysql table %s delete error: %s", this.TableName, errMsg)
		}

		return delErr
	}

	return nil
}
