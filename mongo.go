package yiigo

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

// Mode indicates the user's preference on reads.
type Mode int

const (
	Primary            Mode = 1 // Default mode. All operations read from the current replica set primary.
	PrimaryPreferred   Mode = 2 // Read from the primary if available. Read from the secondary otherwise.
	Secondary          Mode = 3 // Read from one of the nearest secondary members of the replica set.
	SecondaryPreferred Mode = 4 // Read from one of the nearest secondaries if available. Read from primary otherwise.
	Nearest            Mode = 5 // Read from one of the nearest members, irrespective of it being primary or secondary.
)

type mongoOptions struct {
	connTimeout            time.Duration
	poolSize               int
	maxConnIdleTime        time.Duration
	localThreshold         time.Duration
	serverSelectionTimeout time.Duration
	socketTimeout          time.Duration
	heartbeatInterval      time.Duration
	retryWrites            bool
	direct                 bool
	mode                   Mode
}

// MongoOption configures how we set up the mongo
type MongoOption interface {
	apply(options *mongoOptions)
}

// funcMongoOption implements mongo option
type funcMongoOption struct {
	f func(options *mongoOptions)
}

func (fo *funcMongoOption) apply(o *mongoOptions) {
	fo.f(o)
}

func newFuncMongoOption(f func(options *mongoOptions)) *funcMongoOption {
	return &funcMongoOption{f: f}
}

// WithMongoConnTimeout specifies the `ConnTimeout` to mongo.
func WithMongoConnTimeout(d time.Duration) MongoOption {
	return newFuncMongoOption(func(o *mongoOptions) {
		o.connTimeout = d
	})
}

// WithMongoPoolSize specifies the `PoolSize` to mongo.
// MaxPoolSize specifies the max size of a server's connection pool.
func WithMongoPoolSize(n int) MongoOption {
	return newFuncMongoOption(func(o *mongoOptions) {
		o.poolSize = n
	})
}

// WithMongoMaxConnIdleTime specifies the `MaxConnIdleTime` to mongo.
// MaxConnIdleTime specifies the maximum number of milliseconds that a connection can remain idle
// in a connection pool before being removed and closed.
func WithMongoMaxConnIdleTime(d time.Duration) MongoOption {
	return newFuncMongoOption(func(o *mongoOptions) {
		o.maxConnIdleTime = d
	})
}

// WithMongoLocalThreshold specifies the `LocalThreshold` to mongo.
// LocalThreshold specifies how far to distribute queries, beyond the server with the fastest
// round-trip time. If a server's roundtrip time is more than LocalThreshold slower than the
// the fastest, the driver will not send queries to that server.
func WithMongoLocalThreshold(d time.Duration) MongoOption {
	return newFuncMongoOption(func(o *mongoOptions) {
		o.localThreshold = d
	})
}

// WithMongoServerSelectionTimeout specifies the `ServerSelectionTimeout` to mongo.
// ServerSelectionTimeout specifies a timeout in milliseconds to block for server selection.
func WithMongoServerSelectionTimeout(d time.Duration) MongoOption {
	return newFuncMongoOption(func(o *mongoOptions) {
		o.serverSelectionTimeout = d
	})
}

// WithMongoSocketTimeout specifies the `SocketTimeout` to mongo.
// SocketTimeout specifies the time in milliseconds to attempt to send or receive on a socket
// before the attempt times out.
func WithMongoSocketTimeout(d time.Duration) MongoOption {
	return newFuncMongoOption(func(o *mongoOptions) {
		o.socketTimeout = d
	})
}

// WithMongoHeartbeatInterval specifies the `HeartbeatInterval` to mongo.
// HeartbeatInterval specifies the interval to wait between server monitoring checks.
func WithMongoHeartbeatInterval(d time.Duration) MongoOption {
	return newFuncMongoOption(func(o *mongoOptions) {
		o.heartbeatInterval = d
	})
}

// WithRetryWrites specifies the `RetryWrites` to mongo.
// RetryWrites specifies whether the client has retryable writes enabled.
func WithRetryWrites() MongoOption {
	return newFuncMongoOption(func(o *mongoOptions) {
		o.retryWrites = true
	})
}

// WithDirect specifies the `Direct` to mongo.
// Direct specifies whether the driver should connect directly to the server instead of
// auto-discovering other servers in the cluster.
func WithDirect() MongoOption {
	return newFuncMongoOption(func(o *mongoOptions) {
		o.direct = true
	})
}

// WithMongoMode specifies the `Mode` to mongo.
// Mode specifies the read preference.
func WithMongoMode(m Mode) MongoOption {
	return newFuncMongoOption(func(o *mongoOptions) {
		o.mode = m
	})
}

var (
	// Mongo default mongo client
	Mongo  *mongo.Client
	mgoMap sync.Map
)

func mongoDial(dsn string, mgoOptions ...MongoOption) (*mongo.Client, error) {
	o := &mongoOptions{
		connTimeout:     10 * time.Second,
		poolSize:        10,
		maxConnIdleTime: 60 * time.Second,
	}

	if len(mgoOptions) > 0 {
		for _, option := range mgoOptions {
			option.apply(o)
		}
	}

	clientOptions := options.Client()

	clientOptions.ApplyURI(dsn)
	clientOptions.SetConnectTimeout(o.connTimeout)
	clientOptions.SetMaxPoolSize(uint16(o.poolSize))
	clientOptions.SetMaxConnIdleTime(o.maxConnIdleTime)

	if o.localThreshold != 0 {
		clientOptions.SetLocalThreshold(o.localThreshold)
	}

	if o.serverSelectionTimeout != 0 {
		clientOptions.SetServerSelectionTimeout(o.serverSelectionTimeout)
	}

	if o.socketTimeout != 0 {
		clientOptions.SetSocketTimeout(o.socketTimeout)
	}

	if o.heartbeatInterval != 0 {
		clientOptions.SetHeartbeatInterval(o.heartbeatInterval)
	}

	if o.retryWrites {
		clientOptions.SetRetryWrites(true)
	}

	if o.direct {
		clientOptions.SetDirect(true)
	}

	if o.mode != 0 {
		switch o.mode {
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

	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	client, err := mongo.Connect(ctx, clientOptions)

	return client, err
}

// RegisterMongoDB register a mongodb, `dsn` eg: `mongodb://username:password@localhost:27017`.
// The default `ConnTimeout` is 10s.
// The default `PoolSize` is 10.
// The default `MaxConnIdleTime` is 60s.
func RegisterMongoDB(name, dsn string, options ...MongoOption) error {
	client, err := mongoDial(dsn, options...)

	if err != nil {
		return err
	}

	mgoMap.Store(name, client)

	if name == AsDefault {
		Mongo = client
	}

	return nil
}

// UseMongo returns a mongo client.
func UseMongo(name string) *mongo.Client {
	v, ok := mgoMap.Load(name)

	if !ok {
		panic(fmt.Errorf("yiigo: mongo.%s is not registered", name))
	}

	return v.(*mongo.Client)
}
