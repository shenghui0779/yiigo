package yiigo

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"go.uber.org/zap"
)

// MongoMode indicates the user's preference on reads.
type MongoMode string

const (
	Primary            MongoMode = "primary"             // Default mode. All operations read from the current replica set primary.
	PrimaryPreferred   MongoMode = "primary_preferred"   // Read from the primary if available. Read from the secondary otherwise.
	Secondary          MongoMode = "secondary"           // Read from one of the nearest secondary members of the replica set.
	SecondaryPreferred MongoMode = "secondary_preferred" // Read from one of the nearest secondaries if available. Read from primary otherwise.
	Nearest            MongoMode = "nearest"             // Read from one of the nearest members, irrespective of it being primary or secondary.
)

type mongoConfig struct {
	DSN             string    `toml:"dsn"`
	ConnectTimeout  int       `toml:"connect_timeout"`
	MinPoolSize     int       `toml:"min_pool_size"`
	MaxPoolSize     int       `toml:"max_pool_size"`
	MaxConnIdleTime int       `toml:"max_conn_idle_time"`
	Mode            MongoMode `toml:"mode"`
}

var (
	defaultMongo *mongo.Client
	mgoMap       sync.Map
)

func mongoDial(cfg *mongoConfig) (*mongo.Client, error) {
	clientOptions := options.Client()

	clientOptions.ApplyURI(cfg.DSN)
	clientOptions.SetConnectTimeout(time.Duration(cfg.ConnectTimeout) * time.Second)
	clientOptions.SetMinPoolSize(uint64(cfg.MinPoolSize))
	clientOptions.SetMaxPoolSize(uint64(cfg.MaxPoolSize))
	clientOptions.SetMaxConnIdleTime(time.Duration(cfg.MaxConnIdleTime) * time.Second)

	switch cfg.Mode {
	case Primary:
		clientOptions.SetReadPreference(readpref.Primary())
	case PrimaryPreferred:
		clientOptions.SetReadPreference(readpref.PrimaryPreferred())
	case Secondary:
		clientOptions.SetReadPreference(readpref.Secondary())
	case SecondaryPreferred:
		clientOptions.SetReadPreference(readpref.SecondaryPreferred())
	case Nearest:
		clientOptions.SetReadPreference(readpref.Nearest())
	}

	// validates the client options
	if err := clientOptions.Validate(); err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.TODO(), time.Duration(cfg.ConnectTimeout)*time.Second)

	defer cancel()

	return mongo.Connect(ctx, clientOptions)
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

		if name == AsDefault {
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
			logger.Panic(fmt.Sprintf("yiigo: unknown mongodb.%s (forgotten configure?)", AsDefault))
		}

		return defaultMongo
	}

	v, ok := mgoMap.Load(name[0])

	if !ok {
		logger.Panic(fmt.Sprintf("yiigo: unknown mongodb.%s (forgotten configure?)", name[0]))
	}

	return v.(*mongo.Client)
}
