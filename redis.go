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
	err := r.Conn.Close()

	if err != nil {
		LogError("[Redis] Close ", err.Error())
	}
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
		LogError("[Redis] ", err.Error())
		panic(err)
	}

	return conn, err
}

/**
 * 获取Redis资源
 * @return pools.Resource, error
 */
func getRedisConn() (pools.Resource, error) {
	if redisPool == nil {
		LogError("[Redis] redis is not initialized")
		panic(errors.New("redis error: redis is not initialized"))
	}

	if redisPool.IsClosed() {
		initRedis()
	}

	ctx := context.TODO()
	rc, err := redisPool.Get(ctx)

	if err != nil {
		LogError("[Redis] ", err.Error())
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
 * @return bool
 */
func (r *Redis) Set(key string, data interface{}) bool {
	rc, err := getRedisConn()

	if err != nil {
		return false
	}

	defer redisPool.Put(rc)

	conn := rc.(ResourceConn).Conn

	cacheData, jsonErr := json.Marshal(data)

	if jsonErr != nil {
		LogError("[Redis] [SET] ", jsonErr.Error())
		return false
	}

	cacheKey := r.getKey(key)

	_, doErr := conn.Do("SET", cacheKey, cacheData)

	if doErr != nil {
		LogError("[Redis] [SET] ", doErr.Error())
		return false
	}

	return true
}

/**
 * MSET
 * @param data map[string]interface{}
 * @return bool
 */
func (r *Redis) MSet(data map[string]interface{}) bool {
	rc, err := getRedisConn()

	if err != nil {
		return false
	}

	defer redisPool.Put(rc)

	conn := rc.(ResourceConn).Conn

	args := []interface{}{}

	for k, v := range data {
		cacheData, jsonErr := json.Marshal(v)

		if jsonErr != nil {
			LogError("[Redis] [SET] ", jsonErr.Error())
			return false
		}

		args = append(args, r.getKey(k), cacheData)
	}

	_, doErr := conn.Do("MSET", args...)

	if doErr != nil {
		LogError("[Redis] [MSET] ", doErr.Error())
		return false
	}

	return true
}

/**
 * GET
 * @param key string
 * @param data interface{} (指针)
 * @return bool
 */
func (r *Redis) Get(key string, data interface{}) bool {
	rc, err := getRedisConn()

	if err != nil {
		return false
	}

	defer redisPool.Put(rc)

	conn := rc.(ResourceConn).Conn

	cacheKey := r.getKey(key)

	cacheData, doErr := conn.Do("GET", cacheKey)

	if doErr != nil {
		LogError("[Redis] [GET] ", doErr.Error())
		return false
	}

	if cacheData == nil {
		return false
	}

	jsonErr := json.Unmarshal(cacheData.([]byte), data)

	if jsonErr != nil {
		LogError("[Redis] [GET] ", jsonErr.Error())
		return false
	}

	return true
}

/**
 * MGET
 * @param keys []string
 * @param data interface{} (切片指针)
 * @return bool
 */
func (r *Redis) MGet(keys []string, data interface{}) bool {
	rc, err := getRedisConn()

	if err != nil {
		return false
	}

	defer redisPool.Put(rc)

	conn := rc.(ResourceConn).Conn

	args := []interface{}{}

	for _, key := range keys {
		args = append(args, r.getKey(key))
	}

	cacheData, doErr := redis.ByteSlices(conn.Do("MGET", args...))

	if doErr != nil {
		LogError("[Redis] [MGET] ", doErr.Error())
		return false
	}

	if cacheData == nil {
		return false
	}

	if len(cacheData) > 0 {
		refVal := reflect.Indirect(reflect.ValueOf(data))

		if refVal.Kind() == reflect.Slice {
			refValType := refVal.Type().Elem()
			refVal.Set(reflect.MakeSlice(refVal.Type(), 0, 0))

			for _, v := range cacheData {
				if v != nil {
					elem := reflect.New(refValType).Elem()
					jsonErr := json.Unmarshal(v, elem.Addr().Interface())

					if jsonErr != nil {
						LogError("[Redis] [MGET] ", jsonErr.Error())
						return false
					}

					refVal.Set(reflect.Append(refVal, elem))
				}
			}
		}
	}

	return true
}

// hash cmd

/**
 * HSET
 * @param key string
 * @param field interface{}
 * @data interface{}
 * @return bool
 */
func (r *Redis) HSet(key string, field interface{}, data interface{}) bool {
	rc, err := getRedisConn()

	if err != nil {
		return false
	}

	defer redisPool.Put(rc)

	conn := rc.(ResourceConn).Conn

	cacheData, jsonErr := json.Marshal(data)

	if jsonErr != nil {
		LogError("[Redis] [HSET] ", jsonErr.Error())
		return false
	}

	cacheKey := r.getKey(key)

	_, doErr := conn.Do("HSET", cacheKey, field, cacheData)

	if doErr != nil {
		LogError("[Redis] [HSET] ", doErr.Error())
		return false
	}

	return true
}

/**
 * HMSET
 * @param key string
 * @param data map[interface{}]interface{}
 * @return bool
 */
func (r *Redis) HMSet(key string, data map[interface{}]interface{}) bool {
	rc, err := getRedisConn()

	if err != nil {
		return false
	}

	defer redisPool.Put(rc)

	conn := rc.(ResourceConn).Conn

	args := []interface{}{}
	args = append(args, r.getKey(key))

	for field, v := range data {
		cacheData, jsonErr := json.Marshal(v)

		if jsonErr != nil {
			LogError("[Redis] [HMSET] ", jsonErr.Error())
			return false
		}

		args = append(args, field, cacheData)
	}

	_, doErr := conn.Do("HMSet", args...)

	if doErr != nil {
		LogError("[Redis] [HMSET] ", doErr.Error())
		return false
	}

	return true
}

/**
 * HGET
 * @param key string
 * @param field interface{}
 * @param data interface{} (指针)
 * @return bool
 */
func (r *Redis) HGet(key string, field interface{}, data interface{}) bool {
	rc, err := getRedisConn()

	if err != nil {
		return false
	}

	defer redisPool.Put(rc)

	conn := rc.(ResourceConn).Conn

	cacheKey := r.getKey(key)

	cacheData, doErr := conn.Do("HGET", cacheKey, field)

	if doErr != nil {
		LogError("[Redis] [HGET] ", doErr.Error())
		return false
	}

	if cacheData == nil {
		return false
	}

	jsonErr := json.Unmarshal(cacheData.([]byte), data)

	if jsonErr != nil {
		LogError("[Redis] [HGET] ", jsonErr.Error())
		return false
	}

	return true
}

/**
 * HMGET
 * @param key string
 * @param fields []interface{}
 * @param data interface{} (切片指针)
 * @return bool
 */
func (r *Redis) HMGet(key string, fields []interface{}, data interface{}) bool {
	rc, err := getRedisConn()

	if err != nil {
		return false
	}

	defer redisPool.Put(rc)

	conn := rc.(ResourceConn).Conn

	args := []interface{}{}
	args = append(args, r.getKey(key))

	for _, field := range fields {
		args = append(args, field)
	}

	cacheData, doErr := redis.ByteSlices(conn.Do("HMGET", args...))

	if doErr != nil {
		LogError("[Redis] [HMGET] ", doErr.Error())
		return false
	}

	if cacheData == nil {
		return false
	}

	if len(cacheData) > 0 {
		refVal := reflect.Indirect(reflect.ValueOf(data))

		if refVal.Kind() == reflect.Slice {
			refValType := refVal.Type().Elem()
			refVal.Set(reflect.MakeSlice(refVal.Type(), 0, 0))

			for _, v := range cacheData {
				if v != nil {
					elem := reflect.New(refValType).Elem()
					jsonErr := json.Unmarshal(v, elem.Addr().Interface())

					if jsonErr != nil {
						LogError("[Redis] [HMGET] ", jsonErr.Error())
						return false
					}

					refVal.Set(reflect.Append(refVal, elem))
				}
			}
		}
	}

	return true
}

/**
 * HDEL
 * @param key
 * @param field interface{}
 * @return bool
 */
func (r *Redis) HDel(key string, field interface{}) bool {
	rc, err := getRedisConn()

	if err != nil {
		return false
	}

	defer redisPool.Put(rc)

	conn := rc.(ResourceConn).Conn

	cacheKey := r.getKey(key)

	_, doErr := conn.Do("HDEL", cacheKey, field)

	if doErr != nil {
		LogError("[Redis] [HDEL] ", doErr.Error())
		return false
	}

	return true
}

/**
 * HLEN
 * @param key string
 * @return int64, bool
 */
func (r *Redis) HLen(key string) (int64, bool) {
	rc, err := getRedisConn()

	if err != nil {
		return 0, false
	}

	defer redisPool.Put(rc)

	conn := rc.(ResourceConn).Conn

	cacheKey := r.getKey(key)

	result, doErr := conn.Do("HLEN", cacheKey)

	if doErr != nil {
		LogError("[Redis] [HLEN] ", doErr.Error())
		return 0, false
	}

	count, ok := result.(int64)

	if !ok {
		LogErrorf("[Redis] [HLEN] invalid type assertion, result %v is %v", result, reflect.TypeOf(result))
		return 0, false
	}

	return count, true
}

/**
 * HINCRBY
 * @param key string
 * @param field interface{}
 * @param inc int
 * @return int64, bool
 */
func (r *Redis) HIncrBy(key string, field interface{}, inc int) (int64, bool) {
	rc, err := getRedisConn()

	if err != nil {
		return 0, false
	}

	defer redisPool.Put(rc)

	conn := rc.(ResourceConn).Conn

	cacheKey := r.getKey(key)

	result, doErr := conn.Do("HINCRBY", cacheKey, field, inc)

	if doErr != nil {
		LogError("[Redis] [HINCRBY] ", doErr.Error())
		return 0, false
	}

	count, ok := result.(int64)

	if !ok {
		LogErrorf("[Redis] [HINCRBY] invalid type assertion, result %v is %v", result, reflect.TypeOf(result))
		return 0, false
	}

	return count, true
}

// list cmd

/**
 * LPUSH
 * @param key string
 * @param data interface{}
 * @return bool
 */
func (r *Redis) LPush(key string, data interface{}) bool {
	rc, err := getRedisConn()

	if err != nil {
		return false
	}

	defer redisPool.Put(rc)

	conn := rc.(ResourceConn).Conn

	cacheData, jsonErr := json.Marshal(data)

	if jsonErr != nil {
		LogError("[Redis] [LPUSH] ", jsonErr.Error())
		return false
	}

	cacheKey := r.getKey(key)

	_, doErr := conn.Do("LPUSH", cacheKey, cacheData)

	if doErr != nil {
		LogError("[Redis] [LPUSH] ", doErr.Error())
		return false
	}

	return true
}

/**
 * LPOP
 * @param key string
 * @param data interface{} (指针)
 * @return bool
 */
func (r *Redis) LPop(key string, data interface{}) bool {
	rc, err := getRedisConn()

	if err != nil {
		return false
	}

	defer redisPool.Put(rc)

	conn := rc.(ResourceConn).Conn

	cacheKey := r.getKey(key)

	cacheData, doErr := conn.Do("LPOP", cacheKey)

	if doErr != nil {
		LogError("[Redis] [LPOP] ", doErr.Error())
		return false
	}

	if cacheData == nil {
		return false
	}

	jsonErr := json.Unmarshal(cacheData.([]byte), data)

	if jsonErr != nil {
		LogError("[Redis] [LPOP] ", jsonErr.Error())
		return false
	}

	return true
}

/**
 * RPUSH
 * @param key string
 * @param data interface{}
 * @return bool
 */
func (r *Redis) RPush(key string, data interface{}) bool {
	rc, err := getRedisConn()

	if err != nil {
		return false
	}

	defer redisPool.Put(rc)

	conn := rc.(ResourceConn).Conn

	cacheData, jsonErr := json.Marshal(data)

	if jsonErr != nil {
		LogError("[Redis] [RPUSH] ", jsonErr.Error())
		return false
	}

	cacheKey := r.getKey(key)

	_, doErr := conn.Do("RPUSH", cacheKey, cacheData)

	if doErr != nil {
		LogError("[Redis] [RPUSH] ", doErr.Error())
		return false
	}

	return true
}

/**
 * RPOP
 * @param key string
 * @param data interface{} (指针)
 * @return bool
 */
func (r *Redis) RPop(key string, data interface{}) bool {
	rc, err := getRedisConn()

	if err != nil {
		return false
	}

	defer redisPool.Put(rc)

	conn := rc.(ResourceConn).Conn

	cacheKey := r.getKey(key)

	cacheData, doErr := conn.Do("RPOP", cacheKey)

	if doErr != nil {
		LogError("[Redis] [RPOP] ", doErr.Error())
		return false
	}

	if cacheData == nil {
		return false
	}

	jsonErr := json.Unmarshal(cacheData.([]byte), data)

	if jsonErr != nil {
		LogError("[Redis] [RPOP] ", jsonErr.Error())
		return false
	}

	return true
}

/**
 * LLEN
 * @param key string
 * return int64, bool
 */
func (r *Redis) LLen(key string) (int64, bool) {
	rc, err := getRedisConn()

	if err != nil {
		return 0, false
	}

	defer redisPool.Put(rc)

	conn := rc.(ResourceConn).Conn

	cacheKey := r.getKey(key)

	result, doErr := conn.Do("LLEN", cacheKey)

	if doErr != nil {
		LogError("[Redis] [LLEN] ", doErr.Error())
		return 0, false
	}

	count, ok := result.(int64)

	if !ok {
		LogErrorf("[Redis] [LLEN] invalid type assertion, result %v is %v", result, reflect.TypeOf(result))
		return 0, false
	}

	return count, true
}

/**
 * LRANGE
 * @param key string
 * @param start int
 * @param end int
 * @param data interface{} (切片指针)
 * @return bool
 */
func (r *Redis) LRange(key string, start int, end int, data interface{}) bool {
	rc, err := getRedisConn()

	if err != nil {
		return false
	}

	defer redisPool.Put(rc)

	conn := rc.(ResourceConn)

	cacheKey := r.getKey(key)

	cacheData, doErr := redis.ByteSlices(conn.Do("LRANGE", cacheKey, start, end))

	if doErr != nil {
		LogError("[Redis] [LRANGE] ", doErr.Error())
		return false
	}

	if cacheData == nil {
		return false
	}

	if len(cacheData) > 0 {
		refVal := reflect.Indirect(reflect.ValueOf(data))

		if refVal.Kind() == reflect.Slice {
			refValType := refVal.Type().Elem()
			refVal.Set(reflect.MakeSlice(refVal.Type(), 0, 0))

			for _, v := range cacheData {
				if v != nil {
					elem := reflect.New(refValType).Elem()
					jsonErr := json.Unmarshal(v, elem.Addr().Interface())

					if jsonErr != nil {
						LogError("[Redis] [LRANGE] ", jsonErr.Error())
						return false
					}

					refVal.Set(reflect.Append(refVal, elem))
				}
			}
		}
	}

	return true
}

// key cmd

/**
 * HMSET
 * @param key string
 * @return bool
 */
func (r *Redis) Del(key string) bool {
	rc, err := getRedisConn()

	if err != nil {
		return false
	}

	defer redisPool.Put(rc)

	conn := rc.(ResourceConn).Conn

	cacheKey := r.getKey(key)

	_, doErr := conn.Do("DEL", cacheKey)

	if doErr != nil {
		LogError("[Redis] [DEL] ", doErr.Error())
		return false
	}

	return true
}

/**
 * EXPIRE
 * @param key string
 * @param time int
 * @return bool
 */
func (r *Redis) Expire(key string, time int) bool {
	rc, err := getRedisConn()

	if err != nil {
		return false
	}

	defer redisPool.Put(rc)

	conn := rc.(ResourceConn).Conn

	cacheKey := r.getKey(key)

	_, doErr := conn.Do("EXPIRE", cacheKey, time)

	if doErr != nil {
		LogError("[Redis] [EXPIRE] ", doErr.Error())
		return false
	}

	return true
}

/**
 * INCR
 * @param key string
 * @return int64, bool
 */
func (r *Redis) Incr(key string) (int64, bool) {
	rc, err := getRedisConn()

	if err != nil {
		return 0, false
	}

	defer redisPool.Put(rc)

	conn := rc.(ResourceConn).Conn

	cacheKey := r.getKey(key)

	result, doErr := conn.Do("INCR", cacheKey)

	if doErr != nil {
		LogError("[Redis] [INCR] ", doErr.Error())
		return 0, false
	}

	count, ok := result.(int64)

	if !ok {
		LogErrorf("[Redis] [INCR] invalid type assertion, result %v is %v", result, reflect.TypeOf(result))
		return 0, false
	}

	return count, true
}

/**
 * INCRBY
 * @param key string
 * @param inc int
 * @return int64, bool
 */
func (r *Redis) IncrBy(key string, inc int) (int64, bool) {
	rc, err := getRedisConn()

	if err != nil {
		return 0, false
	}

	defer redisPool.Put(rc)

	conn := rc.(ResourceConn).Conn

	cacheKey := r.getKey(key)

	result, doErr := conn.Do("INCRBY", cacheKey, inc)

	if doErr != nil {
		LogError("[Redis] [INCRBY] ", doErr.Error())
		return 0, false
	}

	count, ok := result.(int64)

	if !ok {
		LogErrorf("[Redis] [INCRBY] invalid type assertion, result %v is %v", result, reflect.TypeOf(result))
		return 0, false
	}

	return count, true
}

/**
 * DECR
 * @param key string
 * @return int64, bool
 */
func (r *Redis) Decr(key string) (int64, bool) {
	rc, err := getRedisConn()

	if err != nil {
		return 0, false
	}

	defer redisPool.Put(rc)

	conn := rc.(ResourceConn).Conn

	cacheKey := r.getKey(key)

	result, doErr := conn.Do("DECR", cacheKey)

	if doErr != nil {
		LogError("[Redis] [DECR] ", doErr.Error())
		return 0, false
	}

	count, ok := result.(int64)

	if !ok {
		LogErrorf("[Redis] [DECR] invalid type assertion, result %v is %v", result, reflect.TypeOf(result))
		return 0, false
	}

	return count, true
}

/**
 * DECRBY
 * @param key string
 * @param inc int
 * @return int64, bool
 */
func (r *Redis) DecrBy(key string, inc int) (int64, bool) {
	rc, err := getRedisConn()

	if err != nil {
		return 0, false
	}

	defer redisPool.Put(rc)

	conn := rc.(ResourceConn).Conn

	cacheKey := r.getKey(key)

	result, doErr := conn.Do("DECRBY", cacheKey, inc)

	if doErr != nil {
		LogError("[Redis] [DECRBY] ", doErr.Error())
		return 0, false
	}

	count, ok := result.(int64)

	if !ok {
		LogErrorf("[Redis] [DECRBY] invalid type assertion, result %v is %v", result, reflect.TypeOf(result))
		return 0, false
	}

	return count, true
}
