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

var (
	defaultMongo *mongo.Client
	mgoMap       sync.Map
)

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

	if name == Default {
		defaultMongo = client
	}

	mgoMap.Store(name, client)

	logger.Info(fmt.Sprintf("mongodb.%s is OK", name))
}

// Mongo 返回一个MongoDB客户端
func Mongo(name ...string) *mongo.Client {
	if len(name) == 0 || name[0] == Default {
		if defaultMongo == nil {
			logger.Panic(fmt.Sprintf("unknown mongodb.%s (forgotten configure?)", Default))
		}

		return defaultMongo
	}

	v, ok := mgoMap.Load(name[0])
	if !ok {
		logger.Panic(fmt.Sprintf("unknown mongodb.%s (forgotten configure?)", name[0]))
	}

	return v.(*mongo.Client)
}
