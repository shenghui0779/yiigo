package yiigo

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"sync"
	"time"

	ini "gopkg.in/ini.v1"

	"github.com/garyburd/redigo/redis"
	"github.com/youtube/vitess/go/pools"
	"golang.org/x/net/context"
)

type redisPool struct {
	name string
	pool *pools.ResourcePool
	mux  sync.Mutex
}

type ResourceConn struct {
	redis.Conn
}

var (
	RedisPool *redisPool
	redisMap  map[string]*redisPool
	redisMux  sync.RWMutex
)

// Close close connection resorce
func (r ResourceConn) Close() {
	r.Conn.Close()
}

func initRedis() {
	sections := childSections("redis")

	if len(sections) > 0 {
		initMultiRedis(sections)
		return
	}

	initSingleRedis()
}

func initSingleRedis() {
	RedisPool := &redisPool{name: "redis"}
	RedisPool.dial()
}

func initMultiRedis(sections []*ini.Section) {
	redisMap = make(map[string]*redisPool, len(sections))

	for _, v := range sections {
		pool := &redisPool{name: v.Name()}
		pool.dial()

		redisMap[v.Name()] = pool
	}

	if redis, ok := redisMap["redis.default"]; ok {
		RedisPool = redis
	}
}

// RedisConn get redis connection
func RedisConn(conn ...string) (redis.Conn, error) {
	redisMux.RLock()
	defer redisMux.RUnlock()

	c := "default"

	if len(conn) > 0 {
		c = conn[0]
	}

	schema := fmt.Sprintf("redis.%s", c)

	rp, ok := redisMap[schema]

	if !ok {
		return nil, fmt.Errorf("redis %s is not connected", schema)
	}

	rc, err := rp.get()

	if err != nil {
		return nil, err
	}

	defer rp.pool.Put(rc)

	return rc.(ResourceConn).Conn, nil
}

func (r *redisPool) dial() {
	r.mux.Lock()
	defer r.mux.Unlock()

	if r.pool != nil {
		return
	}

	poolMinActive := EnvInt(r.name, "poolMinActive", 10)
	poolMaxActive := EnvInt(r.name, "poolMaxActive", 20)
	poolIdleTimeout := EnvDuration(r.name, "poolIdleTimeout", time.Duration(60000)*time.Millisecond)

	r.pool = pools.NewResourcePool(func() (pools.Resource, error) {
		dsn := fmt.Sprintf("%s:%d", EnvString(r.name, "host", "localhost"), EnvInt("redis", "port", 6379))

		dialOptions := []redis.DialOption{
			redis.DialPassword(EnvString(r.name, "password", "")),
			redis.DialDatabase(EnvInt(r.name, "database", 0)),
			redis.DialConnectTimeout(EnvDuration(r.name, "connectTimeout", time.Duration(10000)*time.Millisecond)),
			redis.DialReadTimeout(EnvDuration(r.name, "readTimeout", time.Duration(10000)*time.Millisecond)),
			redis.DialWriteTimeout(EnvDuration(r.name, "writeTimeout", time.Duration(10000)*time.Millisecond)),
		}

		conn, err := redis.Dial("tcp", dsn, dialOptions...)

		if err != nil {
			return nil, err
		}

		return ResourceConn{conn}, nil
	}, poolMinActive, poolMaxActive, poolIdleTimeout*time.Millisecond)
}

func (r *redisPool) get() (pools.Resource, error) {
	if r.pool.IsClosed() {
		r.dial()
	}

	ctx := context.TODO()
	rc, err := r.pool.Get(ctx)

	if err != nil {
		return nil, err
	}

	return rc, nil
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
