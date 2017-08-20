package yiigo

import (
	"fmt"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type Sequence struct {
	ID  string `bson:"_id"`
	Seq int    `bson:"seq"`
}

var mongoSession *mgo.Session

// initMongo init mongo connections
func initMongo() error {
	var err error

	host := EnvString("mongo", "host", "localhost")
	port := EnvInt("mongo", "port", 27017)
	username := EnvString("mongo", "username", "")
	password := EnvString("mongo", "password", "")
	poolLimit := EnvInt("mongo", "poolLimit", 10)

	dsn := fmt.Sprintf("mongodb://%s:%d", host, port)

	if username != "" {
		dsn = fmt.Sprintf("mongodb://%s:%s@%s:%d", username, password, host, port)
	}

	mongoSession, err = mgo.Dial(dsn)

	if err != nil {
		return err
	}

	mongoSession.SetPoolLimit(poolLimit)

	return nil
}

// Mongo get a session
func Mongo() *mgo.Session {
	session := mongoSession.Clone()
	return session
}

// Sequence get an auto increment _id
func Sequence(db string, collection string, seqs ...int) (int, error) {
	session := Mongo()
	defer session.Close()

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
