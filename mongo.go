package yiigo

import (
	"errors"
	"fmt"
	"sync"
	"time"

	toml "github.com/pelletier/go-toml"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type mongoConf struct {
	Name      string `toml:"name"`
	Host      string `toml:"host"`
	Port      int    `toml:"port"`
	Username  string `toml:"username"`
	Password  string `toml:"password"`
	Timeout   int    `toml:"timeout"`
	PoolLimit int    `toml:"poolLimit"`
	Mode      int    `toml:"mode"`
}

// Sequence model for _id auto_increment of mongo
type Sequence struct {
	ID  string `bson:"_id"`
	Seq int64  `bson:"seq"`
}

var (
	// Mongo default mongo session
	Mongo  *mgo.Session
	mgoMap sync.Map
)

func initMongo() error {
	result := Env.Get("mongo")

	if result == nil {
		return nil
	}

	switch node := result.(type) {
	case *toml.Tree:
		conf := &mongoConf{}
		err := node.Unmarshal(conf)

		if err != nil {
			return err
		}

		err = initSingleMongo(conf)

		if err != nil {
			return err
		}
	case []*toml.Tree:
		conf := make([]*mongoConf, 0, len(node))

		for _, v := range node {
			c := &mongoConf{}
			err := v.Unmarshal(c)

			if err != nil {
				return err
			}

			conf = append(conf, c)
		}

		err := initMultiMongo(conf)

		if err != nil {
			return err
		}
	default:
		return errors.New("yiigo: invalid mongo config")
	}

	return nil
}

func initSingleMongo(conf *mongoConf) error {
	var err error

	Mongo, err = mongoDial(conf)

	if err != nil {
		return fmt.Errorf("yiigo: mongo.default connect error: %s", err.Error())
	}

	mgoMap.Store("default", Mongo)

	return nil
}

func initMultiMongo(conf []*mongoConf) error {
	for _, v := range conf {
		m, err := mongoDial(v)

		if err != nil {
			return fmt.Errorf("yiigo: mongo.%s connect error: %s", v.Name, err.Error())
		}

		mgoMap.Store(v.Name, m)
	}

	if v, ok := mgoMap.Load("default"); ok {
		Mongo = v.(*mgo.Session)
	}

	return nil
}

func mongoDial(conf *mongoConf) (*mgo.Session, error) {
	dsn := fmt.Sprintf("mongodb://%s:%d", conf.Host, conf.Port)

	if conf.Username != "" {
		dsn = fmt.Sprintf("mongodb://%s:%s@%s:%d", conf.Username, conf.Password, conf.Host, conf.Port)
	}

	m, err := mgo.DialWithTimeout(dsn, time.Duration(conf.Timeout)*time.Second)

	if err != nil {
		return nil, err
	}

	if err := m.Ping(); err != nil {
		return nil, err
	}

	if conf.PoolLimit != 0 {
		m.SetPoolLimit(conf.PoolLimit)
	}

	if conf.Mode != 0 {
		m.SetMode(mgo.Mode(conf.Mode), true)
	}

	return m, nil
}

// MongoSession returns a mongo session.
func MongoSession(conn ...string) (*mgo.Session, error) {
	schema := "default"

	if len(conn) > 0 {
		schema = conn[0]
	}

	v, ok := mgoMap.Load(schema)

	if !ok {
		return nil, fmt.Errorf("yiigo: mongo.%s is not connected", schema)
	}

	session := v.(*mgo.Session)

	return session.Clone(), nil
}

// SeqID returns _id auto_increment to mongo.
func SeqID(session *mgo.Session, db string, collection string, seqs ...int64) (int64, error) {
	var seq int64 = 1

	if len(seqs) > 0 {
		seq = seqs[0]
	}

	condition := bson.M{"_id": collection}

	change := mgo.Change{
		Update:    bson.M{"$inc": bson.M{"seq": seq}},
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
