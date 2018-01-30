package yiigo

import (
	"fmt"
	"sync"

	ini "gopkg.in/ini.v1"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

// Sequence model for _id auto_increment
type Sequence struct {
	ID  string `bson:"_id"`
	Seq int    `bson:"seq"`
}

var (
	// Mongo default session
	Mongo    *mgo.Session
	mongoMap sync.Map
)

func initMongo() error {
	sections := childSections("mongo")

	if len(sections) > 0 {
		return initMultiMongo(sections)
	}

	return initSingleMongo()
}

func initSingleMongo() error {
	var err error

	section := env.Section("mongo")

	Mongo, err = mongoDial(section)

	if err != nil {
		return fmt.Errorf("mongo error: %s", err.Error())
	}

	return nil
}

func initMultiMongo(sections []*ini.Section) error {
	for _, v := range sections {
		m, err := mongoDial(v)

		if err != nil {
			return fmt.Errorf("mongo error: %s", err.Error())
		}

		mongoMap.Store(v.Name(), m)
	}

	if v, ok := mongoMap.Load("mongo.default"); ok {
		Mongo = v.(*mgo.Session)
	}

	return nil
}

func mongoDial(section *ini.Section) (*mgo.Session, error) {
	host := section.Key("host").MustString("127.0.0.1")
	port := section.Key("port").MustInt(27017)
	username := section.Key("username").MustString("")
	password := section.Key("password").MustString("")
	poolLimit := section.Key("poolLimit").MustInt(10)

	dsn := fmt.Sprintf("mongodb://%s:%d", host, port)

	if username != "" {
		dsn = fmt.Sprintf("mongodb://%s:%s@%s:%d", username, password, host, port)
	}

	m, err := mgo.Dial(dsn)

	if err != nil {
		return nil, err
	}

	m.SetPoolLimit(poolLimit)

	return m, nil
}

// MongoSession get mongo session
func MongoSession(conn ...string) (*mgo.Session, error) {
	c := "default"

	if len(conn) > 0 {
		c = conn[0]
	}

	schema := fmt.Sprintf("mongo.%s", c)

	v, ok := mongoMap.Load(schema)

	if !ok {
		return nil, fmt.Errorf("mongodb %s is not connected", schema)
	}

	session := v.(*mgo.Session)

	return session.Clone(), nil
}

// Seq get auto increment id
func Seq(session *mgo.Session, db string, collection string, seqs ...int) (int, error) {
	if len(seqs) == 0 {
		seqs = append(seqs, 1)
	}

	condition := bson.M{"_id": collection}

	change := mgo.Change{
		Update:    bson.M{"$inc": bson.M{"seq": seqs[0]}},
		Upsert:    true,
		ReturnNew: true,
	}

	sequence := Sequence{}

	_, err := session.DB(db).C("sequence").Find(condition).Apply(change, &sequence)

	if err != nil {
		return 0, err
	}

	return sequence.Seq, nil
}
