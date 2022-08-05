package yiigo

import (
	"context"
	"time"

	"github.com/gomodule/redigo/redis"
)

// Mutex is a reader/writer mutual exclusion lock.
type Mutex interface {
	// Lock attempts to acquire lock at regular intervals.
	Lock(ctx context.Context, interval, timeout time.Duration) error

	// UnLock releases the lock.
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

// MutexOption mutex option
type MutexOption func(d *distributed)

// WithMutexRedis specifies redis pool for mutex.
func WithMutexRedis(name string) MutexOption {
	return func(d *distributed) {
		d.pool = Redis(name)
	}
}

// WithMutexExpire specifies expire time (ms) for mutex.
func WithMutexExpire(e time.Duration) MutexOption {
	return func(d *distributed) {
		d.expire = e
	}
}

// DistributedMutex returns a simple distributed mutual exclusion lock.
// uniqueID: suggest to use the request id.
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
