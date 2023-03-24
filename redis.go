package yiigo

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"runtime/debug"
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
		logger.Error("err conn closed", zap.Error(err))
	}
}

// RedisPool redis pool resource
type RedisPool interface {
	// Get returns a connection resource from the pool.
	// Context with timeout can specify the wait timeout for pool.
	Get(ctx context.Context) (*RedisConn, error)

	// Put returns a connection resource to the pool.
	Put(rc *RedisConn)

	// Do sends a command to the server and returns the received reply.
	Do(ctx context.Context, cmd string, args ...any) (any, error)

	// DoFunc sends commands to the server with callback function.
	DoFunc(ctx context.Context, f func(ctx context.Context, conn *RedisConn) error) error
}

// RedisConfig keeps the settings to setup redis connection.
type RedisConfig struct {
	// Addr host:port address.
	Addr string `json:"addr"`

	// Options optional settings to setup redis connection.
	Options *RedisOptions `json:"options"`
}

// RedisOptions optional settings to setup redis connection.
type RedisOptions struct {
	// Dialer is a custom dial function for creating TCP connections,
	// otherwise a net.Dialer customized via the other options is used.
	Dialer func(ctx context.Context, network, addr string) (net.Conn, error) `json:"dialer"`

	// Username to be used when connecting to the Redis server when Redis ACLs are used.
	Username string `json:"username"`

	// Password to be used when connecting to the Redis server.
	Password string `json:"password"`

	// Database to be selected when dialing a connection.
	Database int `json:"database"`

	// ConnTimeout is the timeout for connecting to the Redis server.
	// Use value -1 for no timeout and 0 for default.
	// Default is 10 seconds.
	ConnTimeout time.Duration `json:"conn_timeout"`

	// ReadTimeout is the timeout for reading a single command reply.
	// Use value -1 for no timeout and 0 for default.
	// Default is 10 seconds.
	ReadTimeout time.Duration `json:"read_timeout"`

	// WriteTimeout is the timeout for writing a single command.
	// Use value -1 for no timeout and 0 for default.
	// Default is 10 seconds.
	WriteTimeout time.Duration `json:"write_timeout"`

	// PoolSize is the maximum number of possible resources in the pool.
	// Use value -1 for no timeout and 0 for default.
	// Default is 10.
	PoolSize int `json:"pool_size"`

	// PoolPrefill is the number of resources to be pre-filled in the pool.
	// Default is no pre-filled.
	PoolPrefill int `json:"pool_prefill"`

	// IdleTimeout is the amount of time after which client closes idle connections.
	// Use value -1 for no timeout and 0 for default.
	// Default is 5 minutes.
	IdleTimeout time.Duration `json:"idle_timeout"`

	// TLSConfig to be used when a TLS connection is dialed.
	TLSConfig *tls.Config `json:"tls_config"`
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
	config *RedisConfig
	pool   *vitess_pool.ResourcePool
	mutex  sync.Mutex
}

func (rp *redisResourcePool) dial() (redis.Conn, error) {
	dialOptions := []redis.DialOption{
		redis.DialDatabase(rp.config.Options.Database),
		redis.DialConnectTimeout(rp.config.Options.ConnTimeout),
		redis.DialReadTimeout(rp.config.Options.ReadTimeout),
		redis.DialWriteTimeout(rp.config.Options.WriteTimeout),
	}

	if len(rp.config.Options.Username) != 0 {
		dialOptions = append(dialOptions, redis.DialUsername(rp.config.Options.Username))
	}

	if len(rp.config.Options.Password) != 0 {
		dialOptions = append(dialOptions, redis.DialPassword(rp.config.Options.Password))
	}

	if rp.config.Options.Dialer != nil {
		dialOptions = append(dialOptions, redis.DialContextFunc(rp.config.Options.Dialer))
	}

	if rp.config.Options.TLSConfig != nil {
		dialOptions = append(dialOptions, redis.DialTLSConfig(rp.config.Options.TLSConfig))
	}

	conn, err := redis.Dial("tcp", rp.config.Addr, dialOptions...)

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

	rp.pool = vitess_pool.NewResourcePool(df, rp.config.Options.PoolSize, rp.config.Options.PoolSize, rp.config.Options.IdleTimeout, rp.config.Options.PoolPrefill)
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
		logger.Warn("err pool conn, reconnect", zap.Error(err))

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

func (rp *redisResourcePool) Do(ctx context.Context, cmd string, args ...any) (any, error) {
	conn, err := rp.Get(ctx)

	if err != nil {
		return nil, err
	}

	defer rp.Put(conn)

	return conn.Do(cmd, args...)
}

func (rp *redisResourcePool) DoFunc(ctx context.Context, f func(ctx context.Context, conn *RedisConn) error) error {
	conn, err := rp.Get(ctx)

	if err != nil {
		return err
	}

	defer func() {
		rp.Put(conn)

		if r := recover(); r != nil {
			logger.Error("redis do func panic", zap.Any("error", r), zap.ByteString("stack", debug.Stack()))
		}
	}()

	return f(ctx, conn)
}

var (
	defaultRedis RedisPool
	redisMap     sync.Map
)

func newRedisPool(cfg *RedisConfig) RedisPool {
	pool := &redisResourcePool{
		config: &RedisConfig{
			Addr: cfg.Addr,
			Options: &RedisOptions{
				ConnTimeout:  10 * time.Second,
				ReadTimeout:  10 * time.Second,
				WriteTimeout: 10 * time.Second,
				PoolSize:     10,
				IdleTimeout:  5 * time.Minute,
			},
		},
	}

	if cfg.Options != nil {
		pool.config.Options.rebuild(cfg.Options)
	}

	pool.init()

	return pool
}

func initRedis(name string, cfg *RedisConfig) {
	pool := newRedisPool(cfg)

	// verify connection
	conn, err := pool.Get(context.TODO())

	if err != nil {
		logger.Panic(fmt.Sprintf("err redis.%s pool", name), zap.String("addr", cfg.Addr), zap.Error(err))
	}

	if _, err = conn.Do("PING"); err != nil {
		conn.Close()

		logger.Panic(fmt.Sprintf("err redis.%s ping", name), zap.String("addr", cfg.Addr), zap.Error(err))
	}

	pool.Put(conn)

	if name == Default {
		defaultRedis = pool
	}

	redisMap.Store(name, pool)

	logger.Info(fmt.Sprintf("redis.%s is OK", name))
}

// Redis returns a redis pool.
func Redis(name ...string) RedisPool {
	if len(name) == 0 || name[0] == Default {
		if defaultRedis == nil {
			logger.Panic(fmt.Sprintf("unknown redis.%s (forgotten configure?)", Default))
		}

		return defaultRedis
	}

	v, ok := redisMap.Load(name[0])

	if !ok {
		logger.Panic(fmt.Sprintf("unknown redis.%s (forgotten configure?)", name[0]))
	}

	return v.(RedisPool)
}
