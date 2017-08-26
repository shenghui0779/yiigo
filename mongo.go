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

// initMongo init mongo connection
func initMongo() error {
	var err error

	host := EnvString("mongo", "host", "localhost")
	port := EnvInt("mongo", "port", 27017)
	username := EnvString("mongo", "username", "")
	password := EnvString("mongo", "password", "")

	dsn := fmt.Sprintf("mongodb://%s:%d", host, port)

	if username != "" {
		dsn = fmt.Sprintf("mongodb://%s:%s@%s:%d", username, password, host, port)
	}

	mongoSession, err = mgo.Dial(dsn)

	if err != nil {
		return err
	}

	mongoSession.SetPoolLimit(EnvInt("mongo", "poolLimit", 10))

	return nil
}

// Mongo get mongo session
func Mongo() *mgo.Session {
	session := mongoSession.Clone()

	return session
}

// Seq get auto increment id
func Seq(db string, collection string, seqs ...int) (int, error) {
	session := mongoSession.Clone()
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
