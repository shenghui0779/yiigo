package errgroup

import (
	"context"
	"fmt"
	"runtime/debug"
	"sync"
)

// A Group is a collection of goroutines working on subtasks that are part of
// the same overall task.
//
// A zero Group is valid, has no limit on the number of active goroutines,
// and does not cancel on error. use WithContext instead.
type Group struct {
	err     error
	wg      sync.WaitGroup
	errOnce sync.Once

	workerOnce sync.Once
	ch         chan func(ctx context.Context) error
	pool       []func(ctx context.Context) error

	ctx    context.Context
	cancel context.CancelCauseFunc
}

// WithContext returns a new Group with a canceled Context derived from ctx.
//
// The derived Context is canceled the first time a function passed to Go
// returns a non-nil error or the first time Wait returns, whichever occurs first.
func WithContext(ctx context.Context) *Group {
	ctx, cancel := context.WithCancelCause(ctx)
	return &Group{ctx: ctx, cancel: cancel}
}

func (g *Group) do(fn func(ctx context.Context) error) {
	ctx := g.ctx
	if ctx == nil {
		ctx = context.Background()
	}
	var err error
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("errgroup: panic recovered: %+v\n%s", r, string(debug.Stack()))
		}
		if err != nil {
			g.errOnce.Do(func() {
				g.err = err
				if g.cancel != nil {
					g.cancel(err)
				}
			})
		}
		g.wg.Done()
	}()
	err = fn(ctx)
}

// GOMAXPROCS set max goroutine to work.
func (g *Group) GOMAXPROCS(n int) {
	if n <= 0 {
		return
	}
	g.workerOnce.Do(func() {
		g.ch = make(chan func(context.Context) error, n)
		for i := 0; i < n; i++ {
			go func() {
				for fn := range g.ch {
					g.do(fn)
				}
			}()
		}
	})
}

// Go calls the given function in a new goroutine.
//
// The first call to return a non-nil error cancels the group; its error will be
// returned by Wait.
func (g *Group) Go(fn func(ctx context.Context) error) {
	g.wg.Add(1)
	if g.ch != nil {
		select {
		case g.ch <- fn:
		default:
			g.pool = append(g.pool, fn)
		}
		return
	}
	go g.do(fn)
}

// Wait blocks until all function calls from the Go method have returned, then
// returns the first non-nil error (if any) from them.
func (g *Group) Wait() error {
	if g.ch != nil {
		for _, fn := range g.pool {
			g.ch <- fn
		}
	}
	g.wg.Wait()
	if g.ch != nil {
		close(g.ch) // let all receiver exit
	}
	if g.cancel != nil {
		g.cancel(g.err)
	}
	return g.err
}
