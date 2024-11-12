package gopool

import (
	"context"
	"runtime/debug"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/shenghui0779/yiigo/gopool/linklist"
)

const (
	defaultPoolCap     = 10000
	defaultIdleTimeout = 60 * time.Second
)

// PanicFn 处理Panic方法
type PanicFn func(ctx context.Context, err any, stack []byte)

type worker struct {
	timeUsed time.Time
	cancel   context.CancelFunc
}

// Job 异步执行的任务
type Job struct {
	ctx context.Context
	fn  func(ctx context.Context)
}

// Pool 协程池，控制并发协程数量，降低CPU和内存负载
type Pool struct {
	input chan *Job
	queue chan *Job
	cache *linklist.List[*Job]

	capacity int
	workers  map[string]*worker
	mutex    sync.Mutex

	prefill     int
	nonBlock    bool
	queueCap    int
	idleTimeout time.Duration
	panicFn     PanicFn

	ctx    context.Context
	cancel context.CancelFunc
}

// New 返回一个Worker实例
func New(cap int, opts ...Option) *Pool {
	if cap <= 0 {
		cap = defaultPoolCap
	}

	ctx, cancel := context.WithCancel(context.TODO())
	p := &Pool{
		input: make(chan *Job),
		cache: linklist.New[*Job](),

		capacity: cap,
		workers:  make(map[string]*worker, cap),

		idleTimeout: defaultIdleTimeout,

		ctx:    ctx,
		cancel: cancel,
	}

	for _, fn := range opts {
		fn(p)
	}
	p.queue = make(chan *Job, p.queueCap)
	// 预填充
	if p.prefill > 0 {
		count := p.prefill
		if p.prefill > p.capacity {
			count = p.capacity
		}
		for i := 0; i < count; i++ {
			c, fn := context.WithCancel(context.TODO())
			uniqId := p.spawn(c)
			p.workers[uniqId] = &worker{
				timeUsed: time.Now(),
				cancel:   fn,
			}
		}
	}

	go p.run()
	go p.idleCheck()

	return p
}

// Go 异步执行任务
func (p *Pool) Go(ctx context.Context, fn func(ctx context.Context)) {
	select {
	case <-ctx.Done():
		return
	case p.input <- &Job{ctx: ctx, fn: fn}:
	}
}

// Close 关闭资源
func (p *Pool) Close() {
	// 销毁协程
	p.cancel()
	// 关闭通道
	close(p.input)
	close(p.queue)
}

func (p *Pool) setTimeUsed(uniqId string) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	v, ok := p.workers[uniqId]
	if !ok || v == nil {
		return
	}
	v.timeUsed = time.Now()
}

func (p *Pool) run() {
	for {
		select {
		case <-p.ctx.Done():
			return
		case job := <-p.input:
			select {
			case p.queue <- job:
			default:
				if len(p.workers) < p.capacity {
					// 新开一个协程
					ctx, cancel := context.WithCancel(context.TODO())
					uniqId := p.spawn(ctx)
					// 存储协程信息
					p.mutex.Lock()
					p.workers[uniqId] = &worker{
						timeUsed: time.Now(),
						cancel:   cancel,
					}
					p.mutex.Unlock()
				}
				if p.nonBlock {
					// 非阻塞模式，放入本地缓存
					p.cache.Append(job)
				} else {
					// 阻塞模式，等待闲置协程
					select {
					case <-p.ctx.Done():
						return
					case p.queue <- job:
					}
				}
			}
		}
	}
}

func (p *Pool) idleCheck() {
	ticker := time.NewTicker(p.idleTimeout)
	defer ticker.Stop()

	for {
		select {
		case <-p.ctx.Done():
			return
		case <-ticker.C:
			p.mutex.Lock()
			for k, v := range p.workers {
				if p.idleTimeout > 0 && time.Since(v.timeUsed) > p.idleTimeout {
					v.cancel()
					delete(p.workers, k)
				}
			}
			p.mutex.Unlock()
		}
	}
}

func (p *Pool) spawn(ctx context.Context) string {
	uniqId := strings.ReplaceAll(uuid.New().String(), "-", "")

	go func(ctx context.Context, uniqId string) {
		var jobCtx context.Context
		defer func() {
			if e := recover(); e != nil {
				if p.panicFn != nil {
					p.panicFn(jobCtx, e, debug.Stack())
				}
			}
		}()
		for {
			var job *Job
			// 获取任务
			select {
			case <-p.ctx.Done(): // Pool关闭，销毁
				return
			case <-ctx.Done(): // 闲置超时，销毁
				return
			case job = <-p.queue:
			default:
				// 非阻塞模式，去取缓存的任务执行
				if p.nonBlock {
					job, _ = p.cache.Pop(0)
					if job != nil {
						break
					}
				}
				// 阻塞模式或未取到缓存任务，则等待新任务
				select {
				case <-p.ctx.Done():
					return
				case <-ctx.Done():
					return
				case job = <-p.queue:
				}
			}
			// 执行任务
			p.setTimeUsed(uniqId)
			jobCtx = job.ctx
			job.fn(job.ctx)
		}
	}(ctx, uniqId)

	return uniqId
}

var (
	pool *Pool
	once sync.Once
)

// Init 初始化默认的全局Pool
func Init(cap int, opts ...Option) {
	pool = New(cap, opts...)
}

// Go 异步执行任务
func Go(ctx context.Context, fn func(ctx context.Context)) {
	if pool == nil {
		once.Do(func() {
			pool = New(defaultPoolCap)
		})
	}
	pool.Go(ctx, fn)
}

// Close 关闭默认的全局Pool
func Close() {
	if pool != nil {
		pool.Close()
	}
}
