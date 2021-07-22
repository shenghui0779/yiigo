package yiigo

import (
	"context"
	"sync"
	"time"

	"github.com/shenghui0779/vitess_pool"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
)

type GrpcConn struct {
	*grpc.ClientConn
}

func (gc GrpcConn) Close() {
	if err := gc.ClientConn.Close(); err != nil {
		logger.Error("yiigo: grpc client conn closed error", zap.Error(err))
	}
}

type PoolSettings struct {
	poolSize    int
	poolLimit   int
	idleTimeout time.Duration
	poolPrefill int
}

type PoolOption func(s *PoolSettings)

func WithPoolSize(size int) PoolOption {
	return func(s *PoolSettings) {
		s.poolSize = size
	}
}

func WithPoolLimit(limit int) PoolOption {
	return func(s *PoolSettings) {
		s.poolLimit = limit
	}
}

func WithIdleTimeout(duration time.Duration) PoolOption {
	return func(s *PoolSettings) {
		s.idleTimeout = duration
	}
}

func WithPoolPrefill(prefill int) PoolOption {
	return func(s *PoolSettings) {
		s.poolPrefill = prefill
	}
}

type GrpcPoolResource struct {
	dialFunc func() (*grpc.ClientConn, error)
	settings *PoolSettings
	pool     *vitess_pool.ResourcePool
	mutex    sync.Mutex
}

func (r *GrpcPoolResource) init() {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if r.pool != nil && !r.pool.IsClosed() {
		return
	}

	df := func() (vitess_pool.Resource, error) {
		conn, err := r.dialFunc()

		if err != nil {
			return nil, err
		}

		return GrpcConn{conn}, nil
	}

	r.pool = vitess_pool.NewResourcePool(df, r.settings.poolSize, r.settings.poolLimit, r.settings.idleTimeout, r.settings.poolPrefill)
}

// Get get a connection resource from the pool.
// Context with timeout can specify the wait timeout for pool.
func (r *GrpcPoolResource) Get(ctx context.Context) (GrpcConn, error) {
	if r.pool.IsClosed() {
		r.init()
	}

	resource, err := r.pool.Get(ctx)

	if err != nil {
		return GrpcConn{}, err
	}

	rc := resource.(GrpcConn)

	// If rc is in, close and reconnect
	if state := rc.GetState(); state == connectivity.TransientFailure || state == connectivity.Shutdown {
		conn, err := r.dialFunc()

		if err != nil {
			r.pool.Put(rc)

			return rc, err
		}

		rc.Close()

		return GrpcConn{conn}, nil
	}

	return rc, nil
}

// Put returns a connection resource to the pool.
func (r *GrpcPoolResource) Put(rc GrpcConn) {
	r.pool.Put(rc)
}
