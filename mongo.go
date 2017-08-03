package yiigo

import (
	"fmt"
	"reflect"
	"strings"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type Mongo struct {
	DB         string
	Collection string
}

type Sequence struct {
	ID  string `bson:"_id"`
	Seq int    `bson:"seq"`
}

var mongoSession *mgo.Session

/**
 * 初始化mongodb连接
 */
func initMongo() error {
	var err error

	host := GetEnvString("mongo", "host", "localhost")
	port := GetEnvInt("mongo", "port", 27017)
	username := GetEnvString("mongo", "username", "")
	password := GetEnvString("mongo", "password", "")
	poolLimit := GetEnvInt("mongo", "poolLimit", 10)

	dsn := fmt.Sprintf("mongodb://%s:%d", host, port)

	if username != "" {
		dsn = fmt.Sprintf("mongodb://%s:%s@%s:%d", username, password, host, port)
	}

	mongoSession, err = mgo.Dial(dsn)

	if err != nil {
		return err
	}

	mongoSession.SetPoolLimit(poolLimit) //设置连接池大小

	return nil
}

/**
 * 获取连接资源
 * @return *mgo.Session
 */
func (m *Mongo) getSession() *mgo.Session {
	session := mongoSession.Clone()
	return session
}

/**
 * 刷新当前主键(_id)的自增值
 * @param seqs ...int 自增值，默认是：1
 * @return int, error
 */
func (m *Mongo) refreshSequence(seqs ...int) (int, error) {
	session := m.getSession()
	defer session.Close()

	if len(seqs) == 0 {
		seqs = append(seqs, 1)
	}

	condition := bson.M{"_id": m.Collection}

	change := mgo.Change{
		Update:    bson.M{"$inc": bson.M{"seq": seqs[0]}},
		Upsert:    true,
		ReturnNew: true,
	}

	sequence := Sequence{}

	_, err := session.DB(m.DB).C("sequence").Find(condition).Apply(change, &sequence)

	if err != nil {
		return 0, err
	}

	return sequence.Seq, nil
}

/**
 * Insert 新增记录
 * @param data bson.M 插入数据
 * @return int, error 新增记录ID
 */
func (m *Mongo) Insert(data bson.M) (int, error) {
	session := m.getSession()
	defer session.Close()

	id, err := m.refreshSequence()

	if err != nil {
		return 0, err
	}

	data["_id"] = id

	err = session.DB(m.DB).C(m.Collection).Insert(data)

	if err != nil {
		m.refreshSequence(-1)
		return 0, err
	}

	return id, nil
}

/**
 * Update 更新记录
 * @param query bson.M (map[string]interface{}) 查询条件
 * @param data bson.M (map[string]interface{}) 更新字段
 * @return error
 */
func (m *Mongo) Update(query bson.M, data bson.M) error {
	session := m.getSession()
	defer session.Close()

	_, err := session.DB(m.DB).C(m.Collection).UpdateAll(query, bson.M{"$set": data})

	if err != nil {
		return err
	}

	return nil
}

/**
 * Incr 增量 (增/减)
 * @param query bson.M (map[string]interface{}) 查询条件
 * @param column string 自增字段
 * @param inc int 增量
 * @return error
 */
func (m *Mongo) Incr(query bson.M, column string, inc int) error {
	session := m.getSession()
	defer session.Close()

	data := bson.M{column: inc}
	err := session.DB(m.DB).C(m.Collection).Update(query, bson.M{"$inc": data})

	if err != nil && err.Error() != "not found" {
		return err
	}

	return nil
}

/**
 * FindOne 查询单条记录
 * @param query bson.M 查询条件
 * @param dest interface{} (指针) 查询数据
 * @return error
 */
func (m *Mongo) FindOne(query bson.M, dest interface{}) error {
	session := m.getSession()
	defer session.Close()

	err := session.DB(m.DB).C(m.Collection).Find(query).One(dest)

	if err != nil {
		return err
	}

	return nil
}

/**
 * Find 查询多条记录
 * @param query bson.M 查询条件
 * [
 *     condition bson.M
 *     count *int
 *     order string
 *     skip int
 *     limit int
 * ]
 * @param dest interface{} (切片指针) 查询数据
 * @return error
 */
func (m *Mongo) Find(query bson.M, dest interface{}) error {
	session := m.getSession()
	defer session.Close()

	q := session.DB(m.DB).C(m.Collection).Find(query["condition"].(bson.M))

	if count, ok := query["count"]; ok {
		refVal := reflect.ValueOf(count)
		elem := refVal.Elem()

		total, err := q.Count()

		if err != nil {
			return err
		}

		elem.Set(reflect.ValueOf(total))
	}

	if v, ok := query["order"]; ok {
		order := strings.Split(v.(string), ",")
		q = q.Sort(order...)
	}

	if v, ok := query["skip"]; ok {
		q = q.Skip(v.(int))
	}

	if v, ok := query["limit"]; ok {
		q = q.Limit(v.(int))
	}

	err := q.All(dest)

	if err != nil {
		return err
	}

	return nil
}

/**
 * FindAll 查询所有记录
 * @param dest interface{} (切片指针) 查询数据
 * @return error
 */
func (m *Mongo) FindAll(dest interface{}) error {
	session := m.getSession()
	defer session.Close()

	err := session.DB(m.DB).C(m.Collection).Find(bson.M{}).All(dest)

	if err != nil {
		return err
	}

	return nil
}

/**
 * Delete 删除记录
 * @param query bson.M (map[string]interface{}) 查询条件
 * @return error
 */
func (m *Mongo) Delete(query bson.M) error {
	session := m.getSession()
	defer session.Close()

	_, err := session.DB(m.DB).C(m.Collection).RemoveAll(query)

	if err != nil {
		return err
	}

	return nil
}

/**
 * PipeOne 管道聚合操作
 * @param pipeline interface{} 管道条件，如：字段求和
 * []bson.M{
 *     {"$match": match},
 *     {"$group": bson.M{"_id": 1, "total": bson.M{"$sum": field}}},
 * }
 * @param dest interface{} (切片指针) 管道聚合数据
 * @return error
 */
func (m *Mongo) PipeOne(pipeline interface{}, dest interface{}) error {
	session := m.getSession()
	defer session.Close()

	err := session.DB(m.DB).C(m.Collection).Pipe(pipeline).One(dest)

	if err != nil {
		return err
	}

	return nil
}

/**
 * Pipe 管道聚合操作
 * @param pipeline interface{} 管道条件
 * @param dest interface{} (切片指针) 管道聚合数据
 * @return error
 */
func (m *Mongo) Pipe(pipeline interface{}, dest interface{}) error {
	session := m.getSession()
	defer session.Close()

	err := session.DB(m.DB).C(m.Collection).Pipe(pipeline).All(dest)

	if err != nil {
		return err
	}

	return nil
}
