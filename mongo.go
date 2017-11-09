package yiigo

import (
	"fmt"
	"sync"

	ini "gopkg.in/ini.v1"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type Sequence struct {
	ID  string `bson:"_id"`
	Seq int    `bson:"seq"`
}

var (
	Mongo    *mgo.Session
	mongoMap map[string]*mgo.Session
	mongoMux sync.RWMutex
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

	host := EnvString("mongo", "host", "localhost")
	port := EnvInt("mongo", "port", 27017)
	username := EnvString("mongo", "username", "")
	password := EnvString("mongo", "password", "")

	dsn := fmt.Sprintf("mongodb://%s:%d", host, port)

	if username != "" {
		dsn = fmt.Sprintf("mongodb://%s:%s@%s:%d", username, password, host, port)
	}

	Mongo, err = mgo.Dial(dsn)

	if err != nil {
		return err
	}

	Mongo.SetPoolLimit(EnvInt("mongo", "poolLimit", 10))

	return nil
}

func initMultiMongo(sections []*ini.Section) error {
	mongoMap = make(map[string]*mgo.Session, len(sections))

	for _, v := range sections {
		host := v.Key("host").MustString("localhost")
		port := v.Key("port").MustInt(27017)
		username := v.Key("username").MustString("")
		password := v.Key("password").MustString("")

		dsn := fmt.Sprintf("mongodb://%s:%d", host, port)

		if username != "" {
			dsn = fmt.Sprintf("mongodb://%s:%s@%s:%d", username, password, host, port)
		}

		mongo, err := mgo.Dial(dsn)

		if err != nil {
			return err
		}

		mongo.SetPoolLimit(EnvInt("mongo", "poolLimit", 10))

		mongoMap[v.Name()] = mongo
	}

	if mongo, ok := mongoMap["mongo.default"]; ok {
		Mongo = mongo
	}

	return nil
}

// MongoSession get mongo session
func MongoSession(conn ...string) (*mgo.Session, error) {
	mongoMux.RLock()
	defer mongoMux.RUnlock()

	c := "default"

	if len(conn) > 0 {
		c = conn[0]
	}

	schema := fmt.Sprintf("mongo.%s", c)

	mongo, ok := mongoMap[schema]

	if !ok {
		return nil, fmt.Errorf("mongodb %s is not connected", schema)
	}

	return mongo.Clone(), nil
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
