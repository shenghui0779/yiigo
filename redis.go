package yiigo

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

var redisMap = make(map[string]redis.UniversalClient)

func initRedis(name string, cfg *redis.UniversalOptions) error {
	cli := redis.NewUniversalClient(cfg)

	// verify connection
	if err := cli.Ping(context.Background()).Err(); err != nil {
		cli.Close()
		return err
	}

	redisMap[name] = cli

	return nil
}

// Redis 返回一个Redis连接池实例
func Redis(name ...string) (redis.UniversalClient, error) {
	key := Default
	if len(name) != 0 {
		key = name[0]
	}

	cli, ok := redisMap[key]
	if !ok {
		return nil, fmt.Errorf("unknown redis.%s (forgotten configure?)", key)
	}

	return cli, nil
}

// MustRedis 返回一个Redis连接池实例，如果不存在，则Panic
func MustRedis(name ...string) redis.UniversalClient {
	cli, err := Redis(name...)
	if err != nil {
		logger.Panic(err.Error())
	}

	return cli
}
