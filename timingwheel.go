package yiigo

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
)

type ctxTWKey int

// CtxTaskAddedAt Context存储任务入队时间的Key
const CtxTaskAddedAt ctxTWKey = 0

// TWTask 时间轮任务
type TWTask struct {
	ctx         context.Context
	uniqID      string
	round       int
	attempts    uint16        // 当前尝试的次数
	maxAttempts uint16        // 最大尝试次数
	remainder   time.Duration // 任务执行前的剩余延迟（小于时间轮精度）
	cumulative  int64         // 多次重试的累计时长（单位：ns）
	deferFn     func(attempts uint16) time.Duration
	callback    func(ctx context.Context, taskID string) error
}

// TimingWheel 一个简易个单时间轮
type TimingWheel interface {
	// AddTask 添加一个任务，到期被执行，默认仅执行一次；若指定了重试次数，则在发生错误后重试
	// 注意：任务是异步执行的，故Context应该是一个克隆的且不带超时时间的
	AddTask(ctx context.Context, taskID string, handler func(ctx context.Context, taskID string) error, options ...TaskOption)

	// Stop 终止时间轮
	Stop()
}

type timewheel struct {
	slot   int
	tick   time.Duration
	size   int
	bucket []sync.Map
	stop   chan struct{}
	log    func(ctx context.Context, v ...any)
}

func (tw *timewheel) AddTask(ctx context.Context, taskID string, handler func(ctx context.Context, taskID string) error, options ...TaskOption) {
	task := &TWTask{
		ctx:         context.WithValue(ctx, CtxTaskAddedAt, time.Now()),
		uniqID:      taskID,
		callback:    handler,
		maxAttempts: 1,
		deferFn: func(attempts uint16) time.Duration {
			return 0
		},
	}

	for _, f := range options {
		f(task)
	}

	tw.requeue(task)
}

func (tw *timewheel) Stop() {
	select {
	case <-tw.stop:
		tw.log(context.Background(), "timingwheel stopped")

		return
	default:
	}

	close(tw.stop)

	tw.log(context.Background(), fmt.Sprintf("timingwheel stoped at: %s", time.Now().String()))
}

func (tw *timewheel) requeue(task *TWTask) {
	if task.attempts >= task.maxAttempts {
		return
	}

	select {
	case <-tw.stop:
		tw.log(task.ctx, fmt.Sprintf("task(%s) attempt(%d) failed, because the timingwheel has stopped", task.uniqID, task.attempts+1))

		return
	default:
	}

	task.ctx = context.WithValue(task.ctx, CtxTaskAddedAt, time.Now())

	task.attempts++

	tick := tw.tick.Nanoseconds()
	delay := task.deferFn(task.attempts)
	duration := delay.Nanoseconds()

	task.cumulative += duration
	task.round = int(duration / (tick * int64(tw.size)))

	slot := int(task.cumulative/tick) % tw.size

	if slot == tw.slot {
		if task.round == 0 {
			task.remainder = delay
			go tw.run(task)

			return
		}

		task.round--
	}

	task.remainder = time.Duration(task.cumulative % tick)

	tw.bucket[slot].Store(task.uniqID, task)
}

func (tw *timewheel) scheduler() {
	ticker := time.NewTicker(tw.tick)
	defer ticker.Stop()

	for {
		select {
		case <-tw.stop:
			return
		case <-ticker.C:
			tw.slot = (tw.slot + 1) % tw.size
			go tw.process(tw.slot)
		}
	}
}

func (tw *timewheel) process(slot int) {
	tw.bucket[slot].Range(func(key, value any) bool {
		select {
		case <-tw.stop:
			return false
		default:
		}

		task := value.(*TWTask)

		if task.round > 0 {
			task.round--

			return true
		}

		go tw.run(task)

		tw.bucket[slot].Delete(key)

		return true
	})
}

func (tw *timewheel) run(task *TWTask) {
	if task.remainder > 0 {
		time.Sleep(task.remainder)
	}

	defer func() {
		if err := recover(); err != nil {
			tw.log(task.ctx, fmt.Sprintf("task(%s) run panic: %v", task.uniqID, err))
		}
	}()

	if err := task.callback(task.ctx, task.uniqID); err != nil {
		tw.log(task.ctx, fmt.Sprintf("task(%s) run error: %v", task.uniqID, err))
		tw.requeue(task)

		return
	}
}

// TWOption 时间轮选项
type TWOption func(tw *timewheel)

// WithTWLogger 指定时间轮日志
func WithTWLogger(fn func(ctx context.Context, v ...any)) TWOption {
	return func(tw *timewheel) {
		tw.log = fn
	}
}

// TaskOption 时间轮任务选项
type TaskOption func(t *TWTask)

// WithTaskAttempts 指定任务重试次数；默认：1
func WithTaskAttempts(attempts uint16) TaskOption {
	return func(t *TWTask) {
		t.maxAttempts = attempts
	}
}

// WithTaskDefer 指定任务延迟执行时间；默认：立即执行
func WithTaskDefer(fn func(attempts uint16) time.Duration) TaskOption {
	return func(t *TWTask) {
		t.deferFn = fn
	}
}

// NewTimingWheel 返回一个时间轮实例
func NewTimingWheel(tick time.Duration, size int, options ...TWOption) TimingWheel {
	tw := &timewheel{
		tick:   tick,
		size:   size,
		bucket: make([]sync.Map, size),
		stop:   make(chan struct{}),
		log: func(ctx context.Context, v ...any) {
			logger.Error("err timingwheel", zap.String("err", fmt.Sprint(v...)))
		},
	}

	for _, f := range options {
		f(tw)
	}

	go tw.scheduler()

	return tw
}
