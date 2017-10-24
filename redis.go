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

// ScanJSON scans src to the struct pointed to by dest
func ScanJSON(reply interface{}, dest interface{}) error {
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

// ScanJSONSlice scans src to the slice pointed to by dest
func ScanJSONSlice(reply interface{}, dest interface{}) error {
	bytes, err := redis.ByteSlices(reply, nil)

	if err != nil {
		return err
	}

	if len(bytes) == 0 {
		return nil
	}

	v := reflect.Indirect(reflect.ValueOf(dest))

	if v.Kind() != reflect.Slice {
		return errors.New("the dest must be a slice")
	}

	t := v.Type()
	v.Set(reflect.MakeSlice(t, 0, 0))

	for _, b := range bytes {
		elem := reflect.New(t.Elem()).Elem()
		err := json.Unmarshal(b, elem.Addr().Interface())

		if err != nil {
			return err
		}

		v.Set(reflect.Append(v, elem))
	}

	return nil
}
