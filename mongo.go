package yiigo

import (
	"context"
	"crypto/tls"
	"fmt"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readconcern"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"go.mongodb.org/mongo-driver/mongo/writeconcern"
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

// Concern for replica sets and replica set shards determines which data to return from a query.
type Concern int

const (
	Local        Concern = 1 // the query should return the instance’s most recent data.
	Available    Concern = 2 // the query should return data from the instance with no guarantee that the data has been written to a majority of the replica set members (i.e. may be rolled back).
	Majority     Concern = 3 // the query should return the instance’s most recent data acknowledged as having been written to a majority of members in the replica set.
	Linearizable Concern = 4 // that the query should return data that reflects all successful writes issued with a write concern of "majority" and acknowledged prior to the start of the read operation.
	Snapshot     Concern = 5 // only available for operations within multi-document transactions.
)

type mongoOptions struct {
	appName                string
	connTimeout            time.Duration
	poolSize               int
	maxConnIdleTime        time.Duration
	localThreshold         time.Duration
	serverSelectionTimeout time.Duration
	socketTimeout          time.Duration
	heartbeatInterval      time.Duration
	compressors            []string
	hosts                  []string
	replicaSet             string
	retryWrites            bool
	direct                 bool
	mode                   Mode
	readConcern            Concern
	writeConcern           []writeconcern.Option
	tlsConfig              *tls.Config
	zlibLevel              int
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

// WithMongoAppName specifies the `AppName` to mongo.
// AppName sets the client application name.
// This value is used by MongoDB when it logs connection information and profile information, such as slow queries.
func WithMongoAppName(s string) MongoOption {
	return newFuncMongoOption(func(o *mongoOptions) {
		o.appName = s
	})
}

// WithMongoConnTimeout specifies the `ConnTimeout` to mongo.
// ConnectTimeout sets the timeout for an initial connection to a server.
func WithMongoConnTimeout(d time.Duration) MongoOption {
	return newFuncMongoOption(func(o *mongoOptions) {
		o.connTimeout = d
	})
}

// WithMongoPoolSize specifies the `PoolSize` to mongo.
// MaxPoolSize sets the max size of a server's connection pool.
func WithMongoPoolSize(n int) MongoOption {
	return newFuncMongoOption(func(o *mongoOptions) {
		o.poolSize = n
	})
}

// WithMongoMaxConnIdleTime specifies the `MaxConnIdleTime` to mongo.
// MaxConnIdleTime sets the maximum number of milliseconds that a connection can remain idle
// in a connection pool before being removed and closed.
func WithMongoMaxConnIdleTime(d time.Duration) MongoOption {
	return newFuncMongoOption(func(o *mongoOptions) {
		o.maxConnIdleTime = d
	})
}

// WithMongoLocalThreshold specifies the `LocalThreshold` to mongo.
// LocalThreshold sets how far to distribute queries, beyond the server with the fastest
// round-trip time. If a server's roundtrip time is more than LocalThreshold slower than the
// the fastest, the driver will not send queries to that server.
func WithMongoLocalThreshold(d time.Duration) MongoOption {
	return newFuncMongoOption(func(o *mongoOptions) {
		o.localThreshold = d
	})
}

// WithMongoServerSelectionTimeout specifies the `ServerSelectionTimeout` to mongo.
// ServerSelectionTimeout sets a timeout in milliseconds to block for server selection.
func WithMongoServerSelectionTimeout(d time.Duration) MongoOption {
	return newFuncMongoOption(func(o *mongoOptions) {
		o.serverSelectionTimeout = d
	})
}

// WithMongoSocketTimeout specifies the `SocketTimeout` to mongo.
// SocketTimeout sets the time in milliseconds to attempt to send or receive on a socket
// before the attempt times out.
func WithMongoSocketTimeout(d time.Duration) MongoOption {
	return newFuncMongoOption(func(o *mongoOptions) {
		o.socketTimeout = d
	})
}

// WithMongoHeartbeatInterval specifies the `HeartbeatInterval` to mongo.
// HeartbeatInterval sets the interval to wait between server monitoring checks.
func WithMongoHeartbeatInterval(d time.Duration) MongoOption {
	return newFuncMongoOption(func(o *mongoOptions) {
		o.heartbeatInterval = d
	})
}

// WithMongoCompressors specifies the `Compressors` to mongo.
// Compressors sets the compressors that can be used when communicating with a server.
func WithMongoCompressors(s ...string) MongoOption {
	return newFuncMongoOption(func(o *mongoOptions) {
		o.compressors = s
	})
}

// WithMongoHosts specifies the `Hosts` to mongo.
// Hosts sets the initial list of addresses from which to discover the rest of the cluster.
func WithMongoHosts(s ...string) MongoOption {
	return newFuncMongoOption(func(o *mongoOptions) {
		o.hosts = s
	})
}

// WithMongoReplicaSet specifies the `ReplicaSet` to mongo.
// ReplicaSet sets the name of the replica set of the cluster.
func WithMongoReplicaSet(s string) MongoOption {
	return newFuncMongoOption(func(o *mongoOptions) {
		o.replicaSet = s
	})
}

// WithMongoRetryWrites specifies the `RetryWrites` to mongo.
// RetryWrites sets whether the client has retryable writes enabled.
func WithMongoRetryWrites(b bool) MongoOption {
	return newFuncMongoOption(func(o *mongoOptions) {
		o.retryWrites = b
	})
}

// WithMongoDirect specifies the `Direct` to mongo.
// Direct sets whether the driver should connect directly to the server instead of
// auto-discovering other servers in the cluster.
func WithMongoDirect(b bool) MongoOption {
	return newFuncMongoOption(func(o *mongoOptions) {
		o.direct = b
	})
}

// WithMongoMode specifies the `Mode` to mongo.
// Mode sets the read preference.
func WithMongoMode(m Mode) MongoOption {
	return newFuncMongoOption(func(o *mongoOptions) {
		o.mode = m
	})
}

// WithMongoReadConcern specifies the `ReadConcern` to mongo.
// ReadConcern sets the read concern.
func WithMongoReadConcern(c Concern) MongoOption {
	return newFuncMongoOption(func(o *mongoOptions) {
		o.readConcern = c
	})
}

// WithMongoWriteConcern specifies the `WriteConcern` to mongo.
// WriteConcern sets the write concern.
func WithMongoWriteConcern(c ...writeconcern.Option) MongoOption {
	return newFuncMongoOption(func(o *mongoOptions) {
		o.writeConcern = c
	})
}

// WithMongoTLSConfig specifies the `TLSConfig` to mongo.
// SetTLSConfig sets the tls.Config.
func WithMongoTLSConfig(c *tls.Config) MongoOption {
	return newFuncMongoOption(func(o *mongoOptions) {
		o.tlsConfig = c
	})
}

// WithMongoZlibLevel specifies the `ZlibLevel` to mongo.
// ZlibLevel sets the level for the zlib compressor.
func WithMongoZlibLevel(l int) MongoOption {
	return newFuncMongoOption(func(o *mongoOptions) {
		o.zlibLevel = l
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

	if o.appName != "" {
		clientOptions.SetAppName(o.appName)
	}

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

	if len(o.compressors) > 0 {
		clientOptions.SetCompressors(o.compressors)
	}

	if len(o.hosts) > 0 {
		clientOptions.SetHosts(o.hosts)
	}

	if o.replicaSet != "" {
		clientOptions.SetReplicaSet(o.replicaSet)
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

	if o.readConcern != 0 {
		switch o.readConcern {
		case Local:
			clientOptions.SetReadConcern(readconcern.Local())
		case Available:
			clientOptions.SetReadConcern(readconcern.Available())
		case Majority:
			clientOptions.SetReadConcern(readconcern.Majority())
		case Linearizable:
			clientOptions.SetReadConcern(readconcern.Linearizable())
		case Snapshot:
			clientOptions.SetReadConcern(readconcern.Snapshot())
		}
	}

	if len(o.writeConcern) > 0 {
		clientOptions.SetWriteConcern(writeconcern.New(o.writeConcern...))
	}

	if o.tlsConfig != nil {
		clientOptions.SetTLSConfig(o.tlsConfig)
	}

	if o.zlibLevel != 0 {
		clientOptions.SetZlibLevel(o.zlibLevel)
	}

	// validates the client options
	if err := clientOptions.Validate(); err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.TODO(), 10*time.Second)

	defer cancel()

	client, err := mongo.Connect(ctx, clientOptions)

	return client, err
}

// RegisterMongoDB register a mongodb, the param `dsn` eg: `mongodb://username:password@localhost:27017`.
//
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
