package yiigo

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/pelletier/go-toml"
	"github.com/shenghui0779/vitess_pool"
	"go.uber.org/zap"
)

type redisConfig struct {
	Address            string `toml:"address"`
	Password           string `toml:"password"`
	Database           int    `toml:"database"`
	ConnTimeout        int    `toml:"conn_timeout"`
	ReadTimeout        int    `toml:"read_timeout"`
	WriteTimeout       int    `toml:"write_timeout"`
	PoolSize           int    `toml:"pool_size"`
	PoolLimit          int    `toml:"pool_limit"`
	IdleTimeout        int    `toml:"idle_timeout"`
	WaitTimeout        int    `toml:"wait_timeout"`
	PrefillParallelism int    `toml:"prefill_parallelism"`
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
	config *redisConfig
	pool   *vitess_pool.ResourcePool
	mutex  sync.Mutex
}

func (r *RedisPoolResource) dial() (redis.Conn, error) {
	dialOptions := []redis.DialOption{
		redis.DialPassword(r.config.Password),
		redis.DialDatabase(r.config.Database),
		redis.DialConnectTimeout(time.Duration(r.config.ConnTimeout) * time.Second),
		redis.DialReadTimeout(time.Duration(r.config.ReadTimeout) * time.Second),
		redis.DialWriteTimeout(time.Duration(r.config.WriteTimeout) * time.Second),
	}

	conn, err := redis.Dial("tcp", r.config.Address, dialOptions...)

	return conn, err
}

func (r *RedisPoolResource) init() {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if r.pool != nil && !r.pool.IsClosed() {
		return
	}

	df := func() (vitess_pool.Resource, error) {
		conn, err := r.dial()

		if err != nil {
			return nil, err
		}

		return RedisConn{conn}, nil
	}

	r.pool = vitess_pool.NewResourcePool(df, r.config.PoolSize, r.config.PoolLimit, time.Duration(r.config.IdleTimeout)*time.Second, r.config.PrefillParallelism)
}

// Get get a connection resource from the pool.
func (r *RedisPoolResource) Get() (RedisConn, error) {
	if r.pool.IsClosed() {
		r.init()
	}

	ctx := context.TODO()

	if r.config.WaitTimeout != 0 {
		var cancel context.CancelFunc

		ctx, cancel = context.WithTimeout(ctx, time.Duration(r.config.WaitTimeout)*time.Second)

		defer cancel()
	}

	resource, err := r.pool.Get(ctx)

	if err != nil {
		return RedisConn{}, err
	}

	rc := resource.(RedisConn)

	// If rc is error, close and reconnect
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
	defaultRedis *RedisPoolResource
	redisMap     sync.Map
)

func initRedis() {
	tree, ok := env.get("redis").(*toml.Tree)

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

		cfg := new(redisConfig)

		if err := node.Unmarshal(cfg); err != nil {
			logger.Panic("yiigo: redis init error", zap.String("name", v), zap.Error(err))
		}

		poolResource := &RedisPoolResource{config: cfg}

		poolResource.init()

		if v == AsDefault {
			defaultRedis = poolResource
		}

		redisMap.Store(v, poolResource)

		logger.Info(fmt.Sprintf("yiigo: redis.%s is OK.", v))
	}
}

// Redis returns a redis pool.
func Redis(name ...string) *RedisPoolResource {
	if len(name) == 0 {
		if defaultRedis == nil {
			logger.Panic(fmt.Sprintf("yiigo: unknown redis.%s (forgotten configure?)", AsDefault))
		}

		return defaultRedis
	}

	v, ok := redisMap.Load(name[0])

	if !ok {
		logger.Panic(fmt.Sprintf("yiigo: unknown redis.%s (forgotten configure?)", name[0]))
	}

	return v.(*RedisPoolResource)
}
