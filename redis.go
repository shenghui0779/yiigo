package yiigo

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/shenghui0779/vitess_pool"
	"go.uber.org/zap"
)

// RedisConn redis connection resource
type RedisConn struct {
	redis.Conn
}

// Close closes the connection resource
func (rc *RedisConn) Close() {
	if err := rc.Conn.Close(); err != nil {
		logger.Error("[yiigo] redis conn closed error", zap.Error(err))
	}
}

// RedisPool redis pool resource
type RedisPool interface {
	// Get returns a connection resource from the pool.
	// Context with timeout can specify the wait timeout for pool.
	Get(ctx context.Context) (*RedisConn, error)

	// Put returns a connection resource to the pool.
	Put(rc *RedisConn)
}

type RedisOptions struct {
	// Dialer is a custom dial function for creating TCP connections,
	// otherwise a net.Dialer customized via the other options is used.
	Dialer func(ctx context.Context, network, addr string) (net.Conn, error)

	// Username to be used when connecting to the Redis server when Redis ACLs are used.
	Username string

	// Password to be used when connecting to the Redis server.
	Password string

	// Database to be selected when dialing a connection.
	Database int

	// ConnTimeout is the timeout for connecting to the Redis server.
	// Use value -1 for no timeout and 0 for default.
	// Default is 10 seconds.
	ConnTimeout time.Duration

	// ReadTimeout is the timeout for reading a single command reply.
	// Use value -1 for no timeout and 0 for default.
	// Default is 10 seconds.
	ReadTimeout time.Duration

	// WriteTimeout is the timeout for writing a single command.
	// Use value -1 for no timeout and 0 for default.
	// Default is 10 seconds.
	WriteTimeout time.Duration

	// PoolSize is the maximum number of possible resources in the pool.
	// Use value -1 for no timeout and 0 for default.
	// Default is 10.
	PoolSize int

	// PoolPrefill is the number of resources to be pre-filled in the pool.
	PoolPrefill int

	// IdleTimeout is the amount of time after which client closes idle connections.
	// Use value -1 for no timeout and 0 for default.
	// Default is 5 minutes.
	IdleTimeout time.Duration

	// TLSConfig to be used when a TLS connection is dialed.
	TLSConfig *tls.Config
}

func (o *RedisOptions) rebuild(opt *RedisOptions) {
	o.Dialer = opt.Dialer
	o.TLSConfig = opt.TLSConfig

	if len(opt.Username) != 0 {
		o.Username = opt.Username
	}

	if len(opt.Password) != 0 {
		o.Password = opt.Password
	}

	if opt.Database > 0 {
		o.Database = opt.Database
	}

	if opt.ConnTimeout > 0 {
		o.ConnTimeout = opt.ConnTimeout
	} else {
		if opt.ConnTimeout == -1 {
			o.ConnTimeout = 0
		}
	}

	if opt.ReadTimeout > 0 {
		o.ReadTimeout = opt.ReadTimeout
	} else {
		if opt.ReadTimeout == -1 {
			o.ReadTimeout = 0
		}
	}

	if opt.WriteTimeout > 0 {
		o.WriteTimeout = opt.WriteTimeout
	} else {
		if opt.WriteTimeout == -1 {
			o.WriteTimeout = 0
		}
	}

	if opt.PoolSize > 0 {
		o.PoolSize = opt.PoolSize
	}

	if opt.PoolPrefill > 0 {
		o.PoolPrefill = opt.PoolPrefill
	}

	if opt.IdleTimeout > 0 {
		o.IdleTimeout = opt.IdleTimeout
	} else {
		if opt.IdleTimeout == -1 {
			o.IdleTimeout = 0
		}
	}
}

type redisResourcePool struct {
	addr    string
	options *RedisOptions
	pool    *vitess_pool.ResourcePool
	mutex   sync.Mutex
}

func (rp *redisResourcePool) dial() (redis.Conn, error) {
	dialOptions := []redis.DialOption{
		redis.DialDatabase(rp.options.Database),
		redis.DialConnectTimeout(rp.options.ConnTimeout),
		redis.DialReadTimeout(rp.options.ReadTimeout),
		redis.DialWriteTimeout(rp.options.WriteTimeout),
	}

	if len(rp.options.Username) != 0 {
		dialOptions = append(dialOptions, redis.DialUsername(rp.options.Username))
	}

	if len(rp.options.Password) != 0 {
		dialOptions = append(dialOptions, redis.DialUsername(rp.options.Password))
	}

	if rp.options.Dialer != nil {
		dialOptions = append(dialOptions, redis.DialContextFunc(rp.options.Dialer))
	}

	if rp.options.TLSConfig != nil {
		dialOptions = append(dialOptions, redis.DialTLSConfig(rp.options.TLSConfig))
	}

	conn, err := redis.Dial("tcp", rp.addr, dialOptions...)

	return conn, err
}

func (rp *redisResourcePool) init() {
	rp.mutex.Lock()
	defer rp.mutex.Unlock()

	if rp.pool != nil && !rp.pool.IsClosed() {
		return
	}

	df := func() (vitess_pool.Resource, error) {
		conn, err := rp.dial()

		if err != nil {
			return nil, err
		}

		return &RedisConn{conn}, nil
	}

	rp.pool = vitess_pool.NewResourcePool(df, rp.options.PoolSize, rp.options.PoolSize, rp.options.IdleTimeout, rp.options.PoolPrefill)
}

func (rp *redisResourcePool) Get(ctx context.Context) (*RedisConn, error) {
	if rp.pool.IsClosed() {
		rp.init()
	}

	resource, err := rp.pool.Get(ctx)

	if err != nil {
		return nil, err
	}

	rc := resource.(*RedisConn)

	// If rc is error, close and reconnect
	if err = rc.Err(); err != nil {
		logger.Warn("[yiigo] redis pool conn is error, reconnect", zap.Error(err))

		conn, dialErr := rp.dial()

		if dialErr != nil {
			rp.pool.Put(rc)

			return nil, dialErr
		}

		rc.Close()

		return &RedisConn{conn}, nil
	}

	return rc, nil
}

func (rp *redisResourcePool) Put(conn *RedisConn) {
	rp.pool.Put(conn)
}

var (
	defaultRedis RedisPool
	redisMap     sync.Map
)

func newRedisPool(addr string, opt *RedisOptions) RedisPool {
	pool := &redisResourcePool{
		addr: addr,
		options: &RedisOptions{
			ConnTimeout:  10 * time.Second,
			ReadTimeout:  10 * time.Second,
			WriteTimeout: 10 * time.Second,
			PoolSize:     10,
			IdleTimeout:  5 * time.Minute,
		},
	}

	if opt != nil {
		pool.options.rebuild(opt)
	}

	pool.init()

	return pool
}

func initRedis(name, address string, opt *RedisOptions) {
	pool := newRedisPool(address, opt)

	// verify connection
	conn, err := pool.Get(context.TODO())

	if err != nil {
		logger.Panic("[yiigo] redis init error", zap.String("name", name), zap.Error(err))
	}

	if _, err = conn.Do("PING"); err != nil {
		conn.Close()

		logger.Panic("[yiigo] redis init error", zap.String("name", name), zap.Error(err))
	}

	pool.Put(conn)

	if name == Default {
		defaultRedis = pool
	}

	redisMap.Store(name, pool)

	logger.Info(fmt.Sprintf("[yiigo] redis.%s is OK", name))
}

// Redis returns a redis pool.
func Redis(name ...string) RedisPool {
	if len(name) == 0 || name[0] == Default {
		if defaultRedis == nil {
			logger.Panic(fmt.Sprintf("[yiigo] unknown redis.%s (forgotten configure?)", Default))
		}

		return defaultRedis
	}

	v, ok := redisMap.Load(name[0])

	if !ok {
		logger.Panic(fmt.Sprintf("[yiigo] unknown redis.%s (forgotten configure?)", name[0]))
	}

	return v.(RedisPool)
}
