package yiigo

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
)

type ctxTWKey int

// CtxTaskAddedAt is the key that holds the time in a task context when the task added to timingwheel.
const CtxTaskAddedAt ctxTWKey = 0

// TWTask timingwheel task.
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

// TimingWheel a simple single timingwheel.
type TimingWheel interface {
	// AddTask adds a task which will be executed when expired, it will be retried if attempts > 0 and an error is returned.
	// NOTE: Context should be cloned without timeout for executing tasks asynchronously.
	AddTask(ctx context.Context, taskID string, handler func(ctx context.Context, taskID string) error, options ...TaskOption)

	// Stop stops the timingwheel.
	Stop()
}

type timewheel struct {
	slot   int
	tick   time.Duration
	size   int
	bucket []sync.Map
	stop   chan struct{}
	log    func(ctx context.Context, v ...interface{})
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
	tw.bucket[slot].Range(func(key, value interface{}) bool {
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

// TWOption timingwheel option.
type TWOption func(tw *timewheel)

// WithTWLogger specifies logger for timingwheel.
func WithTWLogger(fn func(ctx context.Context, v ...interface{})) TWOption {
	return func(tw *timewheel) {
		tw.log = fn
	}
}

// TaskOption timingwheel task option.
type TaskOption func(t *TWTask)

// WithTaskAttempts specifies attempt count for timingwheel task and 1 for default.
func WithTaskAttempts(attempts uint16) TaskOption {
	return func(t *TWTask) {
		t.maxAttempts = attempts
	}
}

// WithTaskDefer specifies the task to be executed until the timeout expires and 0 for default.
func WithTaskDefer(fn func(attempts uint16) time.Duration) TaskOption {
	return func(t *TWTask) {
		t.deferFn = fn
	}
}

// NewTimingWheel returns a new timingwheel.
func NewTimingWheel(tick time.Duration, size int, options ...TWOption) TimingWheel {
	tw := &timewheel{
		tick:   tick,
		size:   size,
		bucket: make([]sync.Map, size),
		stop:   make(chan struct{}),
		log: func(ctx context.Context, v ...interface{}) {
			logger.Error("err timingwheel", zap.String("err", fmt.Sprint(v...)))
		},
	}

	for _, f := range options {
		f(tw)
	}

	go tw.scheduler()

	return tw
}
