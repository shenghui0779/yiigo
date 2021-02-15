package yiigo

import (
	"context"
	"fmt"
	"sync"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
)

type mongoConfig struct {
	DSN string `toml:"dsn"`
}

var (
	defaultMongo *mongo.Client
	mgoMap       sync.Map
)

func mongoDial(cfg *mongoConfig) (*mongo.Client, error) {
	opts := options.Client().ApplyURI(cfg.DSN)

	ctx := context.TODO()

	if opts.ConnectTimeout != nil {
		var cancel context.CancelFunc

		ctx, cancel = context.WithTimeout(ctx, *opts.ConnectTimeout)

		defer cancel()
	}

	c, err := mongo.Connect(ctx, opts)

	if err != nil {
		return nil, err
	}

	// verify connection
	if err = c.Ping(ctx, opts.ReadPreference); err != nil {
		return nil, err
	}

	return c, nil
}

func initMongoDB() {
	configs := make(map[string]*mongoConfig, 0)

	if err := env.Get("mongo").Unmarshal(&configs); err != nil {
		logger.Panic("yiigo: mongodb init error", zap.Error(err))
	}

	if len(configs) == 0 {
		return
	}

	for name, cfg := range configs {
		client, err := mongoDial(cfg)

		if err != nil {
			logger.Panic("yiigo: mongodb init error", zap.String("name", name), zap.Error(err))
		}

		if name == defalutConn {
			defaultMongo = client
		}

		mgoMap.Store(name, client)

		logger.Info(fmt.Sprintf("yiigo: mongodb.%s is OK.", name))
	}
}

// Mongo returns a mongo client.
func Mongo(name ...string) *mongo.Client {
	if len(name) == 0 {
		if defaultMongo == nil {
			logger.Panic(fmt.Sprintf("yiigo: unknown mongodb.%s (forgotten configure?)", defalutConn))
		}

		return defaultMongo
	}

	v, ok := mgoMap.Load(name[0])

	if !ok {
		logger.Panic(fmt.Sprintf("yiigo: unknown mongodb.%s (forgotten configure?)", name[0]))
	}

	return v.(*mongo.Client)
}
