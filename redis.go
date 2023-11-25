package yiigo

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
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

// CloseRedis 关闭Redis连接，如果未指定名称，则关闭全部
func CloseRedis(name ...string) {
	if len(name) == 0 {
		for key, cli := range redisMap {
			if err := cli.Close(); err != nil {
				logger.Error(fmt.Sprintf("redis.%s close error", key), zap.Error(err))
			}
		}

		return
	}

	for _, key := range name {
		if cli, ok := redisMap[key]; ok {
			if err := cli.Close(); err != nil {
				logger.Error(fmt.Sprintf("redis.%s close error", key), zap.Error(err))
			}
		}
	}
}
