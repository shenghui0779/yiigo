package yiigo

import (
	"errors"
	"fmt"
	"github.com/garyburd/redigo/redis"
	"github.com/youtube/vitess/go/pools"
	"golang.org/x/net/context"
	"sync"
	"time"
)

type RedisBase struct {
	CacheName string
}

var (
	redisPool    *pools.ResourcePool
	redisPoolMux sync.Mutex
)

type RedisConn struct {
	redis.Conn
}

func (r RedisConn) Close() {
	err := r.Conn.Close()

	if err != nil {
		LogError("redis connection close error: ", err.Error())
	}
}

func initRedis() (redis.Conn, error) {
	host := GetConfigString("redis", "host", "localhost")
	port := GetConfigInt("redis", "port", 6379)
	connectTimeout := GetConfigInt("redis", "connectTimeout", 0)
	readTimeout := GetConfigInt("redis", "readTimeout", 10000)
	writeTimeout := GetConfigInt("redis", "writeTimeout", 10000)

	address := fmt.Sprintf("%s:%d", host, port)
	conn, err := redis.DialTimeout("tcp", address, time.Duration(connectTimeout)*time.Millisecond, time.Duration(readTimeout)*time.Millisecond, time.Duration(writeTimeout)*time.Millisecond)

	if err != nil {
		LogError("init redis error: %s", err.Error())
		return nil, err
	}

	return conn, err
}

func initRedisPool() {
	if redisPool == nil || redisPool.IsClosed() {
		redisPoolMux.Lock()
		defer redisPoolMux.Unlock()

		if redisPool == nil {
			poolMinActive := GetConfigInt("redis", "poolMinActive", 100)
			poolMaxActive := GetConfigInt("redis", "poolMaxActive", 200)
			poolIdleTimeout := GetConfigInt("redis", "poolIdleTimeout", 2000)

			redisPool = pools.NewResourcePool(func() (pools.Resource, error) {
				conn, err := initRedis()
				return RedisConn{conn}, err
			}, poolMinActive, poolMaxActive, time.Duration(poolIdleTimeout)*time.Millisecond)
		}
	}
}

func poolGetRedisConn() (pools.Resource, error) {
	initRedisPool()

	if redisPool == nil {
		LogError("redis pool is null")
		return nil, errors.New("redis pool is null")
	}

	ctx := context.TODO()
	redisResource, err := redisPool.Get(ctx)

	if err != nil {
		LogError("redis get connection err: ", err.Error())
		return nil, err
	}

	if redisResource == nil {
		LogError("redis pool resource is null")
		return nil, errors.New("redis pool resource is null")
	}

	redisConn := redisResource.(RedisConn)

	if redisConn.Conn.Err() != nil {
		LogError("redis resource connection err: ", redisConn.Conn.Err().Error())

		redisConn.Close()
		//连接断开，重新打开
		conn, connErr := initRedis()

		if connErr != nil {
			LogError("redis reconnection err: ", connErr.Error())
			return nil, connErr
		} else {
			return RedisConn{conn}, nil
		}
	}

	return redisResource, nil
}
