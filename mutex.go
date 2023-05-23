package yiigo

import (
	"context"
	"time"

	"github.com/gomodule/redigo/redis"
)

// Mutex 基于Redis实现的分布式锁
type Mutex interface {
	// Lock 尝试获取锁
	// interval - 每隔指定时间尝试获取一次锁
	// timeout - 获取锁的超时时间
	Lock(ctx context.Context, interval, timeout time.Duration) error

	// UnLock 释放锁
	UnLock(ctx context.Context) error
}

type distributed struct {
	pool   RedisPool
	key    string
	uniqID string
	expire time.Duration
}

func (d *distributed) Lock(ctx context.Context, interval, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	conn, err := d.pool.Get(ctx)

	if err != nil {
		return err
	}

	defer d.pool.Put(conn)

	for {
		select {
		case <-ctx.Done(): // timeout or canceled
			return ctx.Err()
		default:
		}

		ok, err := d.attempt(conn)

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
	conn, err := d.pool.Get(ctx)

	if err != nil {
		return err
	}

	defer d.pool.Put(conn)

	v, err := redis.String(conn.Do("GET", d.key))

	if err != nil {
		return err
	}

	if v != d.uniqID {
		return nil
	}

	_, err = conn.Do("DEL", d.key)

	return err
}

func (d *distributed) attempt(conn *RedisConn) (bool, error) {
	// attempt to acquire lock with `setnx`
	reply, err := redis.String(conn.Do("SET", d.key, d.uniqID, "PX", d.expire.Milliseconds(), "NX"))

	if err != nil && err != redis.ErrNil {
		return false, err
	}

	if reply == OK {
		return true, nil
	}

	return false, nil
}

// MutexOption 锁选项
type MutexOption func(d *distributed)

// WithMutexRedis 指定Redis实例
func WithMutexRedis(name string) MutexOption {
	return func(d *distributed) {
		d.pool = Redis(name)
	}
}

// WithMutexExpire 设置锁的有效期
func WithMutexExpire(e time.Duration) MutexOption {
	return func(d *distributed) {
		d.expire = e
	}
}

// DistributedMutex 返回一个分布式锁实例
// uniqueID - 建议使用RequestID
func DistributedMutex(key, uniqueID string, options ...MutexOption) Mutex {
	mutex := &distributed{
		pool:   defaultRedis,
		key:    key,
		uniqID: uniqueID,
		expire: 10 * time.Second,
	}

	for _, f := range options {
		f(mutex)
	}

	return mutex
}
