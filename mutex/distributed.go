package redis

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

// Mutex 基于Redis实现的分布式锁
type Mutex interface {
	// Lock 尝试获取锁；interval - 每隔指定时间尝试获取一次锁；timeout - 获取锁的超时时间
	Lock(ctx context.Context, interval, timeout time.Duration) error
	// UnLock 释放锁
	UnLock(ctx context.Context) error
}

type distributed struct {
	cli    *redis.Client
	key    string
	uniqID string
	expire time.Duration
}

func (d *distributed) Lock(ctx context.Context, interval, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	for {
		select {
		case <-ctx.Done(): // timeout or canceled
			return ctx.Err()
		default:
		}

		// attempt to acquire lock with `setnx`
		ok, err := d.cli.SetNX(ctx, d.key, d.uniqID, d.expire).Result()
		if err != nil {
			return err
		}
		if ok {
			return nil
		}

		time.Sleep(interval)
	}
}

func (d *distributed) UnLock(ctx context.Context) error {
	v, err := d.cli.Get(ctx, d.key).Result()
	if err != nil {
		return err
	}
	if v != d.uniqID {
		return nil
	}

	return d.cli.Del(ctx, d.key).Err()
}

// DistributedMutex 返回一个分布式锁实例
// uniqueID - 建议使用RequestID
func DistributedMutex(cli *redis.Client, key, uniqueID string, expire time.Duration) Mutex {
	mutex := &distributed{
		cli:    cli,
		key:    key,
		uniqID: uniqueID,
		expire: expire,
	}

	if mutex.expire == 0 {
		mutex.expire = time.Second * 10
	}

	return mutex
}
