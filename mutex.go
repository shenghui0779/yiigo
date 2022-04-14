package yiigo

import (
	"context"
	"runtime/debug"
	"time"

	"github.com/gomodule/redigo/redis"
	"go.uber.org/zap"
)

// MutexHandler the function to execute after lock acquired.
type MutexHandler func(ctx context.Context) error

// Mutex is a reader/writer mutual exclusion lock.
type Mutex interface {
	// Acquire attempt to acquire lock at regular intervals.
	Acquire(ctx context.Context, callback MutexHandler, interval, timeout time.Duration) error
}

type distributed struct {
	pool   RedisPool
	key    string
	expire int64
}

func (d *distributed) Acquire(ctx context.Context, callback MutexHandler, interval, timeout time.Duration) error {
	mutexCtx := ctx

	if timeout > 0 {
		var cancel context.CancelFunc

		mutexCtx, cancel = context.WithTimeout(mutexCtx, timeout)
		defer cancel()
	}

	conn, err := d.pool.Get(mutexCtx)

	if err != nil {
		return err
	}

	defer d.pool.Put(conn)

	ok, err := d.attempt(conn)

	if err != nil {
		return err
	}

	// if not ok, attempt regularly
	if !ok {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-mutexCtx.Done():
				// timeout or canceled
				return mutexCtx.Err()
			case <-ticker.C:
				ok, err = d.attempt(conn)

				if err != nil {
					return err
				}
			}

			if ok {
				break
			}
		}
	}

	// release lock
	defer func() {
		defer conn.Do("DEL", d.key)

		if err := recover(); err != nil {
			logger.Error("mutex callback panic", zap.Any("error", err), zap.ByteString("stack", debug.Stack()))
		}
	}()

	return callback(ctx)
}

func (d *distributed) attempt(conn *RedisConn) (bool, error) {
	// attempt to acquire lock with `setnx`
	reply, err := redis.String(conn.Do("SET", d.key, time.Now().Nanosecond(), "EX", d.expire, "NX"))

	if err != nil && err != redis.ErrNil {
		return false, err
	}

	if reply == OK {
		return true, nil
	}

	return false, nil
}

// MutexOption mutex option
type MutexOption func(d *distributed)

// WithMutexRedis specifies redis pool for mutex.
func WithMutexRedis(name string) MutexOption {
	return func(d *distributed) {
		d.pool = Redis(name)
	}
}

// WithMutexExpire specifies expire seconds for mutex.
func WithMutexExpire(e time.Duration) MutexOption {
	return func(d *distributed) {
		if sec := int64(e.Seconds()); sec > 0 {
			d.expire = sec
		}
	}
}

// DistributedMutex returns a simple distributed mutual exclusion lock.
func DistributedMutex(key string, options ...MutexOption) Mutex {
	mutex := &distributed{
		pool:   defaultRedis,
		key:    key,
		expire: 10,
	}

	for _, f := range options {
		f(mutex)
	}

	return mutex
}
