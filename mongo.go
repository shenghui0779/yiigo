package yiigo

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
)

var mgoMap sync.Map

func initMongoDB(name, dsn string) {
	opts := options.Client().ApplyURI(dsn)

	client, err := mongo.Connect(context.Background(), opts)
	if err != nil {
		logger.Panic(fmt.Sprintf("err mongodb.%s connect", name), zap.String("dsn", dsn), zap.Error(err))
	}

	timeout := 10 * time.Second
	if opts.ConnectTimeout != nil {
		timeout = *opts.ConnectTimeout
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// verify connection
	if err = client.Ping(ctx, opts.ReadPreference); err != nil {
		logger.Panic(fmt.Sprintf("err mongodb.%s ping", name), zap.String("dsn", dsn), zap.Error(err))
	}

	mgoMap.Store(name, client)

	logger.Info(fmt.Sprintf("mongodb.%s is OK", name))
}

// Mongo 返回一个MongoDB客户端
func Mongo(name ...string) (*mongo.Client, error) {
	key := Default
	if len(name) != 0 {
		key = name[0]
	}

	v, ok := mgoMap.Load(name[0])
	if !ok {
		return nil, fmt.Errorf("unknown mongodb.%s (forgotten configure?)", key)
	}

	return v.(*mongo.Client), nil
}

// MustMongo 返回一个MongoDB客户端，如果不存在，则Panic
func MustMongo(name ...string) *mongo.Client {
	cli, err := Mongo(name...)
	if err != nil {
		logger.Panic(err.Error())
	}

	return cli
}
