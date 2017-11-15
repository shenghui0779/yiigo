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

// RedisPoolResource redis pool resource
type RedisPoolResource struct {
	name string
	pool *pools.ResourcePool
	mux  sync.Mutex
}

// RedisResourceConn redis connection resource
type RedisResourceConn struct {
	redis.Conn
}

var (
	// RedisPool default connection pool
	RedisPool *RedisPoolResource
	redisMap  sync.Map
)

// Close close connection resorce
func (r RedisResourceConn) Close() {
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
	RedisPool = &RedisPoolResource{name: "redis"}
	RedisPool.dial()
}

func initMultiRedis(sections []*ini.Section) {
	for _, v := range sections {
		pool := &RedisPoolResource{name: v.Name()}
		pool.dial()

		redisMap.Store(v.Name(), pool)
	}

	if v, ok := redisMap.Load("redis.default"); ok {
		RedisPool = v.(*RedisPoolResource)
	}
}

// RedisConnPool get an redis pool
func RedisConnPool(conn ...string) (*RedisPoolResource, error) {
	c := "default"

	if len(conn) > 0 {
		c = conn[0]
	}

	schema := fmt.Sprintf("redis.%s", c)

	v, ok := redisMap.Load(schema)

	if !ok {
		return nil, fmt.Errorf("redis %s is not connected", schema)
	}

	return v.(*RedisPoolResource), nil
}

func (r *RedisPoolResource) dial() {
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

		return RedisResourceConn{conn}, nil
	}, poolMinActive, poolMaxActive, poolIdleTimeout*time.Millisecond)
}

// Get get a connection resource from the pool
func (r *RedisPoolResource) Get() (RedisResourceConn, error) {
	if r.pool.IsClosed() {
		r.dial()
	}

	ctx := context.TODO()
	resource, err := r.pool.Get(ctx)

	if err != nil {
		return RedisResourceConn{}, err
	}

	rc := resource.(RedisResourceConn)

	if err = rc.Err(); err != nil {
		r.pool.Put(rc)
		return rc, err
	}

	return rc, nil
}

// Put return a connection resource to the pool
func (r *RedisPoolResource) Put(rc RedisResourceConn) {
	r.pool.Put(rc)
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
