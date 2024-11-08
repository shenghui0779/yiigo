package worker

import (
	"container/list"
	"context"
	"runtime/debug"
	"sync"
	"time"
)

const defaultSize = 1000

// PanicFn 处理Panic方法
type PanicFn func(ctx context.Context, err interface{}, stack []byte)

// Job 异步执行的任务
type Job struct {
	ctx context.Context
	fn  func(ctx context.Context)
}

// Worker 控制并发协程数量
type Worker struct {
	input chan Job

	queue  chan Job
	global *list.List

	pool chan struct{}

	ctx    context.Context
	cancel context.CancelFunc

	panicFn PanicFn
}

// New 返回一个Worker实例
func New(size int, panicFn PanicFn) *Worker {
	if size <= 0 {
		size = defaultSize
	}
	ctx, cancel := context.WithCancel(context.Background())
	w := &Worker{
		input: make(chan Job),

		queue:  make(chan Job, size),
		global: list.New(),

		pool: make(chan struct{}, size),

		ctx:    ctx,
		cancel: cancel,

		panicFn: panicFn,
	}
	go w.consumer()
	go w.producer()
	go w.keepalive()
	for i := 0; i < size; i++ {
		w.pool <- struct{}{}
	}
	return w
}

// Go 异步执行任务
func (w *Worker) Go(ctx context.Context, fn func(ctx context.Context)) {
	select {
	case <-ctx.Done():
		return
	default:
	}
	w.input <- Job{ctx: ctx, fn: fn}
}

// Len 返回当前可用的协程数量
func (w *Worker) Len() int {
	return len(w.pool)
}

// Close 关闭资源
func (w *Worker) Close() {
	w.cancel()
	close(w.input)
	close(w.queue)
	close(w.pool)
}

func (w *Worker) producer() {
	for {
		select {
		case <-w.ctx.Done():
			return
		case job := <-w.input:
			select {
			case <-w.ctx.Done():
				return
			case w.queue <- job:
			default:
				w.global.PushBack(job)
			}
		}
	}
}

func (w *Worker) consumer() {
	for {
		select {
		case <-w.ctx.Done():
			return
		case <-w.pool:
			select {
			case <-w.ctx.Done():
				return
			case job := <-w.queue:
				w.do(job)
			default:
				if e := w.global.Front(); e != nil {
					if v := w.global.Remove(e); v != nil {
						w.do(v.(Job))
						break
					}
				}
				w.do(<-w.queue)
			}
		}
	}
}

// keepalive 没有它，会报：fatal error: all goroutines are asleep - deadlock!
func (w *Worker) keepalive() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-w.ctx.Done():
			return
		case <-ticker.C:
			select {
			case <-w.ctx.Done():
				return
			case w.input <- Job{}:
			default:
			}
		}
	}
}

func (w *Worker) do(job Job) {
	if job.fn == nil {
		select {
		case <-w.ctx.Done():
			return
		default:
		}
		w.pool <- struct{}{}
		return
	}
	go func(job Job) {
		defer func() {
			if e := recover(); e != nil {
				if w.panicFn != nil {
					w.panicFn(job.ctx, e, debug.Stack())
				}
			}
			select {
			case <-w.ctx.Done():
				return
			default:
			}
			w.pool <- struct{}{}
		}()
		job.fn(job.ctx)
	}(job)
}

var (
	gw   *Worker
	once sync.Once
)

// Init 初始化一个默认的全局Worker
func Init(size int, panicFn PanicFn) {
	gw = New(size, panicFn)
}

// Go 默认的全局Worker异步执行任务
func Go(ctx context.Context, fn func(ctx context.Context)) {
	if gw == nil {
		once.Do(func() {
			gw = New(defaultSize, nil)
		})
	}
	gw.Go(ctx, fn)
}

// Close 返回默认的全局Worker当前可用的协程数量
func Len() int {
	if gw == nil {
		return 0
	}
	return gw.Len()
}

// Close 关闭默认的全局Worker
func Close() {
	if gw != nil {
		gw.Close()
	}
}
