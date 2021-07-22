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

// GRPCConn grpc connection resource
type GRPCConn struct {
	*grpc.ClientConn
}

// Close close connection resorce
func (gc GRPCConn) Close() {
	if err := gc.ClientConn.Close(); err != nil {
		logger.Error("yiigo: grpc client conn closed error", zap.Error(err))
	}
}

// PoolSettings pool settings
type PoolSettings struct {
	poolSize    int
	poolLimit   int
	idleTimeout time.Duration
	poolPrefill int
}

// PoolOption configures how we set up the pool.
type PoolOption func(s *PoolSettings)

// WithPoolSize specifies the number of possible resources in the pool.
func WithPoolSize(size int) PoolOption {
	return func(s *PoolSettings) {
		s.poolSize = size
	}
}

// WithPoolLimit specifies the extent to which the pool can be resized in the future.
// You cannot resize the pool beyond poolLimit.
func WithPoolLimit(limit int) PoolOption {
	return func(s *PoolSettings) {
		s.poolLimit = limit
	}
}

// WithIdleTimeout specifies the maximum amount of time a connection may be idle.
// An idleTimeout of 0 means that there is no timeout.
func WithIdleTimeout(duration time.Duration) PoolOption {
	return func(s *PoolSettings) {
		s.idleTimeout = duration
	}
}

// WithPoolPrefill specifies how many resources can be opened in parallel.
func WithPoolPrefill(prefill int) PoolOption {
	return func(s *PoolSettings) {
		s.poolPrefill = prefill
	}
}

// GRPCPoolResource grpc pool resource
type GRPCPoolResource struct {
	dialFunc func() (*grpc.ClientConn, error)
	settings *PoolSettings
	pool     *vitess_pool.ResourcePool
	mutex    sync.Mutex
}

func (r *GRPCPoolResource) init() {
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

		return GRPCConn{conn}, nil
	}

	r.pool = vitess_pool.NewResourcePool(df, r.settings.poolSize, r.settings.poolLimit, r.settings.idleTimeout, r.settings.poolPrefill)
}

// Get get a connection resource from the pool.
// Context with timeout can specify the wait timeout for pool.
func (r *GRPCPoolResource) Get(ctx context.Context) (GRPCConn, error) {
	if r.pool.IsClosed() {
		r.init()
	}

	resource, err := r.pool.Get(ctx)

	if err != nil {
		return GRPCConn{}, err
	}

	rc := resource.(GRPCConn)

	// If rc is in unexpected state, close and reconnect
	if state := rc.GetState(); state == connectivity.TransientFailure || state == connectivity.Shutdown {
		conn, err := r.dialFunc()

		if err != nil {
			r.pool.Put(rc)

			return rc, err
		}

		rc.Close()

		return GRPCConn{conn}, nil
	}

	return rc, nil
}

// Put returns a connection resource to the pool.
func (r *GRPCPoolResource) Put(rc GRPCConn) {
	r.pool.Put(rc)
}

var grpcMap sync.Map

// SetGRPCPool set an grpc pool
func SetGRPCPool(name string, dial func() (*grpc.ClientConn, error), options ...PoolOption) {
	rc := &GRPCPoolResource{
		dialFunc: dial,
		settings: &PoolSettings{
			poolSize:    10,
			poolLimit:   20,
			idleTimeout: 60,
		},
	}

	for _, f := range options {
		f(rc.settings)
	}

	rc.init()

	grpcMap.Store(name, rc)
}

// GetGRPCPool get an grpc pool
func GetGRPCPool(name string) *GRPCPoolResource {
	v, ok := grpcMap.Load(name)

	if !ok {
		logger.Panic(fmt.Sprintf("yiigo: unknown grpc.%s (forgotten set?)", name))
	}

	return v.(*GRPCPoolResource)
}
