package yiigo

import (
	"fmt"
	"reflect"
	"strings"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type MongoBase struct {
	Collection string
}

type Sequence struct {
	Id  string `bson:"_id"`
	Seq int64  `bson:"seq"`
}

var mongoSession *mgo.Session

/**
 * 初始化mongodb连接
 */
func InitMongo() {
	var err error

	host := GetEnvString("mongo", "host", "localhost")
	port := GetEnvInt("mongo", "port", 27017)
	poolLimit := GetEnvInt("mongo", "poolLimit", 10)

	dsn := fmt.Sprintf("mongodb://%s:%d", host, port)

	mongoSession, err = mgo.Dial(dsn)

	if err != nil {
		LogError("connect to mongo server failed, ", err.Error())
		panic(err)
	}

	mongoSession.SetPoolLimit(poolLimit) //设置连接池大小
}

/**
 * 获取连接资源
 * @return *mgo.Session, string, error
 */
func getSession() (*mgo.Session, string, error) {
	session := mongoSession.Clone()
	db := GetEnvString("mongo", "database", "test")

	return session, db, nil
}

/**
 * 刷新当前主键(_id)的自增值
 * @return int64, error
 */
func (m *MongoBase) refreshSequence() (int64, error) {
	session, db, err := getSession()

	if err != nil {
		return 0, err
	}

	defer session.Close()

	c := session.DB(db).C("sequence")

	condition := bson.M{"_id": m.Collection}

	change := mgo.Change{
		Update:    bson.M{"$inc": bson.M{"seq": 1}},
		Upsert:    true,
		ReturnNew: true,
	}

	sequence := Sequence{}

	_, applyErr := c.Find(condition).Apply(change, &sequence)

	if applyErr != nil {
		LogError("mongo refresh sequence error: ", applyErr.Error())
		return 0, applyErr
	}

	return sequence.Seq, nil
}

/**
 * Insert 新增记录
 * @param data interface{} 插入数据 (struct指针)
 * @return error
 */
func (m *MongoBase) Insert(data interface{}) error {
	session, db, err := getSession()

	if err != nil {
		return err
	}

	defer session.Close()

	refVal := reflect.ValueOf(data)
	elem := refVal.Elem()

	if elem.Kind() != reflect.Struct {
		LogErrorf("cannot use (type %v) as mongo Insert param", elem.Type())
		return fmt.Errorf("cannot use (type %v) as mongo Insert param", elem.Type())
	}

	id, seqErr := m.refreshSequence()

	if seqErr != nil {
		return seqErr
	}

	elem.Field(0).SetInt(id)

	c := session.DB(db).C(m.Collection)

	insertErr := c.Insert(data)

	if insertErr != nil {
		LogErrorf("mongo collection %s insert error: %s", m.Collection, insertErr.Error())
		return insertErr
	}

	return nil
}

/**
 * Update 更新记录
 * @param query bson.M (map[string]interface{}) 查询条件
 * @param data bson.M (map[string]interface{}) 更新字段
 * @return error
 */
func (m *MongoBase) Update(query bson.M, data bson.M) error {
	session, db, err := getSession()

	if err != nil {
		return err
	}

	defer session.Close()

	c := session.DB(db).C(m.Collection)

	updateErr := c.Update(query, bson.M{"$set": data})

	if updateErr != nil {
		if updateErr.Error() != "not found" {
			LogErrorf("mongo collection %s update error: %s", m.Collection, updateErr.Error())
		}

		return updateErr
	}

	return nil
}

/**
 * Increment 自增
 * @param query bson.M (map[string]interface{}) 查询条件
 * @param column string 自增字段
 * @param inc int 增量
 * @return error
 */
func (m *MongoBase) Increment(query bson.M, column string, incr int) error {
	session, db, err := getSession()

	if err != nil {
		return err
	}

	defer session.Close()

	c := session.DB(db).C(m.Collection)

	data := bson.M{column: incr}
	updateErr := c.Update(query, bson.M{"$inc": data})

	if updateErr != nil {
		if updateErr.Error() != "not found" {
			LogErrorf("mongo collection %s update error: %s", m.Collection, updateErr.Error())
		}

		return updateErr
	}

	return nil
}

/**
 * FindOne 查询
 * @param data interface{} (指针) 查询数据
 * @param query bson.M (map[string]interface{}) 查询条件
 * @return error
 */
func (m *MongoBase) FindOne(data interface{}, query bson.M) error {
	session, db, err := getSession()

	if err != nil {
		return err
	}

	defer session.Close()

	c := session.DB(db).C(m.Collection)

	findErr := c.Find(query).One(data)

	if findErr != nil {
		if findErr.Error() != "not found" {
			LogErrorf("mongo collection %s findone error: %s", m.Collection, findErr.Error())
		}

		return findErr
	}

	return nil
}

/**
 * Find 查询
 * @param data interface{} (切片指针) 查询数据
 * @param query map[string]interface{} 查询条件
 * [
 * 		condition bson.M
 *      count *int
 *      order string
 *      skip int
 *      limit int
 * ]
 * @return error
 */
func (m *MongoBase) Find(data interface{}, query map[string]interface{}) error {
	session, db, err := getSession()

	if err != nil {
		return err
	}

	defer session.Close()

	q := session.DB(db).C(m.Collection).Find(query["condition"].(bson.M))

	if count, ok := query["count"]; ok {
		refVal := reflect.ValueOf(count)
		elem := refVal.Elem()

		total, countErr := q.Count()

		if countErr != nil {
			LogError("mongo collection %s count error: %s", m.Collection, countErr.Error())
			elem.Set(reflect.ValueOf(0))
		} else {
			elem.Set(reflect.ValueOf(total))
		}
	}

	if order, ok := query["order"]; ok {
		if ordStr, ok := order.(string); ok {
			ordArr := strings.Split(ordStr, ",")
			q = q.Sort(ordArr...)
		}
	}

	if skip, ok := query["skip"]; ok {
		if skp, ok := skip.(int); ok {
			q = q.Skip(skp)
		}
	}

	if limit, ok := query["limit"]; ok {
		if lmt, ok := limit.(int); ok {
			q = q.Limit(lmt)
		}
	}

	findErr := q.All(data)

	if findErr != nil {
		if findErr.Error() != "not found" {
			LogErrorf("mongo collection %s find error: %s", m.Collection, findErr.Error())
		}

		return findErr
	}

	return nil
}

/**
 * Delete 删除记录
 * @param query bson.M (map[string]interface{}) 查询条件
 * @return error
 */
func (m *MongoBase) Delete(query bson.M) error {
	session, db, err := getSession()

	if err != nil {
		return err
	}

	defer session.Close()

	c := session.DB(db).C(m.Collection)

	_, delErr := c.RemoveAll(query)

	if delErr != nil {
		LogErrorf("mongo collection %s delete error: %s", m.Collection, delErr.Error())
		return delErr
	}

	return nil
}

/**
 * Sum 字段求和
 * @param match bson.M (map[string]interface{}) 匹配条件
 * @param field string 聚合字段 (如："$count")
 * @return int, error
 */
func (m *MongoBase) Sum(match bson.M, field string) (int, error) {
	session, db, err := getSession()

	if err != nil {
		return 0, err
	}

	defer session.Close()

	c := session.DB(db).C(m.Collection)

	p := c.Pipe([]bson.M{
		{"$match": match},
		{"$group": bson.M{"_id": 1, "total": bson.M{"$sum": field}}},
	})

	result := bson.M{}

	pipeErr := p.One(&result)

	if pipeErr != nil {
		LogErrorf("mongo collection %s sum error: %s", m.Collection, pipeErr.Error())
		return 0, pipeErr
	}

	fmt.Println(result)
	total, ok := result["total"].(int)

	if !ok {
		LogErrorf("mongo collection %s sum error: type assertion error, result %v is %v", m.Collection, result["total"], reflect.TypeOf(result["total"]))
		return 0, fmt.Errorf("mongo collection %s sum error: type assertion error, result %v is %v", m.Collection, result["total"], reflect.TypeOf(result["total"]))
	}

	return total, nil
}
