package yiigo

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"sync"
	"time"

	ini "gopkg.in/ini.v1"

	"github.com/mediocregopher/radix.v2/pool"
	"github.com/mediocregopher/radix.v2/redis"
)

var (
	// Redis default connection pool
	Redis    *pool.Pool
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
		Redis = v.(*pool.Pool)
	}

	return nil
}

func redisDial(section *ini.Section) (*pool.Pool, error) {
	df := func(network, addr string) (*redis.Client, error) {
		client, err := redis.Dial(network, addr)

		if err != nil {
			return nil, err
		}

		if password := section.Key("password").MustString(""); password != "" {
			// 密码验证
			if err = client.Cmd("AUTH", password).Err; err != nil {
				client.Close()

				return nil, err
			}
		}

		if database := section.Key("database").MustInt(0); database != 0 {
			// 选择数据库
			if err = client.Cmd("SELECT", database).Err; err != nil {
				client.Close()

				return nil, err
			}
		}

		return client, nil
	}

	host := section.Key("password").MustString("127.0.0.1")
	port := section.Key("password").MustInt(6379)
	poolSize := section.Key("poolSize").MustInt(10)

	p, err := pool.NewCustom("tcp", fmt.Sprintf("%s:%d", host, port), poolSize, df)

	if err != nil {
		return nil, err
	}

	// 设置心跳检测
	if poolIdle := section.Key("poolIdle").MustInt(0); poolIdle != 0 {
		go keepalive(p, poolIdle)
	}

	return p, nil
}

func keepalive(p *pool.Pool, idle int) {
	for {
		p.Cmd("PING")
		time.Sleep(time.Duration(idle) * time.Second)
	}
}

// RedisPool get redis connection pool
func RedisPool(conn ...string) (*pool.Pool, error) {
	c := "default"

	if len(conn) > 0 {
		c = conn[0]
	}

	schema := fmt.Sprintf("redis.%s", c)

	v, ok := redisMap.Load(schema)

	if !ok {
		return nil, fmt.Errorf("redis %s is not connected", schema)
	}

	return v.(*pool.Pool), nil
}

// ScanJSON scans src to the struct pointed to by dest
func ScanJSON(reply *redis.Resp, dest interface{}) error {
	bytes, err := reply.Bytes()

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
func ScanJSONSlice(reply *redis.Resp, dest interface{}) error {
	bytes, err := reply.ListBytes()

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
