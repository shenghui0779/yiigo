package yiigo

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/shenghui0779/vitess_pool"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
)

// GRPCClientDialer grpc client dial function
type GRPCClientDialer func() (*grpc.ClientConn, error)

// GRPCConn grpc connection resource
type GRPCConn struct {
	*grpc.ClientConn
}

// Close closes the connection resource
func (gc *GRPCConn) Close() {
	if err := gc.ClientConn.Close(); err != nil {
		logger.Error("[yiigo] grpc client conn closed error", zap.Error(err))
	}
}

// GRPCClientPool grpc client pool resource
type GRPCClientPool interface {
	// Get returns a connection resource from the pool.
	// Context with timeout can specify the wait timeout for pool.
	Get(ctx context.Context) (*GRPCConn, error)

	// Put returns a connection resource to the pool.
	Put(gc *GRPCConn)
}

// GRPCPoolConfig keeps the settings to setup grpc client connection pool.
type GRPCPoolConfig struct {
	// Dialer is a function that can be used to create a client connection.
	Dialer GRPCClientDialer `json:"dialer"`

	// Options optional settings to setup grpc client connection pool.
	Options *PoolOptions `json:"options"`
}

// PoolOptions optional settings to setup db connection.
type PoolOptions struct {
	// PoolSize is the maximum number of possible resources in the pool.
	// Use value -1 for no timeout and 0 for default.
	// Default is 10.
	PoolSize int `json:"pool_size"`

	// PoolPrefill is the number of resources to be pre-filled in the pool.
	// Default is no pre-filled.
	PoolPrefill int `json:"pool_prefill"`

	// IdleTimeout is the amount of time after which client closes idle connections.
	// Use value -1 for no timeout and 0 for default.
	// Default is 5 minutes.
	IdleTimeout time.Duration `json:"idle_timeout"`
}

func (o *PoolOptions) rebuild(opt *PoolOptions) {
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

type gRPCResourcePool struct {
	config *GRPCPoolConfig
	pool   *vitess_pool.ResourcePool
	mutex  sync.Mutex
}

func (rp *gRPCResourcePool) init() {
	rp.mutex.Lock()
	defer rp.mutex.Unlock()

	if rp.pool != nil && !rp.pool.IsClosed() {
		return
	}

	df := func() (vitess_pool.Resource, error) {
		conn, err := rp.config.Dialer()

		if err != nil {
			return nil, err
		}

		return &GRPCConn{conn}, nil
	}

	rp.pool = vitess_pool.NewResourcePool(df, rp.config.Options.PoolSize, rp.config.Options.PoolSize, rp.config.Options.IdleTimeout, rp.config.Options.PoolPrefill)
}

func (rp *gRPCResourcePool) Get(ctx context.Context) (*GRPCConn, error) {
	if rp.pool.IsClosed() {
		rp.init()
	}

	resource, err := rp.pool.Get(ctx)

	if err != nil {
		return nil, err
	}

	rc := resource.(*GRPCConn)

	// If rc is in unexpected state, close and reconnect
	if state := rc.GetState(); state == connectivity.TransientFailure || state == connectivity.Shutdown {
		logger.Warn(fmt.Sprintf("[yiigo] grpc pool conn is %s, reconnect", state.String()))

		conn, err := rp.config.Dialer()

		if err != nil {
			rp.pool.Put(rc)

			return nil, err
		}

		rc.Close()

		return &GRPCConn{conn}, nil
	}

	return rc, nil
}

func (rp *gRPCResourcePool) Put(conn *GRPCConn) {
	rp.pool.Put(conn)
}

// NewGRPCClientPool returns a new grpc client connection pool.
func NewGRPCClientPool(cfg *GRPCPoolConfig) GRPCClientPool {
	pool := &gRPCResourcePool{
		config: &GRPCPoolConfig{
			Dialer: cfg.Dialer,
			Options: &PoolOptions{
				PoolSize:    10,
				IdleTimeout: 5 * time.Minute,
			},
		},
	}

	if cfg.Options != nil {
		pool.config.Options.rebuild(cfg.Options)
	}

	pool.init()

	return pool
}
