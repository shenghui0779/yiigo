package yiigo

import (
	"errors"
	"fmt"
	"reflect"
	"sync"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type MongoBase struct {
	CollectionName string
}

type Sequence struct {
	Id  string `bson:"_id"`
	Seq int64  `bson:"seq"`
}

var (
	mongoSession *mgo.Session
	mongoMux     sync.Mutex
)

func initMongo() error {
	mongoMux.Lock()
	defer mongoMux.Unlock()

	if mongoSession == nil {
		var err error
		host := GetConfigString("mongo", "host", "localhost")
		port := GetConfigInt("mongo", "port", 27017)
		poolLimit := GetConfigInt("mongo", "poolLimit", 10)

		connStr := fmt.Sprintf("mongodb://%s:%d", host, port)

		mongoSession, err = mgo.Dial(connStr)

		if err != nil {
			LogError("connect to mongo server error: ", err.Error())
			return err
		}

		mongoSession.SetPoolLimit(poolLimit)
	}

	return nil
}

func getSession() (*mgo.Session, string, error) {
	if mongoSession == nil {
		err := initMongo()

		if err != nil {
			return nil, "", err
		}
	}

	session := mongoSession.Clone()
	db := GetConfigString("mongo", "database", "test")

	return session, db, nil
}

func (m *MongoBase) refreshSequence() (int64, error) {
	session, db, err := getSession()
	defer session.Close()

	if err != nil {
		return 0, err
	}

	c := session.DB(db).C("sequence")

	condition := bson.M{"_id": m.CollectionName}

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

	fmt.Println(sequence)

	return sequence.Seq, nil
}

func (m *MongoBase) Insert(data interface{}) error {
	session, db, err := getSession()
	defer session.Close()

	if err != nil {
		return err
	}

	refVal := reflect.ValueOf(data)

	if refVal.Kind() != reflect.Ptr {
		refVal = reflect.ValueOf(&data)
	}

	elem := refVal.Elem()
	fmt.Println(elem.Kind())
	return nil

	if elem.Kind() != reflect.Struct {
		LogErrorf("cannot use (type %v) as mongo Insert param", elem.Type())
		return errors.New(fmt.Sprintf("cannot use (type %v) as mongo Insert param", elem.Type()))
	}

	id, seqErr := m.refreshSequence()

	if seqErr != nil {
		return seqErr
	}

	elem.Field(0).SetInt(id)

	c := session.DB(db).C(m.CollectionName)

	insertErr := c.Insert(data)

	if insertErr != nil {
		LogErrorf("mongo collection %s insert error: %s", m.CollectionName, insertErr.Error())
		return insertErr
	}

	return nil
}
