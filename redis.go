package yiigo

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/garyburd/redigo/redis"
	"github.com/youtube/vitess/go/pools"
	"golang.org/x/net/context"
)

type RedisBase struct {
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
		LogError("redis connection close error: ", err.Error())
	}
}

/**
 * 连接Redis
 * @return redis.Conn, error
 */
func redisDial() (redis.Conn, error) {
	host := GetEnvString("redis", "host", "localhost")
	port := GetEnvInt("redis", "port", 6379)
	connectTimeout := GetEnvInt("redis", "connectTimeout", 10000)
	readTimeout := GetEnvInt("redis", "readTimeout", 10000)
	writeTimeout := GetEnvInt("redis", "writeTimeout", 10000)

	dsn := fmt.Sprintf("%s:%d", host, port)
	conn, err := redis.DialTimeout("tcp", dsn, time.Duration(connectTimeout)*time.Millisecond, time.Duration(readTimeout)*time.Millisecond, time.Duration(writeTimeout)*time.Millisecond)

	if err != nil {
		LogError("init redis error: ", err.Error())
		return nil, err
	}

	return conn, err
}

/**
 * 初始化Redis连接池
 */
func InitRedis() {
	redisMux.Lock()
	defer redisMux.Unlock()

	if redisPool == nil {
		poolMinActive := GetEnvInt("redis", "poolMinActive", 10)
		poolMaxActive := GetEnvInt("redis", "poolMaxActive", 20)
		poolIdleTimeout := GetEnvInt("redis", "poolIdleTimeout", 60000)

		redisPool = pools.NewResourcePool(func() (pools.Resource, error) {
			conn, err := redisDial()
			return ResourceConn{conn}, err
		}, poolMinActive, poolMaxActive, time.Duration(poolIdleTimeout)*time.Millisecond)
	}
}

/**
 * 获取Redis资源
 * @return pools.Resource, error
 */
func getRedisConn() (pools.Resource, error) {
	if redisPool.IsClosed() {
		InitRedis()
	}

	ctx := context.TODO()
	rc, err := redisPool.Get(ctx)

	if err != nil {
		LogError("get redis resourceConn error: ", err.Error())
	}

	return rc, err
}

func (r *RedisBase) getKey(key string) string {
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
func (r *RedisBase) Set(key string, data interface{}) bool {
	rc, err := getRedisConn()

	if err != nil {
		return false
	}

	defer redisPool.Put(rc)

	conn := rc.(ResourceConn).Conn

	cacheData, jsonErr := json.Marshal(data)

	if jsonErr != nil {
		LogError("redis SET marshal data error: ", jsonErr.Error())
		return false
	}

	cacheKey := r.getKey(key)

	_, doErr := conn.Do("SET", cacheKey, cacheData)

	if doErr != nil {
		LogError("redis do SET error: ", doErr.Error())
		return false
	}

	return true
}

/**
 * MSET
 * @param data map[string]interface{}
 * @return bool
 */
func (r *RedisBase) MSet(data map[string]interface{}) bool {
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
			LogError("redis SET marshal data error: ", jsonErr.Error())
			return false
		}

		args = append(args, r.getKey(k), cacheData)
	}

	_, doErr := conn.Do("MSET", args...)

	if doErr != nil {
		LogError("redis do MSET error: ", doErr.Error())
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
func (r *RedisBase) Get(key string, data interface{}) bool {
	rc, err := getRedisConn()

	if err != nil {
		return false
	}

	defer redisPool.Put(rc)

	conn := rc.(ResourceConn).Conn

	cacheKey := r.getKey(key)

	cacheData, doErr := conn.Do("GET", cacheKey)

	if doErr != nil {
		LogError("redis do GET error: ", doErr.Error())
		return false
	}

	if cacheData == nil {
		return false
	}

	jsonErr := json.Unmarshal(cacheData.([]byte), data)

	if jsonErr != nil {
		LogError("redis GET unmarshal cacheData error: ", jsonErr.Error())
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
func (r *RedisBase) MGet(keys []string, data interface{}) bool {
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
		LogError("redis do MGET error: ", doErr.Error())
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
						LogError("redis MGET unmarshal cacheData error: ", jsonErr.Error())
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
func (r *RedisBase) HSet(key string, field interface{}, data interface{}) bool {
	rc, err := getRedisConn()

	if err != nil {
		return false
	}

	defer redisPool.Put(rc)

	conn := rc.(ResourceConn).Conn

	cacheData, jsonErr := json.Marshal(data)

	if jsonErr != nil {
		LogError("redis HSET marshal data error: ", jsonErr.Error())
		return false
	}

	cacheKey := r.getKey(key)

	_, doErr := conn.Do("HSET", cacheKey, field, cacheData)

	if doErr != nil {
		LogError("redis do HSET error: ", doErr.Error())
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
func (r *RedisBase) HMSet(key string, data map[interface{}]interface{}) bool {
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
			LogError("redis HMSet marshal data error: ", jsonErr.Error())
			return false
		}

		args = append(args, field, cacheData)
	}

	_, doErr := conn.Do("HMSet", args...)

	if doErr != nil {
		LogError("redis do HMSet error: ", doErr.Error())
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
func (r *RedisBase) HGet(key string, field interface{}, data interface{}) bool {
	rc, err := getRedisConn()

	if err != nil {
		return false
	}

	defer redisPool.Put(rc)

	conn := rc.(ResourceConn).Conn

	cacheKey := r.getKey(key)

	cacheData, doErr := conn.Do("HGET", cacheKey, field)

	if doErr != nil {
		LogError("redis do HGET error: ", doErr.Error())
		return false
	}

	if cacheData == nil {
		return false
	}

	jsonErr := json.Unmarshal(cacheData.([]byte), data)

	if jsonErr != nil {
		LogError("redis HGET unmarshal cacheData error: ", jsonErr.Error())
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
func (r *RedisBase) HMGet(key string, fields []interface{}, data interface{}) bool {
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
		LogError("redis do HMGET error: ", doErr.Error())
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
						LogError("redis HMGET unmarshal cacheData error: ", jsonErr.Error())
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
func (r *RedisBase) HDel(key string, field interface{}) bool {
	rc, err := getRedisConn()

	if err != nil {
		return false
	}

	defer redisPool.Put(rc)

	conn := rc.(ResourceConn).Conn

	cacheKey := r.getKey(key)

	_, doErr := conn.Do("HDEL", cacheKey, field)

	if doErr != nil {
		LogError("redis do HDEL error: ", doErr.Error())
		return false
	}

	return true
}

/**
 * HLEN
 * @param key string
 * @return int64, bool
 */
func (r *RedisBase) HLen(key string) (int64, bool) {
	rc, err := getRedisConn()

	if err != nil {
		return 0, false
	}

	defer redisPool.Put(rc)

	conn := rc.(ResourceConn).Conn

	cacheKey := r.getKey(key)

	result, doErr := conn.Do("HLEN", cacheKey)

	if doErr != nil {
		LogError("redis do HLEN error: ", doErr.Error())
		return 0, false
	}

	count, ok := result.(int64)

	if !ok {
		LogErrorf("redis do HLEN type assertion error: result %v is %v", result, reflect.TypeOf(result))
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
func (r *RedisBase) HIncrBy(key string, field interface{}, inc int) (int64, bool) {
	rc, err := getRedisConn()

	if err != nil {
		return 0, false
	}

	defer redisPool.Put(rc)

	conn := rc.(ResourceConn).Conn

	cacheKey := r.getKey(key)

	result, doErr := conn.Do("HINCRBY", cacheKey, field, inc)

	if doErr != nil {
		LogError("redis do HINCRBY error: ", doErr.Error())
		return 0, false
	}

	count, ok := result.(int64)

	if !ok {
		LogErrorf("redis do HINCRBY type assertion error: result %v is %v", result, reflect.TypeOf(result))
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
func (r *RedisBase) LPush(key string, data interface{}) bool {
	rc, err := getRedisConn()

	if err != nil {
		return false
	}

	defer redisPool.Put(rc)

	conn := rc.(ResourceConn).Conn

	cacheData, jsonErr := json.Marshal(data)

	if jsonErr != nil {
		LogError("redis LPUSH marshal data error: ", jsonErr.Error())
		return false
	}

	cacheKey := r.getKey(key)

	_, doErr := conn.Do("LPUSH", cacheKey, cacheData)

	if doErr != nil {
		LogError("redis do LPUSH error: ", doErr.Error())
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
func (r *RedisBase) LPop(key string, data interface{}) bool {
	rc, err := getRedisConn()

	if err != nil {
		return false
	}

	defer redisPool.Put(rc)

	conn := rc.(ResourceConn).Conn

	cacheKey := r.getKey(key)

	cacheData, doErr := conn.Do("LPOP", cacheKey)

	if doErr != nil {
		LogError("redis do LPOP error: ", doErr.Error())
		return false
	}

	if cacheData == nil {
		return false
	}

	jsonErr := json.Unmarshal(cacheData.([]byte), data)

	if jsonErr != nil {
		LogError("redis LPOP unmarshal cacheData error: ", jsonErr.Error())
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
func (r *RedisBase) RPush(key string, data interface{}) bool {
	rc, err := getRedisConn()

	if err != nil {
		return false
	}

	defer redisPool.Put(rc)

	conn := rc.(ResourceConn).Conn

	cacheData, jsonErr := json.Marshal(data)

	if jsonErr != nil {
		LogError("redis RPUSH marshal data error: ", jsonErr.Error())
		return false
	}

	cacheKey := r.getKey(key)

	_, doErr := conn.Do("RPUSH", cacheKey, cacheData)

	if doErr != nil {
		LogError("redis do RPUSH error: ", doErr.Error())
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
func (r *RedisBase) RPop(key string, data interface{}) bool {
	rc, err := getRedisConn()

	if err != nil {
		return false
	}

	defer redisPool.Put(rc)

	conn := rc.(ResourceConn).Conn

	cacheKey := r.getKey(key)

	cacheData, doErr := conn.Do("RPOP", cacheKey)

	if doErr != nil {
		LogError("redis do RPOP error: ", doErr.Error())
		return false
	}

	if cacheData == nil {
		return false
	}

	jsonErr := json.Unmarshal(cacheData.([]byte), data)

	if jsonErr != nil {
		LogError("redis RPOP unmarshal cacheData error: ", jsonErr.Error())
		return false
	}

	return true
}

/**
 * LLEN
 * @param key string
 * return int64, bool
 */
func (r *RedisBase) LLen(key string) (int64, bool) {
	rc, err := getRedisConn()

	if err != nil {
		return 0, false
	}

	defer redisPool.Put(rc)

	conn := rc.(ResourceConn).Conn

	cacheKey := r.getKey(key)

	result, doErr := conn.Do("LLEN", cacheKey)

	if doErr != nil {
		LogError("redis do LLEN error: ", doErr.Error())
		return 0, false
	}

	count, ok := result.(int64)

	if !ok {
		LogErrorf("redis do LLEN type assertion error: result %v is %v", result, reflect.TypeOf(result))
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
func (r *RedisBase) LRange(key string, start int, end int, data interface{}) bool {
	rc, err := getRedisConn()

	if err != nil {
		return false
	}

	defer redisPool.Put(rc)

	conn := rc.(ResourceConn)

	cacheKey := r.getKey(key)

	cacheData, doErr := redis.ByteSlices(conn.Do("LRANGE", cacheKey, start, end))

	if doErr != nil {
		LogError("redis do LRANGE error: ", doErr.Error())
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
						LogError("redis LRANGE unmarshal cacheData error: ", jsonErr.Error())
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
func (r *RedisBase) Del(key string) bool {
	rc, err := getRedisConn()

	if err != nil {
		return false
	}

	defer redisPool.Put(rc)

	conn := rc.(ResourceConn).Conn

	cacheKey := r.getKey(key)

	_, doErr := conn.Do("DEL", cacheKey)

	if doErr != nil {
		LogError("redis do DEL error: ", doErr.Error())
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
func (r *RedisBase) Expire(key string, time int) bool {
	rc, err := getRedisConn()

	if err != nil {
		return false
	}

	defer redisPool.Put(rc)

	conn := rc.(ResourceConn).Conn

	cacheKey := r.getKey(key)

	_, doErr := conn.Do("EXPIRE", cacheKey, time)

	if doErr != nil {
		LogError("redis do EXPIRE error: ", doErr.Error())
		return false
	}

	return true
}

/**
 * INCR
 * @param key string
 * @return int64, bool
 */
func (r *RedisBase) Incr(key string) (int64, bool) {
	rc, err := getRedisConn()

	if err != nil {
		return 0, false
	}

	defer redisPool.Put(rc)

	conn := rc.(ResourceConn).Conn

	cacheKey := r.getKey(key)

	result, doErr := conn.Do("INCR", cacheKey)

	if doErr != nil {
		LogError("redis do INCR error: ", doErr.Error())
		return 0, false
	}

	count, ok := result.(int64)

	if !ok {
		LogErrorf("redis do INCR type assertion error: result %v is %v", result, reflect.TypeOf(result))
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
func (r *RedisBase) IncrBy(key string, inc int) (int64, bool) {
	rc, err := getRedisConn()

	if err != nil {
		return 0, false
	}

	defer redisPool.Put(rc)

	conn := rc.(ResourceConn).Conn

	cacheKey := r.getKey(key)

	result, doErr := conn.Do("INCRBY", cacheKey, inc)

	if doErr != nil {
		LogError("redis do INCRBY error: ", doErr.Error())
		return 0, false
	}

	count, ok := result.(int64)

	if !ok {
		LogErrorf("redis do INCRBY type assertion error: result %v is %v", result, reflect.TypeOf(result))
		return 0, false
	}

	return count, true
}

/**
 * DECR
 * @param key string
 * @return int64, bool
 */
func (r *RedisBase) Decr(key string) (int64, bool) {
	rc, err := getRedisConn()

	if err != nil {
		return 0, false
	}

	defer redisPool.Put(rc)

	conn := rc.(ResourceConn).Conn

	cacheKey := r.getKey(key)

	result, doErr := conn.Do("DECR", cacheKey)

	if doErr != nil {
		LogError("redis do DECR error: ", doErr.Error())
		return 0, false
	}

	count, ok := result.(int64)

	if !ok {
		LogErrorf("redis do DECR type assertion error: result %v is %v", result, reflect.TypeOf(result))
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
func (r *RedisBase) DecrBy(key string, inc int) (int64, bool) {
	rc, err := getRedisConn()

	if err != nil {
		return 0, false
	}

	defer redisPool.Put(rc)

	conn := rc.(ResourceConn).Conn

	cacheKey := r.getKey(key)

	result, doErr := conn.Do("DECRBY", cacheKey, inc)

	if doErr != nil {
		LogError("redis do DECRBY error: ", doErr.Error())
		return 0, false
	}

	count, ok := result.(int64)

	if !ok {
		LogErrorf("redis do DECRBY type assertion error: result %v is %v", result, reflect.TypeOf(result))
		return 0, false
	}

	return count, true
}
