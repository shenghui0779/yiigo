package yiigo

import (
	"errors"
	"fmt"
	"strings"
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
	mysqlReadPool     *pools.ResourcePool
	mysqlWritePool    *pools.ResourcePool
	mysqlReadPoolMux  sync.Mutex
	mysqlWritePoolMux sync.Mutex
)

type Orm struct {
	Db *gorm.DB
}

func (o Orm) Close() {
	err := o.Db.Close()

	if err != nil {
		LogError("mysql connection close error: ", err.Error())
	}
}

func initWriteDb() (*gorm.DB, error) {
	host := GetConfigString("mysql", "writeHost", "localhost")
	port := GetConfigInt("mysql", "writePort", 3306)
	username := GetConfigString("mysql", "username", "root")
	password := GetConfigString("mysql", "password", "root")
	dbname := GetConfigString("mysql", "dbname", "yiicms")
	charset := GetConfigString("mysql", "charset", "utf8mb4")

	address := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=Local", username, password, host, port, dbname, charset)
	writeDb, err := gorm.Open("mysql", address)

	if err != nil {
		LogError("connect mysql error: ", err.Error())
		return nil, err
	}

	writeDb.SingularTable(true)

	debug := GetConfigBool("default", "debug", true)

	if debug {
		writeDb.LogMode(true)
	}

	return writeDb, err
}

func initReadDb() (*gorm.DB, error) {
	host := GetConfigString("mysql", "readHost", "localhost")
	port := GetConfigInt("mysql", "readPort", 3306)
	username := GetConfigString("mysql", "username", "root")
	password := GetConfigString("mysql", "password", "root")
	dbname := GetConfigString("mysql", "dbname", "yiicms")
	charset := GetConfigString("mysql", "charset", "utf8mb4")

	address := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=Local", username, password, host, port, dbname, charset)
	readDb, err := gorm.Open("mysql", address)

	if err != nil {
		LogError("connect mysql error: ", err.Error())
		return nil, err
	}

	readDb.SingularTable(true)

	debug := GetConfigBool("default", "debug", true)

	if debug {
		readDb.LogMode(true)
	}

	return readDb, err
}

func initWriteDbPool() {
	if mysqlWritePool == nil || mysqlWritePool.IsClosed() {
		mysqlWritePoolMux.Lock()
		defer mysqlWritePoolMux.Unlock()

		poolMinActive := GetConfigInt("mysql", "poolMinActive", 100)
		poolMaxActive := GetConfigInt("mysql", "poolMaxActive", 200)
		poolIdleTimeout := GetConfigInt("mysql", "poolIdleTimeout", 2000)

		if mysqlWritePool == nil {
			mysqlWritePool = pools.NewResourcePool(func() (pools.Resource, error) {
				writeDb, err := initWriteDb()
				return Orm{Db: writeDb}, err
			}, poolMinActive, poolMaxActive, time.Duration(poolIdleTimeout)*time.Millisecond)
		}
	}
}

func initReadDbPool() {
	if mysqlReadPool == nil || mysqlReadPool.IsClosed() {
		mysqlReadPoolMux.Lock()
		defer mysqlReadPoolMux.Unlock()

		if mysqlReadPool == nil {
			poolMinActive := GetConfigInt("mysql", "poolMinActive", 100)
			poolMaxActive := GetConfigInt("mysql", "poolMaxActive", 200)
			poolIdleTimeout := GetConfigInt("mysql", "poolIdleTimeout", 2000)

			mysqlReadPool = pools.NewResourcePool(func() (pools.Resource, error) {
				readDb, err := initReadDb()
				return Orm{Db: readDb}, err
			}, poolMinActive, poolMaxActive, time.Duration(poolIdleTimeout)*time.Millisecond)
		}
	}
}

func poolGetWriteDb() (pools.Resource, error) {
	initWriteDbPool()

	if mysqlWritePool == nil {
		LogError("mysql write db pool is null")
		return nil, errors.New("mysql write db pool is null")
	}

	ctx := context.TODO()
	writeOrmResource, err := mysqlWritePool.Get(ctx)

	if err != nil {
		LogError("mysql get write db error: ", err.Error())
		return nil, err
	}

	if writeOrmResource == nil {
		LogError("mysql write pool resource is null")
		return nil, errors.New("mysql write pool resource is null")
	}

	orm := writeOrmResource.(Orm)

	if orm.Db.Error != nil {
		LogError("mysql write resource connection err: ", orm.Db.Error.Error())

		orm.Close()
		//连接断开，重新打开
		db, connErr := initWriteDb()

		if connErr != nil {
			LogError("mysql write db reconnection err: ", connErr.Error())
			return nil, connErr
		} else {
			return Orm{Db: db}, nil
		}
	}

	return writeOrmResource, err
}

func poolGetReadDb() (pools.Resource, error) {
	initReadDbPool()

	if mysqlReadPool == nil {
		LogError("mysql read db pool is null")
		return nil, errors.New("mysql read db pool is null")
	}

	ctx := context.TODO()
	readOrmResource, err := mysqlReadPool.Get(ctx)

	if err != nil {
		LogError("mysql get read db error: ", err.Error())
		return nil, err
	}

	if readOrmResource == nil {
		LogError("mysql read pool resource is null")
		return nil, errors.New("mysql read pool resource is null")
	}

	orm := readOrmResource.(Orm)

	if orm.Db.Error != nil {
		LogError("mysql read resource connection err: ", orm.Db.Error.Error())

		orm.Close()
		//连接断开，重新打开
		db, connErr := initReadDb()

		if connErr != nil {
			LogError("mysql read db reconnection err: ", connErr.Error())
			return nil, connErr
		} else {
			return Orm{Db: db}, nil
		}
	}

	return readOrmResource, err
}

/**
 * insert 插入
 * data 插入数据 interface{} (指针)
 */
func (m *MysqlBase) Insert(data interface{}) error {
	dbResource, err := poolGetWriteDb()

	if err != nil {
		return err
	}

	defer mysqlWritePool.Put(dbResource)

	db := dbResource.(Orm).Db

	if m.TableName == "" {
		tableErr := errors.New("tablename empty")
		LogError("init db error: tablename empty")

		return tableErr
	}

	var table string
	prefix := GetConfigString("mysql", "prefix", "")

	if prefix != "" {
		table = prefix + m.TableName
	} else {
		table = m.TableName
	}

	insertErr := db.Table(table).Create(data).Error

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
func (m *MysqlBase) Update(query map[string]interface{}, data map[string]interface{}) error {
	dbResource, err := poolGetWriteDb()

	if err != nil {
		return err
	}

	defer mysqlWritePool.Put(dbResource)

	db := dbResource.(Orm).Db

	if m.TableName == "" {
		tableErr := errors.New("tablename empty")
		LogError("init db error: tablename empty")

		return tableErr
	}

	var table string
	prefix := GetConfigString("mysql", "prefix", "")

	if prefix != "" {
		table = prefix + m.TableName
	} else {
		table = m.TableName
	}

	db = db.Table(table)

	db = buildQuery(db, query)

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
func (m *MysqlBase) Increment(query map[string]interface{}, column string, inc int) error {
	dbResource, err := poolGetWriteDb()

	if err != nil {
		return err
	}

	defer mysqlWritePool.Put(dbResource)

	db := dbResource.(Orm).Db

	if m.TableName == "" {
		tableErr := errors.New("tablename empty")
		LogError("init db error: tablename empty")

		return tableErr
	}

	var table string
	prefix := GetConfigString("mysql", "prefix", "")

	if prefix != "" {
		table = prefix + m.TableName
	} else {
		table = m.TableName
	}

	db = db.Table(table)

	db = buildQuery(db, query)

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
func (m *MysqlBase) Decrement(query map[string]interface{}, column string, dec int) error {
	dbResource, err := poolGetWriteDb()

	if err != nil {
		return err
	}

	defer mysqlWritePool.Put(dbResource)

	db := dbResource.(Orm).Db

	if m.TableName == "" {
		tableErr := errors.New("tablename empty")
		LogError("init db error: tablename empty")

		return tableErr
	}

	var table string
	prefix := GetConfigString("mysql", "prefix", "")

	if prefix != "" {
		table = prefix + m.TableName
	} else {
		table = m.TableName
	}

	db = db.Table(table)

	db = buildQuery(db, query)

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
func (m *MysqlBase) FindOne(data interface{}, query map[string]interface{}, fields ...[]string) error {
	dbResource, err := poolGetReadDb()

	if err != nil {
		return err
	}

	defer mysqlReadPool.Put(dbResource)

	db := dbResource.(Orm).Db

	if m.TableName == "" {
		tableErr := errors.New("tablename empty")
		LogError("init db error: tablename empty")

		return tableErr
	}

	var table string
	prefix := GetConfigString("mysql", "prefix", "")

	if prefix != "" {
		table = prefix + m.TableName
	} else {
		table = m.TableName
	}

	db = db.Table(table)

	if len(fields) > 0 {
		db = db.Select(fields[0])
	}

	db = buildQuery(db, query)

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
func (m *MysqlBase) Find(data interface{}, query map[string]interface{}, options ...map[string]interface{}) error {
	dbResource, err := poolGetReadDb()

	if err != nil {
		return err
	}

	defer mysqlReadPool.Put(dbResource)

	db := dbResource.(Orm).Db

	if m.TableName == "" {
		tableErr := errors.New("tablename empty")
		LogError("init db error: tablename empty")

		return tableErr
	}

	var table string
	prefix := GetConfigString("mysql", "prefix", "")

	if prefix != "" {
		table = prefix + m.TableName
	} else {
		table = m.TableName
	}

	db = db.Table(table)

	if len(options) > 0 {
		if fields, ok := options[0]["fields"]; ok {
			db = db.Select(fields)
		}

		db = buildQuery(db, query)

		if count, ok := options[0]["count"]; ok {
			db = db.Count(count)
		}

		if group, ok := options[0]["group"]; ok {
			if gro, ok := group.(string); ok {
				db = db.Group(gro)
			}
		}

		if order, ok := options[0]["order"]; ok {
			if ord, ok := order.(string); ok {
				db = db.Order(ord)
			}
		}

		if offset, ok := options[0]["offset"]; ok {
			if off, ok := offset.(int); ok {
				db = db.Offset(off)
			}
		}

		if limit, ok := options[0]["limit"]; ok {
			if lmt, ok := limit.(int); ok {
				db = db.Limit(lmt)
			}
		}
	} else {
		db = buildQuery(db, query)
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
func (m *MysqlBase) FindOneBySql(data interface{}, query map[string]interface{}, bindParams ...interface{}) error {
	dbResource, err := poolGetReadDb()

	if err != nil {
		return err
	}

	defer mysqlReadPool.Put(dbResource)

	db := dbResource.(Orm).Db

	if m.TableName == "" {
		tableErr := errors.New("tablename empty")
		LogError("init db error: tablename empty")

		return tableErr
	}

	var table string
	prefix := GetConfigString("mysql", "prefix", "")

	if prefix != "" {
		table = prefix + m.TableName
	} else {
		table = m.TableName
	}

	db = db.Table(table)

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
func (m *MysqlBase) FindBySql(data interface{}, query map[string]interface{}, bindParams ...interface{}) error {
	dbResource, err := poolGetReadDb()

	if err != nil {
		return err
	}

	defer mysqlReadPool.Put(dbResource)

	db := dbResource.(Orm).Db

	if m.TableName == "" {
		tableErr := errors.New("tablename empty")
		LogError("init db error: tablename empty")

		return tableErr
	}

	var table string
	prefix := GetConfigString("mysql", "prefix", "")

	if prefix != "" {
		table = prefix + m.TableName
	} else {
		table = m.TableName
	}

	db = db.Table(table)

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

	if order, ok := query["order"]; ok {
		if ord, ok := order.(string); ok {
			db = db.Order(ord)
		}
	}

	if offset, ok := query["offset"]; ok {
		if off, ok := offset.(int); ok {
			db = db.Offset(off)
		}
	}

	if limit, ok := query["limit"]; ok {
		if lmt, ok := limit.(int); ok {
			db = db.Limit(lmt)
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

func buildQuery(db *gorm.DB, query map[string]interface{}) *gorm.DB {
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
