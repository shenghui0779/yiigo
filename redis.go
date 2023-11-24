package yiigo

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/shenghui0779/vitess_pool"
	"go.uber.org/zap"
)

// RedisConn Redis连接
type RedisConn struct {
	name string
	redis.Conn
}

// Name 返回Redis连接名称
func (rc *RedisConn) Name() string {
	return rc.name
}

// Close 关闭Redis连接
func (rc *RedisConn) Close() {
	if err := rc.Conn.Close(); err != nil {
		logger.Error(fmt.Sprintf("err redis.%s conn close", rc.name), zap.Error(err))
	}
}

// RedisPool Redis连接池
type RedisPool interface {
	// Get 获取一个连接；Context可用于控制获取等待超时时长
	Get(ctx context.Context) (*RedisConn, error)

	// Put 连接放回连接池
	Put(rc *RedisConn)

	// Do 执行命令
	Do(ctx context.Context, cmd string, args ...any) (any, error)

	// DoFunc 通过回调函数执行一组命令
	DoFunc(ctx context.Context, f func(ctx context.Context, conn *RedisConn) error) error
}

// RedisConfig Redis初始化配置
type RedisConfig struct {
	// Addr 连接地址
	Addr string `json:"addr"`

	// Options 配置选项
	Options *RedisOptions `json:"options"`
}

// RedisOptions Redis配置选项
type RedisOptions struct {
	// Dialer 自定义TCP连接创建方法；否则，使用其它配置选项
	Dialer func(ctx context.Context, network, addr string) (net.Conn, error) `json:"dialer"`

	// Username 授权用户名
	Username string `json:"username"`

	// Password 授权密码
	Password string `json:"password"`

	// Database 指定数据库
	Database int `json:"database"`

	// ConnTimeout 连接超时；-1：不限；默认：10秒
	ConnTimeout time.Duration `json:"conn_timeout"`

	// ReadTimeout 读超时；-1：不限；默认：10秒
	ReadTimeout time.Duration `json:"read_timeout"`

	// WriteTimeout 写超时；-1：不限；默认：10秒
	WriteTimeout time.Duration `json:"write_timeout"`

	// PoolSize 连接池大小；默认：10
	PoolSize int `json:"pool_size"`

	// PoolPrefill 连接池预填充连接数；默认：不填充
	PoolPrefill int `json:"pool_prefill"`

	// IdleTimeout 连接最大闲置时间；-1：不限；默认：5分钟
	IdleTimeout time.Duration `json:"idle_timeout"`

	// TLSConfig TLS连接配置
	TLSConfig *tls.Config `json:"tls_config"`
}

func (o *RedisOptions) rebuild(opt *RedisOptions) {
	o.Dialer = opt.Dialer
	o.TLSConfig = opt.TLSConfig

	if len(opt.Username) != 0 {
		o.Username = opt.Username
	}

	if len(opt.Password) != 0 {
		o.Password = opt.Password
	}

	if opt.Database > 0 {
		o.Database = opt.Database
	}

	if opt.ConnTimeout > 0 {
		o.ConnTimeout = opt.ConnTimeout
	} else {
		if opt.ConnTimeout == -1 {
			o.ConnTimeout = 0
		}
	}

	if opt.ReadTimeout > 0 {
		o.ReadTimeout = opt.ReadTimeout
	} else {
		if opt.ReadTimeout == -1 {
			o.ReadTimeout = 0
		}
	}

	if opt.WriteTimeout > 0 {
		o.WriteTimeout = opt.WriteTimeout
	} else {
		if opt.WriteTimeout == -1 {
			o.WriteTimeout = 0
		}
	}

	if opt.PoolSize > 0 {
		o.PoolSize = opt.PoolSize
	}

	if opt.PoolPrefill > 0 {
		o.PoolPrefill = opt.PoolPrefill
	}

	if opt.IdleTimeout > 0 {
		o.IdleTimeout = opt.IdleTimeout
	} else {
		if opt.IdleTimeout == -1 {
			o.IdleTimeout = 0
		}
	}
}

type redisResourcePool struct {
	name   string
	config *RedisConfig
	pool   *vitess_pool.ResourcePool
	mutex  sync.Mutex
}

func (rp *redisResourcePool) dial() (redis.Conn, error) {
	dialOptions := []redis.DialOption{
		redis.DialDatabase(rp.config.Options.Database),
		redis.DialConnectTimeout(rp.config.Options.ConnTimeout),
		redis.DialReadTimeout(rp.config.Options.ReadTimeout),
		redis.DialWriteTimeout(rp.config.Options.WriteTimeout),
	}

	if len(rp.config.Options.Username) != 0 {
		dialOptions = append(dialOptions, redis.DialUsername(rp.config.Options.Username))
	}

	if len(rp.config.Options.Password) != 0 {
		dialOptions = append(dialOptions, redis.DialPassword(rp.config.Options.Password))
	}

	if rp.config.Options.Dialer != nil {
		dialOptions = append(dialOptions, redis.DialContextFunc(rp.config.Options.Dialer))
	}

	if rp.config.Options.TLSConfig != nil {
		dialOptions = append(dialOptions, redis.DialTLSConfig(rp.config.Options.TLSConfig))
	}

	conn, err := redis.Dial("tcp", rp.config.Addr, dialOptions...)

	return conn, err
}

func (rp *redisResourcePool) init() {
	rp.mutex.Lock()
	defer rp.mutex.Unlock()

	if rp.pool != nil && !rp.pool.IsClosed() {
		return
	}

	df := func() (vitess_pool.Resource, error) {
		conn, err := rp.dial()
		if err != nil {
			return nil, err
		}

		return &RedisConn{
			name: rp.name,
			Conn: conn,
		}, nil
	}

	rp.pool = vitess_pool.NewResourcePool(df, rp.config.Options.PoolSize, rp.config.Options.PoolSize, rp.config.Options.IdleTimeout, rp.config.Options.PoolPrefill)
}

func (rp *redisResourcePool) Get(ctx context.Context) (*RedisConn, error) {
	if rp.pool.IsClosed() {
		rp.init()
	}

	resource, err := rp.pool.Get(ctx)
	if err != nil {
		return nil, err
	}

	rc := resource.(*RedisConn)

	// If rc is error, close and reconnect
	if err = rc.Err(); err != nil {
		logger.Warn(fmt.Sprintf("err redis.%s pool conn, reconnect", rc.name), zap.Error(err))

		conn, dialErr := rp.dial()
		if dialErr != nil {
			rp.pool.Put(rc)
			return nil, dialErr
		}

		rc.Close()

		return &RedisConn{
			name: rp.name,
			Conn: conn,
		}, nil
	}

	return rc, nil
}

func (rp *redisResourcePool) Put(conn *RedisConn) {
	rp.pool.Put(conn)
}

func (rp *redisResourcePool) Do(ctx context.Context, cmd string, args ...any) (any, error) {
	conn, err := rp.Get(ctx)
	if err != nil {
		return nil, err
	}
	defer rp.Put(conn)

	return conn.Do(cmd, args...)
}

func (rp *redisResourcePool) DoFunc(ctx context.Context, f func(ctx context.Context, conn *RedisConn) error) error {
	conn, err := rp.Get(ctx)
	if err != nil {
		return err
	}
	defer rp.Put(conn)

	return f(ctx, conn)
}

var redisMap = make(map[string]RedisPool)

func newRedisPool(name string, cfg *RedisConfig) RedisPool {
	pool := &redisResourcePool{
		name: name,
		config: &RedisConfig{
			Addr: cfg.Addr,
			Options: &RedisOptions{
				ConnTimeout:  10 * time.Second,
				ReadTimeout:  10 * time.Second,
				WriteTimeout: 10 * time.Second,
				PoolSize:     10,
				IdleTimeout:  5 * time.Minute,
			},
		},
	}

	if cfg.Options != nil {
		pool.config.Options.rebuild(cfg.Options)
	}

	pool.init()

	return pool
}

func initRedis(name string, cfg *RedisConfig) error {
	pool := newRedisPool(name, cfg)

	// verify connection
	conn, err := pool.Get(context.TODO())
	if err != nil {
		return err
	}

	if _, err = conn.Do("PING"); err != nil {
		conn.Close()
		return err
	}

	pool.Put(conn)
	redisMap[name] = pool

	return nil
}

// Redis 返回一个Redis连接池实例
func Redis(name ...string) (RedisPool, error) {
	key := Default
	if len(name) != 0 {
		key = name[0]
	}

	pool, ok := redisMap[key]
	if !ok {
		return nil, fmt.Errorf("unknown redis.%s (forgotten configure?)", key)
	}

	return pool, nil
}

// MustRedis 返回一个Redis连接池实例，如果不存在，则Panic
func MustRedis(name ...string) RedisPool {
	pool, err := Redis(name...)
	if err != nil {
		logger.Panic(err.Error())
	}

	return pool
}
