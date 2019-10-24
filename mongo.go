package yiigo

import (
	"context"
	"sync"
	"time"

	"github.com/pelletier/go-toml"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"go.uber.org/zap"
)

const (
	Primary            = "primary"             // Default mode. All operations read from the current replica set primary.
	PrimaryPreferred   = "primary_preferred"   // Read from the primary if available. Read from the secondary otherwise.
	Secondary          = "secondary"           // Read from one of the nearest secondary members of the replica set.
	SecondaryPreferred = "secondary_preferred" // Read from one of the nearest secondaries if available. Read from primary otherwise.
	Nearest            = "nearest"             // Read from one of the nearest members, irrespective of it being primary or secondary.
)

type mongoConfig struct {
	Dsn             string `toml:"dsn"`
	ConnectTimeout  int    `toml:"connect_timeout"`
	PoolSize        int    `toml:"pool_size"`
	MaxConnIdleTime int    `toml:"max_conn_idle_time"`
	Mode            string `toml:"mode"`
}

var (
	defaultMongo *mongo.Client
	mgoMap       sync.Map
)

func mongoDial(cfg *mongoConfig) (*mongo.Client, error) {
	clientOptions := options.Client()

	clientOptions.ApplyURI(cfg.Dsn)
	clientOptions.SetConnectTimeout(time.Duration(cfg.ConnectTimeout) * time.Second)
	clientOptions.SetMaxPoolSize(uint64(cfg.PoolSize))
	clientOptions.SetMaxConnIdleTime(time.Duration(cfg.MaxConnIdleTime) * time.Second)

	if cfg.Mode != "" {
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
	tree, ok := env.Get("mongo").(*toml.Tree)

	if !ok {
		return
	}

	keys := tree.Keys()

	if len(keys) == 0 {
		return
	}

	for _, v := range keys {
		node, ok := tree.Get(v).(*toml.Tree)

		if !ok {
			continue
		}

		cfg := new(mongoConfig)

		if err := node.Unmarshal(cfg); err != nil {
			logger.Panic("yiigo: mongodb init error", zap.String("name", v), zap.Error(err))
		}

		client, err := mongoDial(cfg)

		if err != nil {
			logger.Panic("yiigo: mongodb init error", zap.String("name", v), zap.Error(err))
		}

		if v == AsDefault {
			defaultMongo = client
		}

		mgoMap.Store(v, client)
	}
}

// Mongo returns a mongo client.
func Mongo(name ...string) *mongo.Client {
	if len(name) == 0 {
		if defaultMongo == nil {
			logger.Panic("yiigo: invalid mongodb", zap.String("name", AsDefault))
		}

		return defaultMongo
	}

	v, ok := mgoMap.Load(name[0])

	if !ok {
		logger.Panic("yiigo: invalid mongodb", zap.String("name", name[0]))
	}

	return v.(*mongo.Client)
}
