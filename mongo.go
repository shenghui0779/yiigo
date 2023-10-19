package yiigo

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var mgoMap = make(map[string]*mongo.Client)

func initMongoDB(name, dsn string) error {
	opts := options.Client().ApplyURI(dsn)

	client, err := mongo.Connect(context.Background(), opts)
	if err != nil {
		return err
	}

	timeout := 10 * time.Second
	if opts.ConnectTimeout != nil {
		timeout = *opts.ConnectTimeout
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// verify connection
	if err = client.Ping(ctx, opts.ReadPreference); err != nil {
		return err
	}

	mgoMap[name] = client

	return nil
}

// Mongo 返回一个MongoDB客户端
func Mongo(name ...string) (*mongo.Client, error) {
	key := Default
	if len(name) != 0 {
		key = name[0]
	}

	cli, ok := mgoMap[key]
	if !ok {
		return nil, fmt.Errorf("unknown mongodb.%s (forgotten configure?)", key)
	}

	return cli, nil
}

// MustMongo 返回一个MongoDB客户端，如果不存在，则Panic
func MustMongo(name ...string) *mongo.Client {
	cli, err := Mongo(name...)
	if err != nil {
		logger.Panic(err.Error())
	}

	return cli
}
