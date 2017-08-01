package yiigo

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"sync"
	"time"

	"github.com/garyburd/redigo/redis"
	"github.com/youtube/vitess/go/pools"
	"golang.org/x/net/context"
)

type Redis struct{}

var (
	redisPool *pools.ResourcePool
	redisMux  sync.Mutex
)

type ResourceConn struct {
	redis.Conn
}

// 关闭连接资源
func (r ResourceConn) Close() {
	r.Conn.Close()
}

/**
 * 初始化 Redis 连接池
 */
func initRedis() {
	redisMux.Lock()
	defer redisMux.Unlock()

	if redisPool == nil {
		poolMinActive := GetEnvInt("redis", "poolMinActive", 10)
		poolMaxActive := GetEnvInt("redis", "poolMaxActive", 20)
		poolIdleTimeout := GetEnvInt("redis", "poolIdleTimeout", 60000)

		redisPool = pools.NewResourcePool(func() (pools.Resource, error) {
			conn, err := dialRedis()
			return ResourceConn{conn}, err
		}, poolMinActive, poolMaxActive, time.Duration(poolIdleTimeout)*time.Millisecond)
	}
}

/**
 * 连接 Redis
 * @return redis.Conn, error
 */
func dialRedis() (redis.Conn, error) {
	host := GetEnvString("redis", "host", "localhost")
	port := GetEnvInt("redis", "port", 6379)
	connectTimeout := GetEnvInt("redis", "connectTimeout", 10000)
	readTimeout := GetEnvInt("redis", "readTimeout", 10000)
	writeTimeout := GetEnvInt("redis", "writeTimeout", 10000)

	dsn := fmt.Sprintf("%s:%d", host, port)
	conn, err := redis.DialTimeout("tcp", dsn, time.Duration(connectTimeout)*time.Millisecond, time.Duration(readTimeout)*time.Millisecond, time.Duration(writeTimeout)*time.Millisecond)

	if err != nil {
		return nil, err
	}

	return conn, nil
}

/**
 * 获取 Redis 连接资源
 * @return pools.Resource, error
 */
func (r *Redis) getConn() (pools.Resource, error) {
	if redisPool == nil {
		return nil, errors.New("redis pool is empty")
	}

	if redisPool.IsClosed() {
		initRedis()
	}

	ctx := context.TODO()
	rc, err := redisPool.Get(ctx)

	if err != nil {
		return nil, err
	}

	return rc, err
}

/**
 * Do 执行一条redis命令
 * @param cmd string
 * @param args ...interface{}
 * @return interface{}, error
 */
func (r *Redis) Do(cmd string, args ...interface{}) (interface{}, error) {
	rc, err := r.getConn()

	if err != nil {
		return nil, err
	}

	defer redisPool.Put(rc)

	conn := rc.(ResourceConn).Conn

	reply, err := conn.Do(cmd, args...)

	return reply, err
}

/**
 * Pipeline redis管道 执行一组redis命令
 * @param cmds map[string][]interface{}
 * @return interface{}, error
 */
func (r *Redis) Pipeline(cmds map[string][]interface{}) (interface{}, error) {
	rc, err := r.getConn()

	if err != nil {
		return nil, err
	}

	defer redisPool.Put(rc)

	conn := rc.(ResourceConn).Conn

	for k, v := range cmds {
		conn.Send(k, v...)
	}

	reply, err := conn.Do("EXEC")

	return reply, err
}

/**
 * ScanJSONSlice 获取json切片缓存值
 * @param reply interface{}
 * @param dest interface{} (切片指针)
 * @return error
 */
func (r *Redis) ScanJSONSlice(reply interface{}, dest interface{}) error {
	bytes, err := redis.ByteSlices(reply, nil)

	if err != nil {
		return err
	}

	if len(bytes) > 0 {
		rv := reflect.Indirect(reflect.ValueOf(dest))

		if rv.Kind() == reflect.Slice {
			rt := rv.Type().Elem()
			rv.Set(reflect.MakeSlice(rv.Type(), 0, 0))

			for _, v := range bytes {
				if v != nil {
					elem := reflect.New(rt).Elem()
					err := json.Unmarshal(v, elem.Addr().Interface())

					if err != nil {
						return err
					}

					rv.Set(reflect.Append(rv, elem))
				}
			}
		}
	}

	return nil
}
