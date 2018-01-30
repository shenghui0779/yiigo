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
)

var (
	// Redis default connection pool
	Redis    *redis.Pool
	redisMap sync.Map
)

func initRedis() error {
	sections := childSections("redis")

	if len(sections) > 0 {
		return initMultiRedis(sections)
	}

	return initSingleRedis()
}

func initSingleRedis() error {
	var err error

	section := env.Section("redis")

	Redis, err = redisDial(section)

	if err != nil {
		return fmt.Errorf("redis error: %s", err.Error())
	}

	return nil
}

func initMultiRedis(sections []*ini.Section) error {
	for _, v := range sections {
		p, err := redisDial(v)

		if err != nil {
			return fmt.Errorf("redis error: %s", err.Error())
		}

		redisMap.Store(v.Name(), p)
	}

	if v, ok := redisMap.Load("redis.default"); ok {
		Redis = v.(*redis.Pool)
	}

	return nil
}

func redisDial(section *ini.Section) (*redis.Pool, error) {
	host := section.Key("host").MustString("127.0.0.1")
	port := section.Key("port").MustInt(6379)
	password := section.Key("password").MustString("")
	database := section.Key("database").MustInt(0)
	connTimeout := section.Key("connTimeout").MustInt(1000)
	readTimeout := section.Key("readTimeout").MustInt(1000)
	writeTimeout := section.Key("writeTimeout").MustInt(1000)
	testOnBorrow := section.Key("testOnBorrow").MustInt(0)
	maxIdleConn := section.Key("maxIdleConn").MustInt(10)
	maxActiveConn := section.Key("maxActiveConn").MustInt(20)
	idleTimeout := section.Key("idleTimeout").MustInt(60000)
	poolWait := section.Key("poolWait").MustBool(false)

	pool := &redis.Pool{
		Dial: func() (redis.Conn, error) {
			dialOptions := []redis.DialOption{
				redis.DialPassword(password),
				redis.DialDatabase(database),
				redis.DialConnectTimeout(time.Duration(connTimeout) * time.Millisecond),
				redis.DialReadTimeout(time.Duration(readTimeout) * time.Millisecond),
				redis.DialWriteTimeout(time.Duration(writeTimeout) * time.Millisecond),
			}

			conn, err := redis.Dial("tcp", fmt.Sprintf("%s:%d", host, port), dialOptions...)

			return conn, err
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			if testOnBorrow == 0 || time.Since(t) < time.Duration(testOnBorrow)*time.Millisecond {
				return nil
			}
			_, err := c.Do("PING")

			return err
		},
		MaxIdle:     maxIdleConn,
		MaxActive:   maxActiveConn,
		IdleTimeout: time.Duration(idleTimeout) * time.Millisecond,
		Wait:        poolWait,
	}

	conn := pool.Get()
	defer conn.Close()

	_, err := conn.Do("PING")

	if err != nil {
		return nil, err
	}

	return pool, nil
}

// RedisPool get redis connection pool
func RedisPool(conn ...string) (*redis.Pool, error) {
	c := "default"

	if len(conn) > 0 {
		c = conn[0]
	}

	schema := fmt.Sprintf("redis.%s", c)

	v, ok := redisMap.Load(schema)

	if !ok {
		return nil, fmt.Errorf("redis %s is not connected", schema)
	}

	return v.(*redis.Pool), nil
}

// ScanJSON scans json string to the struct or struct slice pointed to by dest
func ScanJSON(reply interface{}, dest interface{}) error {
	v := reflect.Indirect(reflect.ValueOf(dest))

	var err error

	switch v.Kind() {
	case reflect.Struct:
		err = scanJSONObj(reply, dest)
	case reflect.Slice:
		err = scanJSONSlice(reply, dest)
	}

	return err
}

func scanJSONObj(reply interface{}, dest interface{}) error {
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

func scanJSONSlice(reply interface{}, dest interface{}) error {
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
