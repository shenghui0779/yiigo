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

type RedisBase struct {
	CacheName string
}

var (
	redisPool    *pools.ResourcePool
	redisPoolMux sync.Mutex
)

type ResourceConn struct {
	redis.Conn
}

func (r ResourceConn) Close() {
	err := r.Conn.Close()

	if err != nil {
		LogError("redis connection close error: ", err.Error())
	}
}

func initRedis() (redis.Conn, error) {
	host := GetConfigString("redis", "host", "localhost")
	port := GetConfigInt("redis", "port", 6379)
	connectTimeout := GetConfigInt("redis", "connectTimeout", 0)
	readTimeout := GetConfigInt("redis", "readTimeout", 10000)
	writeTimeout := GetConfigInt("redis", "writeTimeout", 10000)

	address := fmt.Sprintf("%s:%d", host, port)
	conn, err := redis.DialTimeout("tcp", address, time.Duration(connectTimeout)*time.Millisecond, time.Duration(readTimeout)*time.Millisecond, time.Duration(writeTimeout)*time.Millisecond)

	if err != nil {
		LogError("init redis error: ", err.Error())
		return nil, err
	}

	return conn, err
}

func initRedisPool() {
	if redisPool == nil || redisPool.IsClosed() {
		redisPoolMux.Lock()
		defer redisPoolMux.Unlock()

		if redisPool == nil {
			poolMinActive := GetConfigInt("redis", "poolMinActive", 100)
			poolMaxActive := GetConfigInt("redis", "poolMaxActive", 200)
			poolIdleTimeout := GetConfigInt("redis", "poolIdleTimeout", 2000)

			redisPool = pools.NewResourcePool(func() (pools.Resource, error) {
				conn, err := initRedis()
				return ResourceConn{conn}, err
			}, poolMinActive, poolMaxActive, time.Duration(poolIdleTimeout)*time.Millisecond)
		}
	}
}

func poolGetRedisConn() (pools.Resource, error) {
	initRedisPool()

	if redisPool == nil {
		LogError("redis pool is null")
		return nil, errors.New("redis pool is null")
	}

	ctx := context.TODO()
	redisResource, err := redisPool.Get(ctx)

	if err != nil {
		LogError("redis get connection err: ", err.Error())
		return nil, err
	}

	if redisResource == nil {
		LogError("redis pool resource is null")
		return nil, errors.New("redis pool resource is null")
	}

	resourceConn := redisResource.(ResourceConn)

	if resourceConn.Conn.Err() != nil {
		LogError("redis resource connection err: ", resourceConn.Conn.Err().Error())

		resourceConn.Close()
		//连接断开，重新打开
		conn, connErr := initRedis()

		if connErr != nil {
			LogError("redis reconnection err: ", connErr.Error())
			return nil, connErr
		} else {
			return ResourceConn{conn}, nil
		}
	}

	return redisResource, nil
}

func (r *RedisBase) getKey(key string) string {
	prefix := GetConfigString("redis", "prefix", "yii")

	if strings.Trim(key, " ") == "" {
		return fmt.Sprintf("%s:%s", prefix, r.CacheName)
	}

	return fmt.Sprintf("%s:%s:%s", prefix, r.CacheName, key)
}

// string cmd

func (r *RedisBase) Set(key string, data interface{}) bool {
	redisResource, err := poolGetRedisConn()
	defer redisPool.Put(redisResource)

	if err != nil {
		return false
	}

	redisConn := redisResource.(ResourceConn).Conn

	cacheData, jsonErr := json.Marshal(data)

	if jsonErr != nil {
		LogError("redis SET marshal data error: ", jsonErr.Error())
		return false
	}

	cacheKey := r.getKey(key)

	_, doErr := redisConn.Do("SET", cacheKey, cacheData)

	if doErr != nil {
		LogError("redis do SET error: ", doErr.Error())
		return false
	}

	return true
}

func (r *RedisBase) MSet(data map[string]interface{}) bool {
	redisResource, err := poolGetRedisConn()
	defer redisPool.Put(redisResource)

	if err != nil {
		return false
	}

	redisConn := redisResource.(ResourceConn).Conn

	args := []interface{}{}

	for k, v := range data {
		cacheData, jsonErr := json.Marshal(v)

		if jsonErr != nil {
			LogError("redis SET marshal data error: ", jsonErr.Error())
			return false
		}

		args = append(args, r.getKey(k), cacheData)
	}

	_, doErr := redisConn.Do("MSET", args...)

	if doErr != nil {
		LogError("redis do MSET error: ", doErr.Error())
		return false
	}

	return true
}

func (r *RedisBase) Get(data interface{}, key string) bool {
	redisResource, err := poolGetRedisConn()
	defer redisPool.Put(redisResource)

	if err != nil {
		return false
	}

	redisConn := redisResource.(ResourceConn).Conn

	cacheKey := r.getKey(key)

	cacheData, doErr := redisConn.Do("GET", cacheKey)

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

func (r *RedisBase) MGet(data interface{}, keys []string) bool {
	redisResource, err := poolGetRedisConn()
	defer redisPool.Put(redisResource)

	if err != nil {
		return false
	}

	redisConn := redisResource.(ResourceConn).Conn

	args := []interface{}{}

	for _, key := range keys {
		args = append(args, r.getKey(key))
	}

	cacheData, doErr := redis.ByteSlices(redisConn.Do("MGET", args...))

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

func (r *RedisBase) HSet(key string, field interface{}, data interface{}) bool {
	redisResource, err := poolGetRedisConn()
	defer redisPool.Put(redisResource)

	if err != nil {
		return false
	}

	redisConn := redisResource.(ResourceConn).Conn

	cacheData, jsonErr := json.Marshal(data)

	if jsonErr != nil {
		LogError("redis HSET marshal data error: ", jsonErr.Error())
		return false
	}

	cacheKey := r.getKey(key)

	_, doErr := redisConn.Do("HSET", cacheKey, field, cacheData)

	if doErr != nil {
		LogError("redis do HSET error: ", doErr.Error())
		return false
	}

	return true
}

func (r *RedisBase) HMSet(key string, data map[interface{}]interface{}) bool {
	redisResource, err := poolGetRedisConn()
	defer redisPool.Put(redisResource)

	if err != nil {
		return false
	}

	redisConn := redisResource.(ResourceConn).Conn

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

	_, doErr := redisConn.Do("HMSet", args...)

	if doErr != nil {
		LogError("redis do HMSet error: ", doErr.Error())
		return false
	}

	return true
}

func (r *RedisBase) HGet(data interface{}, key string, field interface{}) bool {
	redisResource, err := poolGetRedisConn()
	defer redisPool.Put(redisResource)

	if err != nil {
		return false
	}

	redisConn := redisResource.(ResourceConn).Conn

	cacheKey := r.getKey(key)

	cacheData, doErr := redisConn.Do("HGET", cacheKey, field)

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

func (r *RedisBase) HMGet(data interface{}, key string, fields []interface{}) bool {
	redisResource, err := poolGetRedisConn()
	defer redisPool.Put(redisResource)

	if err != nil {
		return false
	}

	redisConn := redisResource.(ResourceConn).Conn

	args := []interface{}{}
	args = append(args, r.getKey(key))

	for _, field := range fields {
		args = append(args, field)
	}

	cacheData, doErr := redis.ByteSlices(redisConn.Do("HMGET", args...))

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

func (r *RedisBase) HDel(key string, field interface{}) bool {
	redisResource, err := poolGetRedisConn()
	defer redisPool.Put(redisResource)

	if err != nil {
		return false
	}

	redisConn := redisResource.(ResourceConn).Conn

	cacheKey := r.getKey(key)

	_, doErr := redisConn.Do("HDEL", cacheKey, field)

	if doErr != nil {
		LogError("redis do HDEL error: ", doErr.Error())
		return false
	}

	return true
}

func (r *RedisBase) HLen(key string) (int64, bool) {
	redisResource, err := poolGetRedisConn()
	defer redisPool.Put(redisResource)

	if err != nil {
		return 0, false
	}

	redisConn := redisResource.(ResourceConn).Conn

	cacheKey := r.getKey(key)

	result, doErr := redisConn.Do("HLEN", cacheKey)

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

func (r *RedisBase) HIncrBy(key string, field interface{}, inc int) (int64, bool) {
	redisResource, err := poolGetRedisConn()
	defer redisPool.Put(redisResource)

	if err != nil {
		return 0, false
	}

	redisConn := redisResource.(ResourceConn).Conn

	cacheKey := r.getKey(key)

	result, doErr := redisConn.Do("HINCRBY", cacheKey, field, inc)

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
func (r *RedisBase) LPush(key string, data interface{}) bool {
	redisResource, err := poolGetRedisConn()
	defer redisPool.Put(redisResource)

	if err != nil {
		return false
	}

	redisConn := redisResource.(ResourceConn).Conn

	cacheData, jsonErr := json.Marshal(data)

	if jsonErr != nil {
		LogError("redis LPUSH marshal data error: ", jsonErr.Error())
		return false
	}

	cacheKey := r.getKey(key)

	_, doErr := redisConn.Do("LPUSH", cacheKey, cacheData)

	if doErr != nil {
		LogError("redis do LPUSH error: ", doErr.Error())
		return false
	}

	return true
}

func (r *RedisBase) LPop(data interface{}, key string) bool {
	redisResource, err := poolGetRedisConn()
	defer redisPool.Put(redisResource)

	if err != nil {
		return false
	}

	redisConn := redisResource.(ResourceConn).Conn

	cacheKey := r.getKey(key)

	cacheData, doErr := redisConn.Do("LPOP", cacheKey)

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

func (r *RedisBase) RPush(key string, data interface{}) bool {
	redisResource, err := poolGetRedisConn()
	defer redisPool.Put(redisResource)

	if err != nil {
		return false
	}

	redisConn := redisResource.(ResourceConn).Conn

	cacheData, jsonErr := json.Marshal(data)

	if jsonErr != nil {
		LogError("redis RPUSH marshal data error: ", jsonErr.Error())
		return false
	}

	cacheKey := r.getKey(key)

	_, doErr := redisConn.Do("RPUSH", cacheKey, cacheData)

	if doErr != nil {
		LogError("redis do RPUSH error: ", doErr.Error())
		return false
	}

	return true
}

func (r *RedisBase) RPop(data interface{}, key string) bool {
	redisResource, err := poolGetRedisConn()
	defer redisPool.Put(redisResource)

	if err != nil {
		return false
	}

	redisConn := redisResource.(ResourceConn).Conn

	cacheKey := r.getKey(key)

	cacheData, doErr := redisConn.Do("RPOP", cacheKey)

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

func (r *RedisBase) LLen(key string) (int64, bool) {
	redisResource, err := poolGetRedisConn()
	defer redisPool.Put(redisResource)

	if err != nil {
		return 0, false
	}

	redisConn := redisResource.(ResourceConn).Conn

	cacheKey := r.getKey(key)

	result, doErr := redisConn.Do("LLEN", cacheKey)

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

func (b *DaoRedis) LRange(data interface{}, key string, start int, end int) bool {
	redisResource, err := b.InitRedisPool()
	defer redisPool.Put(redisResource)

	if err != nil {
		return false
	}

	redisConn := redisResource.(ResourceConn)

	cacheKey := r.getKey(key)

	cacheData, doErr := redis.ByteSlices(redisConn.Do("LRANGE", cacheKey, start, end))

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

func (r *RedisBase) Del(key string) bool {
	redisResource, err := poolGetRedisConn()
	defer redisPool.Put(redisResource)

	if err != nil {
		return false
	}

	redisConn := redisResource.(ResourceConn).Conn

	cacheKey := r.getKey(key)

	_, doErr := redisConn.Do("DEL", cacheKey)

	if doErr != nil {
		LogError("redis do DEL error: ", doErr.Error())
		return false
	}

	return true
}

func (r *RedisBase) Expire(key string, time int) bool {
	redisResource, err := poolGetRedisConn()
	defer redisPool.Put(redisResource)

	if err != nil {
		return false
	}

	redisConn := redisResource.(ResourceConn).Conn

	cacheKey := r.getKey(key)

	_, doErr := redisConn.Do("EXPIRE", cacheKey, time)

	if doErr != nil {
		LogError("redis do EXPIRE error: ", doErr.Error())
		return false
	}

	return true
}

func (r *RedisBase) Incr(key string) (int64, bool) {
	redisResource, err := poolGetRedisConn()
	defer redisPool.Put(redisResource)

	if err != nil {
		return 0, false
	}

	redisConn := redisResource.(ResourceConn).Conn

	cacheKey := r.getKey(key)

	result, doErr := redisConn.Do("INCR", cacheKey)

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

func (r *RedisBase) IncrBy(key string, inc int) (int64, bool) {
	redisResource, err := poolGetRedisConn()
	defer redisPool.Put(redisResource)

	if err != nil {
		return 0, false
	}

	redisConn := redisResource.(ResourceConn).Conn

	cacheKey := r.getKey(key)

	result, doErr := redisConn.Do("INCRBY", cacheKey, inc)

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

func (r *RedisBase) Decr(key string) (int64, bool) {
	redisResource, err := poolGetRedisConn()
	defer redisPool.Put(redisResource)

	if err != nil {
		return 0, false
	}

	redisConn := redisResource.(ResourceConn).Conn

	cacheKey := r.getKey(key)

	result, doErr := redisConn.Do("DECR", cacheKey)

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

func (r *RedisBase) DecrBy(key string, inc int) (int64, bool) {
	redisResource, err := poolGetRedisConn()
	defer redisPool.Put(redisResource)

	if err != nil {
		return 0, false
	}

	redisConn := redisResource.(ResourceConn).Conn

	cacheKey := r.getKey(key)

	result, doErr := redisConn.Do("DECRBY", cacheKey, inc)

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
