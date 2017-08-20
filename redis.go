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

// Close close connection resorce
func (r ResourceConn) Close() {
	r.Conn.Close()
}

// initRedis init redis pool
func initRedis() {
	redisMux.Lock()
	defer redisMux.Unlock()

	if redisPool == nil {
		poolMinActive := EnvInt("redis", "poolMinActive", 10)
		poolMaxActive := EnvInt("redis", "poolMaxActive", 20)
		poolIdleTimeout := EnvInt("redis", "poolIdleTimeout", 60000)

		redisPool = pools.NewResourcePool(func() (pools.Resource, error) {
			conn, err := dialRedis()
			return ResourceConn{conn}, err
		}, poolMinActive, poolMaxActive, time.Duration(poolIdleTimeout)*time.Millisecond)
	}
}

// dialRedis dial redis
func dialRedis() (redis.Conn, error) {
	host := EnvString("redis", "host", "localhost")
	port := EnvInt("redis", "port", 6379)
	connectTimeout := EnvInt("redis", "connectTimeout", 10000)
	readTimeout := EnvInt("redis", "readTimeout", 10000)
	writeTimeout := EnvInt("redis", "writeTimeout", 10000)

	dsn := fmt.Sprintf("%s:%d", host, port)
	conn, err := redis.DialTimeout("tcp", dsn, time.Duration(connectTimeout)*time.Millisecond, time.Duration(readTimeout)*time.Millisecond, time.Duration(writeTimeout)*time.Millisecond)

	if err != nil {
		return nil, err
	}

	return conn, nil
}

// Redis get redis
func Redis() *Redis {
	return &Redis{}
}

// getConn get a redis connection
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

// Do sends a command to the server and returns the received reply.
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

// Pipeline sends commands to the server and returns the received reply
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

// ScanRedisJSON scans json to a struct
func ScanRedisJSON(reply interface{}, dest interface{}) error {
	bytes, err := redis.Bytes(reply, nil)

	if err != nil {
		return err
	}

	err = json.Unmarshal(bytes, dest)

	if err != nil {
		return err
	}

	return nil
}

// ScanRedisJSONSlice scans json slice to a struct slice
func ScanRedisJSONSlice(reply interface{}, dest interface{}) error {
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
