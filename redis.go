package yiigo

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/gomodule/redigo/redis"
	toml "github.com/pelletier/go-toml"
	"golang.org/x/net/context"
	"vitess.io/vitess/go/pools"
)

type redisConf struct {
	Name         string `toml:"name"`
	Host         string `toml:"host"`
	Port         int    `toml:"port"`
	Password     string `toml:"password"`
	Database     int    `toml:"database"`
	ConnTimeout  int    `toml:"connTimeout"`
	ReadTimeout  int    `toml:"readTimeout"`
	WriteTimeout int    `toml:"writeTimeout"`
	PoolSize     int    `toml:"poolSize"`
	PoolLimit    int    `toml:"poolLimit"`
	IdleTimeout  int    `toml:"idleTimeout"`
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
	pool   *pools.ResourcePool
	config *redisConf
	mutex  sync.Mutex
}

func (r *RedisPoolResource) dial() (redis.Conn, error) {
	dsn := fmt.Sprintf("%s:%d", r.config.Host, r.config.Port)

	dialOptions := []redis.DialOption{
		redis.DialPassword(r.config.Password),
		redis.DialDatabase(r.config.Database),
		redis.DialConnectTimeout(time.Duration(r.config.ConnTimeout) * time.Second),
		redis.DialReadTimeout(time.Duration(r.config.ReadTimeout) * time.Second),
		redis.DialWriteTimeout(time.Duration(r.config.WriteTimeout) * time.Second),
	}

	conn, err := redis.Dial("tcp", dsn, dialOptions...)

	return conn, err
}

func (r *RedisPoolResource) initPool() {
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

	r.pool = pools.NewResourcePool(df, r.config.PoolSize, r.config.PoolLimit, time.Duration(r.config.IdleTimeout)*time.Second)
}

// Get get a connection resource from the pool.
func (r *RedisPoolResource) Get() (RedisConn, error) {
	if r.pool.IsClosed() {
		r.initPool()
	}

	resource, err := r.pool.Get(context.TODO())

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
	// Redis default connection pool
	Redis    *RedisPoolResource
	redisMap sync.Map
)

func initRedis() error {
	result := Env.Get("redis")

	if result == nil {
		return nil
	}

	switch node := result.(type) {
	case *toml.Tree:
		conf := new(redisConf)

		if err := node.Unmarshal(conf); err != nil {
			return err
		}

		initSingleRedis(conf)
	case []*toml.Tree:
		conf := make([]*redisConf, 0, len(node))

		for _, v := range node {
			c := new(redisConf)

			if err := v.Unmarshal(c); err != nil {
				return err
			}

			conf = append(conf, c)
		}

		initMultiRedis(conf)
	default:
		return errors.New("yiigo: invalid redis config")
	}

	return nil
}

func initSingleRedis(conf *redisConf) {
	Redis = &RedisPoolResource{config: conf}
	Redis.initPool()

	redisMap.Store("default", Redis)
}

func initMultiRedis(conf []*redisConf) {
	for _, v := range conf {
		poolResource := &RedisPoolResource{config: v}
		poolResource.initPool()

		redisMap.Store(v.Name, poolResource)
	}

	if v, ok := redisMap.Load("default"); ok {
		Redis = v.(*RedisPoolResource)
	}
}

// RedisPool returns a redis pool.
func RedisPool(conn ...string) (*RedisPoolResource, error) {
	schema := "default"

	if len(conn) > 0 {
		schema = conn[0]
	}

	v, ok := redisMap.Load(schema)

	if !ok {
		return nil, fmt.Errorf("yiigo: redis.%s is not connected", schema)
	}

	return v.(*RedisPoolResource), nil
}
