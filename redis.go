package yiigo

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sync"
	"time"

	"github.com/garyburd/redigo/redis"
	"github.com/youtube/vitess/go/pools"
	"golang.org/x/net/context"
)

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

	if redisPool != nil {
		return
	}

	poolMinActive := EnvInt("redis", "poolMinActive", 10)
	poolMaxActive := EnvInt("redis", "poolMaxActive", 20)
	poolIdleTimeout := EnvDuration("redis", "poolIdleTimeout", time.Duration(60000)*time.Millisecond)

	redisPool = pools.NewResourcePool(func() (pools.Resource, error) {
		conn, err := dialRedis()
		return ResourceConn{conn}, err
	}, poolMinActive, poolMaxActive, poolIdleTimeout*time.Millisecond)
}

// dialRedis dial redis
func dialRedis() (redis.Conn, error) {
	dsn := fmt.Sprintf("%s:%d", EnvString("redis", "host", "localhost"), EnvInt("redis", "port", 6379))

	dialOptions := []redis.DialOption{
		redis.DialPassword(EnvString("redis", "password", "")),
		redis.DialDatabase(EnvInt("redis", "database", 0)),
		redis.DialConnectTimeout(EnvDuration("redis", "connectTimeout", time.Duration(10000)*time.Millisecond)),
		redis.DialReadTimeout(EnvDuration("redis", "readTimeout", time.Duration(10000)*time.Millisecond)),
		redis.DialWriteTimeout(EnvDuration("redis", "writeTimeout", time.Duration(10000)*time.Millisecond)),
	}

	conn, err := redis.Dial("tcp", dsn, dialOptions...)

	if err != nil {
		return nil, err
	}

	return conn, nil
}

// RedisConn get a redis connection
func RedisConn() (redis.Conn, error) {
	if redisPool == nil || redisPool.IsClosed() {
		initRedis()
	}

	ctx := context.TODO()
	r, err := redisPool.Get(ctx)

	if err != nil {
		return nil, err
	}

	defer redisPool.Put(r)

	return r.(ResourceConn).Conn, nil
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
