package yiigo

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/mediocregopher/radix.v2/redis"
	"github.com/pelletier/go-toml"
	"github.com/vitessio/vitess/go/pools"
)

type redisConf struct {
	Name          string `toml:"name"`
	Host          string `toml:"host"`
	Port          int    `toml:"port"`
	Password      string `toml:"password"`
	Database      int    `toml:"database"`
	ConnTimeout   int    `toml:"connTimeout"`
	ReadTimeout   int    `toml:"readTimeout"`
	WriteTimeout  int    `toml:"writeTimeout"`
	MinActiveConn int    `toml:"minActiveConn"`
	MaxActiveConn int    `toml:"maxActiveConn"`
	IdleTimeout   int    `toml:"idleTimeout"`
}

// RedisClient redis connection resource
type RedisClient struct {
	*redis.Client
}

// Close close connection resorce
func (r RedisClient) Close() {
	r.Client.Close()
}

// RedisPoolResource redis pool resource
type RedisPoolResource struct {
	pool   *pools.ResourcePool
	config *redisConf
	mutex  sync.Mutex
}

func (r *RedisPoolResource) dial() (*redis.Client, error) {
	dsn := fmt.Sprintf("%s:%d", r.config.Host, r.config.Port)

	client, err := redis.DialTimeout("tcp", dsn, time.Duration(r.config.ConnTimeout)*time.Second)

	if err != nil {
		return nil, err
	}

	if r.config.Password != "" {
		if err = client.Cmd("AUTH", r.config.Password).Err; err != nil {
			client.Close()

			return nil, err
		}
	}

	if r.config.Database != 0 {
		if err = client.Cmd("SELECT", r.config.Database).Err; err != nil {
			client.Close()

			return nil, err
		}
	}

	return client, nil
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

		return RedisClient{conn}, nil
	}

	r.pool = pools.NewResourcePool(df, r.config.MinActiveConn, r.config.MaxActiveConn, time.Duration(r.config.IdleTimeout)*time.Second)
}

// Get get a connection resource from the pool.
func (r *RedisPoolResource) Get() (RedisClient, error) {
	if r.pool.IsClosed() {
		r.initPool()
	}

	ctx := context.TODO()
	resource, err := r.pool.Get(ctx)

	if err != nil {
		return RedisClient{}, err
	}

	rc := resource.(RedisClient)

	// if rc is error, close and reconnect
	if rc.LastCritical != nil {
		client, err := r.dial()

		if err != nil {
			r.pool.Put(rc)

			return rc, err
		}

		rc.Close()

		return RedisClient{client}, nil
	}

	return rc, nil
}

// Put returns a connection resource to the pool.
func (r *RedisPoolResource) Put(rc RedisClient) {
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
		conf := &redisConf{}
		err := node.Unmarshal(conf)

		if err != nil {
			return err
		}

		initSingleRedis(conf)
	case []*toml.Tree:
		conf := make([]*redisConf, 0, len(node))

		for _, v := range node {
			c := &redisConf{}
			err := v.Unmarshal(c)

			if err != nil {
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
