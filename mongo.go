package yiigo

import (
	"errors"
	"fmt"
	"sync"
	"time"

	toml "github.com/pelletier/go-toml"
	"gopkg.in/mgo.v2"
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
	var err error

	result := Env.Get("mongo")

	switch node := result.(type) {
	case *toml.Tree:
		conf := &mongoConf{}
		err = node.Unmarshal(conf)

		if err != nil {
			break
		}

		err = initSingleMongo(conf)
	case []*toml.Tree:
		conf := make([]*mongoConf, 0, len(node))

		for _, v := range node {
			c := &mongoConf{}
			err = v.Unmarshal(c)

			if err != nil {
				break
			}

			conf = append(conf, c)
		}

		err = initMultiMongo(conf)
	default:
		return errors.New("mongo error config")
	}

	if err != nil {
		return fmt.Errorf("mongo error: %s", err.Error())
	}

	return nil
}

func initSingleMongo(conf *mongoConf) error {
	var err error

	Mongo, err = mongoDial(conf)

	return err
}

func initMultiMongo(conf []*mongoConf) error {
	for _, v := range conf {
		m, err := mongoDial(v)

		if err != nil {
			return err
		}

		mongoMap.Store(v.Name, m)
	}

	if v, ok := mongoMap.Load("default"); ok {
		Mongo = v.(*mgo.Session)
	}

	return nil
}

func mongoDial(conf *mongoConf) (*mgo.Session, error) {
	dsn := fmt.Sprintf("mongodb://%s:%d", conf.Host, conf.Port)

	if conf.Username != "" {
		dsn = fmt.Sprintf("mongodb://%s:%s@%s:%d", conf.Username, conf.Password, conf.Host, conf.Port)
	}

	m, err := mgo.DialWithTimeout(dsn, time.Duration(conf.Timeout)*time.Millisecond)

	if err != nil {
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

// MongoSession get mongo session
func MongoSession(conn ...string) (*mgo.Session, error) {
	schema := "default"

	if len(conn) > 0 {
		schema = conn[0]
	}

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
