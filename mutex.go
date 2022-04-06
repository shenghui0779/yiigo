package yiigo

import (
	"context"
	"runtime/debug"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/pkg/errors"
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

	conn, err := defaultRedis.Get(mutexCtx)

	if err != nil {
		return errors.Wrap(err, "redis conn")
	}

	defer defaultRedis.Put(conn)

	for {
		select {
		case <-mutexCtx.Done():
			// timeout or canceled
			return errors.Wrap(mutexCtx.Err(), "mutex context")
		default:
		}

		// attempt to acquire lock with `setnx`
		reply, err := redis.String(conn.Do("SET", d.key, time.Now().Nanosecond(), "EX", d.expire, "NX"))

		if err != nil && err != redis.ErrNil {
			return errors.Wrap(err, "redis setnx")
		}

		if reply == "OK" {
			break
		}

		time.Sleep(interval)
	}

	// release lock
	defer func() {
		defer conn.Do("DEL", d.key)

		if err := recover(); err != nil {
			logger.Error("mutex callback panic",
				zap.Any("error", err),
				zap.ByteString("stack", debug.Stack()),
			)
		}
	}()

	return callback(ctx)
}

// DistributedMutex returns a simple distributed mutual exclusion lock.
func DistributedMutex(key string, expire time.Duration) Mutex {
	mutex := &distributed{
		key:    key,
		expire: 10,
	}

	if seconds := expire.Seconds(); seconds > 0 {
		mutex.expire = int64(seconds)
	}

	return mutex
}
