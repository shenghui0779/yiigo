package yiigo

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/garyburd/redigo/redis"
	"github.com/youtube/vitess/go/pools"
	"golang.org/x/net/context"
)

type Redis struct {
	CacheName string
}

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
 * 初始化Redis连接池
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
 * 连接Redis
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
		return nil, fmt.Errorf("[Redis] %v", err)
	}

	return conn, nil
}

/**
 * 获取Redis资源
 * @return pools.Resource, error
 */
func getRedisConn() (pools.Resource, error) {
	if redisPool == nil {
		return nil, errors.New("[Redis] redis is not initialized")
	}

	if redisPool.IsClosed() {
		initRedis()
	}

	ctx := context.TODO()
	rc, err := redisPool.Get(ctx)

	if err != nil {
		return nil, fmt.Errorf("[Redis] %v", err)
	}

	return rc, err
}

func (r *Redis) getKey(key string) string {
	prefix := GetEnvString("redis", "prefix", "yii")

	if strings.TrimSpace(key) == "" {
		return fmt.Sprintf("%s:%s", prefix, r.CacheName)
	}

	return fmt.Sprintf("%s:%s:%s", prefix, r.CacheName, key)
}

// string cmd

/**
 * SET
 * @param key string
 * @param data interface{}
 * @return error
 */
func (r *Redis) Set(key string, data interface{}) error {
	rc, err := getRedisConn()

	if err != nil {
		return err
	}

	defer redisPool.Put(rc)

	conn := rc.(ResourceConn).Conn

	cacheData, err := json.Marshal(data)

	if err != nil {
		return fmt.Errorf("[Redis] %v", err)
	}

	cacheKey := r.getKey(key)

	_, err = conn.Do("SET", cacheKey, cacheData)

	if err != nil {
		return fmt.Errorf("[Redis] %v", err)
	}

	return nil
}

/**
 * MSET
 * @param data map[string]interface{}
 * @return error
 */
func (r *Redis) MSet(data map[string]interface{}) error {
	rc, err := getRedisConn()

	if err != nil {
		return err
	}

	defer redisPool.Put(rc)

	conn := rc.(ResourceConn).Conn

	args := []interface{}{}

	for k, v := range data {
		cacheData, err := json.Marshal(v)

		if err != nil {
			return fmt.Errorf("[Redis] %v", err)
		}

		args = append(args, r.getKey(k), cacheData)
	}

	_, err = conn.Do("MSET", args...)

	if err != nil {
		return fmt.Errorf("[Redis] %v", err)
	}

	return nil
}

/**
 * GET
 * @param key string
 * @param data interface{} (指针)
 * @return error
 */
func (r *Redis) Get(key string, data interface{}) error {
	rc, err := getRedisConn()

	if err != nil {
		return err
	}

	defer redisPool.Put(rc)

	conn := rc.(ResourceConn).Conn

	cacheKey := r.getKey(key)

	cacheData, err := conn.Do("GET", cacheKey)

	if err != nil {
		return fmt.Errorf("[Redis] %v", err)
	}

	if cacheData == nil {
		return errors.New("[Redis] not found")
	}

	err = json.Unmarshal(cacheData.([]byte), data)

	if err != nil {
		return fmt.Errorf("[Redis] %v", err)
	}

	return nil
}

/**
 * MGET
 * @param keys []string
 * @param data interface{} (切片指针)
 * @return error
 */
func (r *Redis) MGet(keys []string, data interface{}) error {
	rc, err := getRedisConn()

	if err != nil {
		return err
	}

	defer redisPool.Put(rc)

	conn := rc.(ResourceConn).Conn

	args := []interface{}{}

	for _, key := range keys {
		args = append(args, r.getKey(key))
	}

	cacheData, err := redis.ByteSlices(conn.Do("MGET", args...))

	if err != nil {
		return fmt.Errorf("[Redis] %v", err)
	}

	if cacheData == nil {
		return errors.New("[Redis] not found")
	}

	if len(cacheData) > 0 {
		refVal := reflect.Indirect(reflect.ValueOf(data))

		if refVal.Kind() == reflect.Slice {
			refValType := refVal.Type().Elem()
			refVal.Set(reflect.MakeSlice(refVal.Type(), 0, 0))

			for _, v := range cacheData {
				if v != nil {
					elem := reflect.New(refValType).Elem()
					err := json.Unmarshal(v, elem.Addr().Interface())

					if err != nil {
						return fmt.Errorf("[Redis] %v", err)
					}

					refVal.Set(reflect.Append(refVal, elem))
				}
			}
		}
	}

	return nil
}

// hash cmd

/**
 * HSET
 * @param key string
 * @param field interface{}
 * @data interface{}
 * @return error
 */
func (r *Redis) HSet(key string, field interface{}, data interface{}) error {
	rc, err := getRedisConn()

	if err != nil {
		return err
	}

	defer redisPool.Put(rc)

	conn := rc.(ResourceConn).Conn

	cacheData, err := json.Marshal(data)

	if err != nil {
		return fmt.Errorf("[Redis] %v", err)
	}

	cacheKey := r.getKey(key)

	_, err = conn.Do("HSET", cacheKey, field, cacheData)

	if err != nil {
		return fmt.Errorf("[Redis] %v", err)
	}

	return nil
}

/**
 * HMSET
 * @param key string
 * @param data map[interface{}]interface{}
 * @return error
 */
func (r *Redis) HMSet(key string, data map[interface{}]interface{}) error {
	rc, err := getRedisConn()

	if err != nil {
		return err
	}

	defer redisPool.Put(rc)

	conn := rc.(ResourceConn).Conn

	args := []interface{}{}
	args = append(args, r.getKey(key))

	for field, v := range data {
		cacheData, err := json.Marshal(v)

		if err != nil {
			return fmt.Errorf("[Redis] %v", err)
		}

		args = append(args, field, cacheData)
	}

	_, err = conn.Do("HMSet", args...)

	if err != nil {
		return fmt.Errorf("[Redis] %v", err)
	}

	return nil
}

/**
 * HGET
 * @param key string
 * @param field interface{}
 * @param data interface{} (指针)
 * @return error
 */
func (r *Redis) HGet(key string, field interface{}, data interface{}) error {
	rc, err := getRedisConn()

	if err != nil {
		return err
	}

	defer redisPool.Put(rc)

	conn := rc.(ResourceConn).Conn

	cacheKey := r.getKey(key)

	cacheData, err := conn.Do("HGET", cacheKey, field)

	if err != nil {
		return fmt.Errorf("[Redis] %v", err)
	}

	if cacheData == nil {
		return errors.New("[Redis] not found")
	}

	err = json.Unmarshal(cacheData.([]byte), data)

	if err != nil {
		return fmt.Errorf("[Redis] %v", err)
	}

	return nil
}

/**
 * HMGET
 * @param key string
 * @param fields []interface{}
 * @param data interface{} (切片指针)
 * @return error
 */
func (r *Redis) HMGet(key string, fields []interface{}, data interface{}) error {
	rc, err := getRedisConn()

	if err != nil {
		return err
	}

	defer redisPool.Put(rc)

	conn := rc.(ResourceConn).Conn

	args := []interface{}{}
	args = append(args, r.getKey(key))

	for _, field := range fields {
		args = append(args, field)
	}

	cacheData, err := redis.ByteSlices(conn.Do("HMGET", args...))

	if err != nil {
		return fmt.Errorf("[Redis] %v", err)
	}

	if cacheData == nil {
		return errors.New("[Redis] not found")
	}

	if len(cacheData) > 0 {
		refVal := reflect.Indirect(reflect.ValueOf(data))

		if refVal.Kind() == reflect.Slice {
			refValType := refVal.Type().Elem()
			refVal.Set(reflect.MakeSlice(refVal.Type(), 0, 0))

			for _, v := range cacheData {
				if v != nil {
					elem := reflect.New(refValType).Elem()
					err := json.Unmarshal(v, elem.Addr().Interface())

					if err != nil {
						return fmt.Errorf("[Redis] %v", err)
					}

					refVal.Set(reflect.Append(refVal, elem))
				}
			}
		}
	}

	return nil
}

/**
 * HDEL
 * @param key
 * @param field interface{}
 * @return error
 */
func (r *Redis) HDel(key string, field interface{}) error {
	rc, err := getRedisConn()

	if err != nil {
		return err
	}

	defer redisPool.Put(rc)

	conn := rc.(ResourceConn).Conn

	cacheKey := r.getKey(key)

	_, err = conn.Do("HDEL", cacheKey, field)

	if err != nil {
		return fmt.Errorf("[Redis] %v", err)
	}

	return nil
}

/**
 * HLEN
 * @param key string
 * @return int64, error
 */
func (r *Redis) HLen(key string) (int64, error) {
	rc, err := getRedisConn()

	if err != nil {
		return 0, err
	}

	defer redisPool.Put(rc)

	conn := rc.(ResourceConn).Conn

	cacheKey := r.getKey(key)

	result, err := conn.Do("HLEN", cacheKey)

	if err != nil {
		return 0, fmt.Errorf("[Redis] %v", err)
	}

	count, ok := result.(int64)

	if !ok {
		return 0, fmt.Errorf("[Redis] invalid type assertion, result %v is %v", result, reflect.TypeOf(result))
	}

	return count, nil
}

/**
 * HINCRBY
 * @param key string
 * @param field interface{}
 * @param inc int
 * @return int64, error
 */
func (r *Redis) HIncrBy(key string, field interface{}, inc int) (int64, error) {
	rc, err := getRedisConn()

	if err != nil {
		return 0, err
	}

	defer redisPool.Put(rc)

	conn := rc.(ResourceConn).Conn

	cacheKey := r.getKey(key)

	result, err := conn.Do("HINCRBY", cacheKey, field, inc)

	if err != nil {
		return 0, fmt.Errorf("[Redis] %v", err)
	}

	count, ok := result.(int64)

	if !ok {
		return 0, fmt.Errorf("[Redis] invalid type assertion, result %v is %v", result, reflect.TypeOf(result))
	}

	return count, nil
}

// list cmd

/**
 * LPUSH
 * @param key string
 * @param data interface{}
 * @return error
 */
func (r *Redis) LPush(key string, data interface{}) error {
	rc, err := getRedisConn()

	if err != nil {
		return err
	}

	defer redisPool.Put(rc)

	conn := rc.(ResourceConn).Conn

	cacheData, err := json.Marshal(data)

	if err != nil {
		return fmt.Errorf("[Redis] %v", err)
	}

	cacheKey := r.getKey(key)

	_, err = conn.Do("LPUSH", cacheKey, cacheData)

	if err != nil {
		return fmt.Errorf("[Redis] %v", err)
	}

	return nil
}

/**
 * LPOP
 * @param key string
 * @param data interface{} (指针)
 * @return error
 */
func (r *Redis) LPop(key string, data interface{}) error {
	rc, err := getRedisConn()

	if err != nil {
		return err
	}

	defer redisPool.Put(rc)

	conn := rc.(ResourceConn).Conn

	cacheKey := r.getKey(key)

	cacheData, err := conn.Do("LPOP", cacheKey)

	if err != nil {
		return fmt.Errorf("[Redis] %v", err)
	}

	if cacheData == nil {
		return errors.New("[Redis] not found")
	}

	err = json.Unmarshal(cacheData.([]byte), data)

	if err != nil {
		return fmt.Errorf("[Redis] %v", err)
	}

	return nil
}

/**
 * RPUSH
 * @param key string
 * @param data interface{}
 * @return error
 */
func (r *Redis) RPush(key string, data interface{}) error {
	rc, err := getRedisConn()

	if err != nil {
		return err
	}

	defer redisPool.Put(rc)

	conn := rc.(ResourceConn).Conn

	cacheData, err := json.Marshal(data)

	if err != nil {
		return fmt.Errorf("[Redis] %v", err)
	}

	cacheKey := r.getKey(key)

	_, err = conn.Do("RPUSH", cacheKey, cacheData)

	if err != nil {
		return fmt.Errorf("[Redis] %v", err)
	}

	return nil
}

/**
 * RPOP
 * @param key string
 * @param data interface{} (指针)
 * @return error
 */
func (r *Redis) RPop(key string, data interface{}) error {
	rc, err := getRedisConn()

	if err != nil {
		return err
	}

	defer redisPool.Put(rc)

	conn := rc.(ResourceConn).Conn

	cacheKey := r.getKey(key)

	cacheData, err := conn.Do("RPOP", cacheKey)

	if err != nil {
		return fmt.Errorf("[Redis] %v", err)
	}

	if cacheData == nil {
		return errors.New("[Redis] not found")
	}

	err = json.Unmarshal(cacheData.([]byte), data)

	if err != nil {
		return fmt.Errorf("[Redis] %v", err)
	}

	return nil
}

/**
 * LLEN
 * @param key string
 * return int64, error
 */
func (r *Redis) LLen(key string) (int64, error) {
	rc, err := getRedisConn()

	if err != nil {
		return 0, err
	}

	defer redisPool.Put(rc)

	conn := rc.(ResourceConn).Conn

	cacheKey := r.getKey(key)

	result, err := conn.Do("LLEN", cacheKey)

	if err != nil {
		return 0, fmt.Errorf("[Redis] %v", err)
	}

	count, ok := result.(int64)

	if !ok {
		return 0, fmt.Errorf("[Redis] invalid type assertion, result %v is %v", result, reflect.TypeOf(result))
	}

	return count, nil
}

/**
 * LRANGE
 * @param key string
 * @param start int
 * @param end int
 * @param data interface{} (切片指针)
 * @return error
 */
func (r *Redis) LRange(key string, start int, end int, data interface{}) error {
	rc, err := getRedisConn()

	if err != nil {
		return err
	}

	defer redisPool.Put(rc)

	conn := rc.(ResourceConn)

	cacheKey := r.getKey(key)

	cacheData, err := redis.ByteSlices(conn.Do("LRANGE", cacheKey, start, end))

	if err != nil {
		return fmt.Errorf("[Redis] %v", err)
	}

	if cacheData == nil {
		return errors.New("[Redis] not found")
	}

	if len(cacheData) > 0 {
		refVal := reflect.Indirect(reflect.ValueOf(data))

		if refVal.Kind() == reflect.Slice {
			refValType := refVal.Type().Elem()
			refVal.Set(reflect.MakeSlice(refVal.Type(), 0, 0))

			for _, v := range cacheData {
				if v != nil {
					elem := reflect.New(refValType).Elem()
					err := json.Unmarshal(v, elem.Addr().Interface())

					if err != nil {
						return fmt.Errorf("[Redis] %v", err)
					}

					refVal.Set(reflect.Append(refVal, elem))
				}
			}
		}
	}

	return nil
}

// key cmd

/**
 * HMSET
 * @param key string
 * @return error
 */
func (r *Redis) Del(key string) error {
	rc, err := getRedisConn()

	if err != nil {
		return err
	}

	defer redisPool.Put(rc)

	conn := rc.(ResourceConn).Conn

	cacheKey := r.getKey(key)

	_, err = conn.Do("DEL", cacheKey)

	if err != nil {
		return fmt.Errorf("[Redis] %v", err)
	}

	return nil
}

/**
 * EXPIRE
 * @param key string
 * @param time int
 * @return error
 */
func (r *Redis) Expire(key string, time int) error {
	rc, err := getRedisConn()

	if err != nil {
		return err
	}

	defer redisPool.Put(rc)

	conn := rc.(ResourceConn).Conn

	cacheKey := r.getKey(key)

	_, err = conn.Do("EXPIRE", cacheKey, time)

	if err != nil {
		return fmt.Errorf("[Redis] %v", err)
	}

	return nil
}

/**
 * INCR
 * @param key string
 * @return int64, error
 */
func (r *Redis) Incr(key string) (int64, error) {
	rc, err := getRedisConn()

	if err != nil {
		return 0, err
	}

	defer redisPool.Put(rc)

	conn := rc.(ResourceConn).Conn

	cacheKey := r.getKey(key)

	result, err := conn.Do("INCR", cacheKey)

	if err != nil {
		return 0, fmt.Errorf("[Redis] %v", err)
	}

	count, ok := result.(int64)

	if !ok {
		return 0, fmt.Errorf("[Redis] invalid type assertion, result %v is %v", result, reflect.TypeOf(result))
	}

	return count, nil
}

/**
 * INCRBY
 * @param key string
 * @param inc int
 * @return int64, error
 */
func (r *Redis) IncrBy(key string, inc int) (int64, error) {
	rc, err := getRedisConn()

	if err != nil {
		return 0, err
	}

	defer redisPool.Put(rc)

	conn := rc.(ResourceConn).Conn

	cacheKey := r.getKey(key)

	result, err := conn.Do("INCRBY", cacheKey, inc)

	if err != nil {
		return 0, fmt.Errorf("[Redis] %v", err)
	}

	count, ok := result.(int64)

	if !ok {
		return 0, fmt.Errorf("[Redis] invalid type assertion, result %v is %v", result, reflect.TypeOf(result))
	}

	return count, nil
}

/**
 * DECR
 * @param key string
 * @return int64, error
 */
func (r *Redis) Decr(key string) (int64, error) {
	rc, err := getRedisConn()

	if err != nil {
		return 0, err
	}

	defer redisPool.Put(rc)

	conn := rc.(ResourceConn).Conn

	cacheKey := r.getKey(key)

	result, err := conn.Do("DECR", cacheKey)

	if err != nil {
		return 0, fmt.Errorf("[Redis] %v", err)
	}

	count, ok := result.(int64)

	if !ok {
		return 0, fmt.Errorf("[Redis] invalid type assertion, result %v is %v", result, reflect.TypeOf(result))
	}

	return count, nil
}

/**
 * DECRBY
 * @param key string
 * @param inc int
 * @return int64, error
 */
func (r *Redis) DecrBy(key string, inc int) (int64, error) {
	rc, err := getRedisConn()

	if err != nil {
		return 0, err
	}

	defer redisPool.Put(rc)

	conn := rc.(ResourceConn).Conn

	cacheKey := r.getKey(key)

	result, err := conn.Do("DECRBY", cacheKey, inc)

	if err != nil {
		return 0, fmt.Errorf("[Redis] %v", err)
	}

	count, ok := result.(int64)

	if !ok {
		return 0, fmt.Errorf("[Redis] invalid type assertion, result %v is %v", result, reflect.TypeOf(result))
	}

	return count, nil
}
