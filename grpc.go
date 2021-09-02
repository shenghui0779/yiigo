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

// GRPCConn grpc connection resource
type GRPCConn struct {
	*grpc.ClientConn
}

// Close closes the connection resource
func (gc *GRPCConn) Close() {
	if err := gc.ClientConn.Close(); err != nil {
		logger.Error("yiigo: grpc client conn closed error", zap.Error(err))
	}
}

// poolSetting pool setting
type poolSetting struct {
	size        int
	limit       int
	idleTimeout time.Duration
	prefill     int
}

// PoolOption configures how we set up the pool.
type PoolOption func(s *poolSetting)

// WithPoolSize specifies the number of possible resources in the pool.
func WithPoolSize(size int) PoolOption {
	return func(s *poolSetting) {
		s.size = size
	}
}

// WithPoolLimit specifies the extent to which the pool can be resized in the future.
// You cannot resize the pool beyond poolLimit.
func WithPoolLimit(limit int) PoolOption {
	return func(s *poolSetting) {
		s.limit = limit
	}
}

// WithPoolIdleTimeout specifies the maximum amount of time a connection may be idle.
// An idleTimeout of 0 means that there is no timeout.
func WithPoolIdleTimeout(duration time.Duration) PoolOption {
	return func(s *poolSetting) {
		s.idleTimeout = duration
	}
}

// WithPoolPrefill specifies how many resources can be opened in parallel.
func WithPoolPrefill(prefill int) PoolOption {
	return func(s *poolSetting) {
		s.prefill = prefill
	}
}

// GRPCPool grpc pool resource
type GRPCPool interface {
	// Get returns a connection resource from the pool.
	// Context with timeout can specify the wait timeout for pool.
	Get(ctx context.Context) (*GRPCConn, error)

	// Put returns a connection resource to the pool.
	Put(gc *GRPCConn)
}

// GRPCDialFunc grpc dial function
type GRPCDialFunc func() (*grpc.ClientConn, error)

type gRPCPoolResource struct {
	dialFunc GRPCDialFunc
	config   *poolSetting
	pool     *vitess_pool.ResourcePool
	mutex    sync.Mutex
}

func (r *gRPCPoolResource) init() {
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

		return &GRPCConn{conn}, nil
	}

	r.pool = vitess_pool.NewResourcePool(df, r.config.size, r.config.limit, r.config.idleTimeout, r.config.prefill)
}

func (r *gRPCPoolResource) Get(ctx context.Context) (*GRPCConn, error) {
	if r.pool.IsClosed() {
		r.init()
	}

	resource, err := r.pool.Get(ctx)

	if err != nil {
		return &GRPCConn{}, err
	}

	gc := resource.(*GRPCConn)

	// If rc is in unexpected state, close and reconnect
	if state := gc.GetState(); state == connectivity.TransientFailure || state == connectivity.Shutdown {
		conn, err := r.dialFunc()

		if err != nil {
			r.pool.Put(gc)

			return gc, err
		}

		gc.Close()

		return &GRPCConn{conn}, nil
	}

	return gc, nil
}

func (r *gRPCPoolResource) Put(conn *GRPCConn) {
	r.pool.Put(conn)
}

// NewGRPCPool returns a new grpc pool with dial func.
func NewGRPCPool(dial GRPCDialFunc, options ...PoolOption) GRPCPool {
	rp := &gRPCPoolResource{
		dialFunc: dial,
		config: &poolSetting{
			size:        10,
			idleTimeout: 60,
		},
	}

	for _, f := range options {
		f(rp.config)
	}

	if rp.config.limit < rp.config.size {
		rp.config.limit = rp.config.size
	}

	rp.init()

	return rp
}
