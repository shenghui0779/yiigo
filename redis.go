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
	pool := &redis.Pool{
		Dial: func() (redis.Conn, error) {
			host := section.Key("host").MustString("127.0.0.1")
			port := section.Key("port").MustInt(6379)

			dialOptions := []redis.DialOption{
				redis.DialPassword(section.Key("password").MustString("")),
				redis.DialDatabase(section.Key("database").MustInt(0)),
				redis.DialConnectTimeout(section.Key("connTimeout").MustDuration(time.Duration(10000) * time.Millisecond)),
				redis.DialReadTimeout(section.Key("readTimeout").MustDuration(time.Duration(10000) * time.Millisecond)),
				redis.DialWriteTimeout(section.Key("writeTimeout").MustDuration(time.Duration(10000) * time.Millisecond)),
			}

			conn, err := redis.Dial("tcp", fmt.Sprintf("%s:%d", host, port), dialOptions...)

			return conn, err
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			period := section.Key("testOnBorrow").MustDuration(time.Duration(60000) * time.Millisecond)

			if time.Since(t) < period {
				return nil
			}
			_, err := c.Do("PING")

			return err
		},
		MaxIdle:     section.Key("maxIdleConn").MustInt(10),
		MaxActive:   section.Key("maxActiveConn").MustInt(20),
		IdleTimeout: section.Key("idleTimeout").MustDuration(time.Duration(60000) * time.Millisecond),
		Wait:        section.Key("poolWait").MustBool(true),
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
