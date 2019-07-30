package yiigo

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/gomodule/redigo/redis"
	"vitess.io/vitess/go/pools"
)

type redisOptions struct {
	password           string
	database           int
	connTimeout        time.Duration
	readTimeout        time.Duration
	writeTimeout       time.Duration
	poolSize           int
	poolLimit          int
	idleTimeout        time.Duration
	waitTimeout        time.Duration
	prefillParallelism int
}

// RedisOption configures how we set up the db
type RedisOption interface {
	apply(options *redisOptions)
}

// funcRedisOption implements redis option
type funcRedisOption struct {
	f func(options *redisOptions)
}

func (fo *funcRedisOption) apply(o *redisOptions) {
	fo.f(o)
}

func newFuncRedisOption(f func(options *redisOptions)) *funcRedisOption {
	return &funcRedisOption{f: f}
}

// WithRedisPassword specifies the `Password` to redis.
func WithRedisPassword(s string) RedisOption {
	return newFuncRedisOption(func(o *redisOptions) {
		o.password = s
	})
}

// WithRedisDatabase specifies the `Database` to redis.
func WithRedisDatabase(n int) RedisOption {
	return newFuncRedisOption(func(o *redisOptions) {
		o.database = n
	})
}

// WithRedisConnTimeout specifies the `ConnTimeout` to redis.
func WithRedisConnTimeout(d time.Duration) RedisOption {
	return newFuncRedisOption(func(o *redisOptions) {
		o.connTimeout = d
	})
}

// WithRedisReadTimeout specifies the `ReadTimeout` to redis.
func WithRedisReadTimeout(d time.Duration) RedisOption {
	return newFuncRedisOption(func(o *redisOptions) {
		o.readTimeout = d
	})
}

// WithRedisWriteTimeout specifies the `WriteTimeout` to redis.
func WithRedisWriteTimeout(d time.Duration) RedisOption {
	return newFuncRedisOption(func(o *redisOptions) {
		o.writeTimeout = d
	})
}

// WithRedisPoolSize specifies the `PoolSize` to redis.
func WithRedisPoolSize(n int) RedisOption {
	return newFuncRedisOption(func(o *redisOptions) {
		o.poolSize = n
	})
}

// WithRedisPoolLimit specifies the `PoolLimit` to redis.
func WithRedisPoolLimit(n int) RedisOption {
	return newFuncRedisOption(func(o *redisOptions) {
		o.poolLimit = n
	})
}

// WithRedisIdleTimeout specifies the `IdleTimeout` to redis.
func WithRedisIdleTimeout(d time.Duration) RedisOption {
	return newFuncRedisOption(func(o *redisOptions) {
		o.idleTimeout = d
	})
}

// WithRedisWaitTimeout specifies the `WaitTimeout` to redis.
// A timeout of 0 means an indefinite wait.
func WithRedisWaitTimeout(d time.Duration) RedisOption {
	return newFuncRedisOption(func(o *redisOptions) {
		o.waitTimeout = d
	})
}

// WithRedisPrefillParallelism specifies the `PrefillParallelism` to redis.
// A non-zero value of prefillParallelism causes the pool to be pre-filled.
func WithRedisPrefillParallelism(n int) RedisOption {
	return newFuncRedisOption(func(o *redisOptions) {
		o.prefillParallelism = n
	})
}

// RedisConn redis connection resource
type RedisConn struct {
	redis.Conn
}

// Close close connection resorce
func (r RedisConn) Close() {
	r.Conn.Close()
}

// RedisPoolResource redis pool resource
type RedisPoolResource struct {
	addr    string
	options *redisOptions
	pool    *pools.ResourcePool
	mutex   sync.Mutex
}

func (r *RedisPoolResource) dial() (redis.Conn, error) {
	dialOptions := []redis.DialOption{
		redis.DialPassword(r.options.password),
		redis.DialDatabase(r.options.database),
		redis.DialConnectTimeout(r.options.connTimeout),
		redis.DialReadTimeout(r.options.readTimeout),
		redis.DialWriteTimeout(r.options.writeTimeout),
	}

	conn, err := redis.Dial("tcp", r.addr, dialOptions...)

	return conn, err
}

func (r *RedisPoolResource) init() {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if r.pool != nil && !r.pool.IsClosed() {
		return
	}

	df := func() (pools.Resource, error) {
		conn, err := r.dial()

		if err != nil {
			return nil, err
		}

		return RedisConn{conn}, nil
	}

	r.pool = pools.NewResourcePool(df, r.options.poolSize, r.options.poolLimit, r.options.idleTimeout, r.options.prefillParallelism)
}

// Get get a connection resource from the pool.
func (r *RedisPoolResource) Get() (RedisConn, error) {
	if r.pool.IsClosed() {
		r.init()
	}

	ctx := context.TODO()

	if r.options.waitTimeout != 0 {
		c, cancel := context.WithTimeout(ctx, r.options.waitTimeout)

		defer cancel()

		ctx = c
	}

	resource, err := r.pool.Get(ctx)

	if err != nil {
		return RedisConn{}, err
	}

	rc := resource.(RedisConn)

	// if rc is error, close and reconnect
	if rc.Err() != nil {
		conn, err := r.dial()

		if err != nil {
			r.pool.Put(rc)

			return rc, err
		}

		rc.Close()

		return RedisConn{conn}, nil
	}

	return rc, nil
}

// Put returns a connection resource to the pool.
func (r *RedisPoolResource) Put(rc RedisConn) {
	r.pool.Put(rc)
}

var (
	// Redis default redis connection pool
	Redis    *RedisPoolResource
	redisMap sync.Map
)

// RegisterRedis register a redis.
//
// The default `ConnTimeout` is 10s.
// The default `ReadTimeout` is 10s.
// The default `WriteTimeout` is 10s.
// The default `PoolSize` is 10.
// The default `PoolLimit` is 20.
// The default `IdleTimeout` is 60s.
// The default `WaitTimeout` is 10s.
// The default `PrefillParallelism` is 0.
func RegisterRedis(name, addr string, options ...RedisOption) {
	o := &redisOptions{
		connTimeout:  10 * time.Second,
		readTimeout:  10 * time.Second,
		writeTimeout: 10 * time.Second,
		poolSize:     10,
		poolLimit:    20,
		idleTimeout:  60 * time.Second,
		waitTimeout:  10 * time.Second,
	}

	if len(options) > 0 {
		for _, option := range options {
			option.apply(o)
		}
	}

	poolResource := &RedisPoolResource{
		addr:    addr,
		options: o,
	}

	poolResource.init()

	redisMap.Store(name, poolResource)

	if name == AsDefault {
		Redis = poolResource
	}
}

// UseRedis returns a redis pool.
func UseRedis(name string) *RedisPoolResource {
	v, ok := redisMap.Load(name)

	if !ok {
		panic(fmt.Errorf("yiigo: redis.%s is not registered", name))
	}

	return v.(*RedisPoolResource)
}
