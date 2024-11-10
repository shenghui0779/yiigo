package worker

import (
	"container/list"
	"context"
	"runtime/debug"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
)

const (
	defaultCap         = 1000
	defaultIdleTimeout = 60 * time.Second
)

// PanicFn 处理Panic方法
type PanicFn func(ctx context.Context, err interface{}, stack []byte)

type worker struct {
	timeUsed time.Time
	cancel   context.CancelFunc
}

// Job 异步执行的任务
type Job struct {
	ctx context.Context
	fn  func(ctx context.Context)
}

// Limiter 控制并发协程数量
type Limiter struct {
	input  chan Job
	queue  chan Job
	global *list.List

	capacity int64
	active   int64
	workers  sync.Map

	idleTimeout time.Duration
	panicFn     PanicFn

	ctx    context.Context
	cancel context.CancelFunc
}

// New 返回一个Worker实例
func New(cap int64, idleTimeout time.Duration, panicFn PanicFn) *Limiter {
	if cap <= 0 {
		cap = defaultCap
	}
	ctx, cancel := context.WithCancel(context.TODO())
	l := &Limiter{
		input:  make(chan Job),
		queue:  make(chan Job, cap),
		global: list.New(),

		capacity: cap,

		idleTimeout: idleTimeout,
		panicFn:     panicFn,

		ctx:    ctx,
		cancel: cancel,
	}
	go l.run()
	go l.idleCheck()
	return l
}

// Go 异步执行任务
func (l *Limiter) Go(ctx context.Context, fn func(ctx context.Context)) {
	select {
	case <-ctx.Done():
		return
	case l.input <- Job{ctx: ctx, fn: fn}:
	}
}

// Active 返回当前活跃的协程数量
func (l *Limiter) Active() int64 {
	return l.active
}

// Close 关闭资源
func (l *Limiter) Close() {
	l.cancel()
	close(l.input)
	close(l.queue)
}

func (l *Limiter) setTimeUsed(uniqId string) {
	v, ok := l.workers.Load(uniqId)
	if !ok || v == nil {
		return
	}
	v.(*worker).timeUsed = time.Now()
}

func (l *Limiter) run() {
	for {
		select {
		case <-l.ctx.Done():
			return
		case job := <-l.input:
			select {
			case l.queue <- job:
			default:
				if l.active < l.capacity {
					atomic.AddInt64(&l.active, 1)
					ctx, cancel := context.WithCancel(context.TODO())
					uniqId := l.spawn(ctx)
					l.workers.Store(uniqId, &worker{
						timeUsed: time.Now(),
						cancel:   cancel,
					})
				}
				l.global.PushBack(job)
			}
		}
	}
}

func (l *Limiter) idleCheck() {
	ticker := time.NewTicker(l.idleTimeout)
	defer ticker.Stop()

	for {
		select {
		case <-l.ctx.Done():
			return
		case <-ticker.C:
			l.workers.Range(func(k, v any) bool {
				w := v.(*worker)
				if l.idleTimeout > 0 && time.Since(w.timeUsed) > l.idleTimeout {
					w.cancel()
					l.workers.Delete(k)
				}
				return true
			})
		}
	}
}

func (l *Limiter) spawn(ctx context.Context) string {
	uniqId := uuid.New().String()

	go func(ctx context.Context, uniqId string) {
		var jobCtx context.Context
		defer func() {
			if e := recover(); e != nil {
				if l.panicFn != nil {
					l.panicFn(jobCtx, e, debug.Stack())
				}
			}
		}()
		for {
			var job Job
			select {
			case <-l.ctx.Done():
				return
			case <-ctx.Done():
				atomic.AddInt64(&l.active, -1)
				return
			case job = <-l.queue:
			default:
				if e := l.global.Front(); e != nil {
					if v := l.global.Remove(e); v != nil {
						job = v.(Job)
						break
					}
				}
				select {
				case <-l.ctx.Done():
					return
				case <-ctx.Done():
					atomic.AddInt64(&l.active, -1)
					return
				case job = <-l.queue:
				}
			}
			l.setTimeUsed(uniqId)
			jobCtx = job.ctx
			job.fn(job.ctx)
		}
	}(ctx, uniqId)

	return uniqId
}

var (
	gw   *Limiter
	once sync.Once
)

// Init 初始化默认的全局Limiter
func Init(cap int64, idleTimeout time.Duration, panicFn PanicFn) {
	gw = New(cap, idleTimeout, panicFn)
}

// Go 异步执行任务
func Go(ctx context.Context, fn func(ctx context.Context)) {
	if gw == nil {
		once.Do(func() {
			gw = New(defaultCap, defaultIdleTimeout, nil)
		})
	}
	gw.Go(ctx, fn)
}

// Active 返回当前活跃的协程数量
func Active() int64 {
	if gw == nil {
		return 0
	}
	return gw.Active()
}

// Close 关闭默认的全局Limiter
func Close() {
	if gw != nil {
		gw.Close()
	}
}
