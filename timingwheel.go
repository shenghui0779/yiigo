package yiigo

import (
	"context"
	"errors"
	"fmt"
	"runtime/debug"
	"sync"
	"time"

	"go.uber.org/zap"
)

// TWHandler the function to execute when task expired.
type TWHandler func(ctx context.Context, taskID string) error

// TWDelay the function returns task next delay time.
type TWDelay func(attempts uint16) time.Duration

// TWTask timing wheel task.
type TWTask struct {
	ctx         context.Context
	round       int
	addedAt     time.Time
	remainder   time.Duration
	callback    TWHandler
	maxAttempts uint16
	attempts    uint16
	delayFunc   TWDelay
}

// TimingWheel a simple single timing wheel.
type TimingWheel interface {
	// AddOnceTask adds a task which will be executed only once when expired.
	AddOnceTask(ctx context.Context, taskID string, callback TWHandler, delay time.Duration)

	// AddRetryTask adds a task which will be executed when expired, and if an error is returned, it will be retried multiple times.
	AddRetryTask(ctx context.Context, taskID string, callback TWHandler, attempts uint16, delay TWDelay)

	// Stop stops the timing wheel.
	Stop()
}

type timewheel struct {
	slot    int
	tick    time.Duration
	size    int
	bucket  []sync.Map
	stop    chan struct{}
	taskCtx func(ctx context.Context) context.Context
	logger  CtxLogger
	debug   bool
}

func (tw *timewheel) AddOnceTask(ctx context.Context, taskID string, callback TWHandler, delay time.Duration) {
	task := &TWTask{
		ctx:         tw.taskCtx(ctx),
		callback:    callback,
		maxAttempts: 1,
		delayFunc: func(attempts uint16) time.Duration {
			return delay
		},
	}

	tw.requeue(taskID, task)
}

func (tw *timewheel) AddRetryTask(ctx context.Context, taskID string, callback TWHandler, attempts uint16, delay TWDelay) {
	task := &TWTask{
		ctx:         tw.taskCtx(ctx),
		callback:    callback,
		maxAttempts: attempts,
		delayFunc:   delay,
	}

	tw.requeue(taskID, task)
}

func (tw *timewheel) Stop() {
	select {
	case <-tw.stop:
		tw.logger.Warn(context.Background(), "TimingWheel has stoped")

		return
	default:
		close(tw.stop)
		tw.logger.Warn(context.Background(), fmt.Sprintf("TimingWheel stoped at: %s", time.Now().String()))
	}
}

func (tw *timewheel) requeue(taskID string, task *TWTask) {
	select {
	case <-tw.stop:
		tw.logger.Err(task.ctx, fmt.Sprintf("err task(%s) requeue", taskID), zap.Uint16("attempts", task.attempts+1), zap.Error(errors.New("TimingWheel has stoped")))

		return
	default:
	}

	if task.attempts >= task.maxAttempts {
		tw.logger.Warn(task.ctx, fmt.Sprintf("task(%s) attempted %d times, giving up", taskID, task.attempts), zap.Uint16("max_attempts", task.maxAttempts))

		return
	}

	task.attempts++

	duration := task.delayFunc(task.attempts)
	slot := tw.place(task, duration)

	task.addedAt = time.Now()

	if duration < tw.tick {
		go tw.run(taskID, task)

		return
	}

	tw.bucket[slot].Store(taskID, task)
}

func (tw *timewheel) place(task *TWTask, delay time.Duration) int {
	tick := tw.tick.Nanoseconds()
	total := tick * int64(tw.size)
	duration := delay.Nanoseconds()

	if duration > total {
		task.round = int(duration / total)
		duration = duration % total

		if duration == 0 {
			task.round--
		}
	}

	task.remainder = time.Duration(duration % tick)

	return (tw.slot + int(duration/tick)) % tw.size
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

		taskID := key.(string)
		task := value.(*TWTask)

		if task.round > 0 {
			task.round--

			return true
		}

		go tw.run(taskID, task)

		tw.bucket[slot].Delete(key)

		return true
	})
}

func (tw *timewheel) run(taskID string, task *TWTask) {
	if task.remainder > 0 {
		time.Sleep(task.remainder)
	}

	delay := time.Since(task.addedAt).String()

	defer func() {
		if err := recover(); err != nil {
			tw.logger.Err(task.ctx, fmt.Sprintf("task(%s) run panic", taskID), zap.Uint16("attempts", task.attempts), zap.String("delay", delay), zap.Any("error", err), zap.ByteString("stack", debug.Stack()))

			if task.attempts < task.maxAttempts {
				tw.requeue(taskID, task)
			}
		}
	}()

	if err := task.callback(task.ctx, taskID); err != nil {
		tw.logger.Err(task.ctx, fmt.Sprintf("err task(%s) run", taskID), zap.Uint16("attempts", task.attempts), zap.String("delay", delay), zap.Error(err))

		if task.attempts < task.maxAttempts {
			tw.requeue(taskID, task)
		}

		return
	}

	if tw.debug {
		tw.logger.Info(task.ctx, fmt.Sprintf("task(%s) run ok", taskID), zap.Uint16("attempts", task.attempts), zap.String("delay", delay))
	}
}

// TWOption timing wheel option.
type TWOption func(tw *timewheel)

// WithTaskCtx clones context for executing tasks asynchronously, the default is `context.Background()`.
func WithTaskCtx(fn func(ctx context.Context) context.Context) TWOption {
	return func(tw *timewheel) {
		tw.taskCtx = fn
	}
}

// WithTWLogger specifies logger for timing wheel.
func WithTWLogger(l CtxLogger) TWOption {
	return func(tw *timewheel) {
		tw.logger = l
	}
}

// WithTWDebug specifies debug mode for timing wheel.
func WithTWDebug() TWOption {
	return func(tw *timewheel) {
		tw.debug = true
	}
}

type twLogger struct{}

func (l *twLogger) Info(ctx context.Context, msg string, fields ...zap.Field) {
	logger.Info(fmt.Sprintf("[tw] %s", msg), fields...)
}

func (l *twLogger) Warn(ctx context.Context, msg string, fields ...zap.Field) {
	logger.Warn(fmt.Sprintf("[tw] %s", msg), fields...)
}

func (l *twLogger) Err(ctx context.Context, msg string, fields ...zap.Field) {
	logger.Error(fmt.Sprintf("[tw] %s", msg), fields...)
}

// NewTimingWheel returns a new timing wheel.
func NewTimingWheel(tick time.Duration, size int, options ...TWOption) TimingWheel {
	tw := &timewheel{
		tick:   tick,
		size:   size,
		bucket: make([]sync.Map, size),
		stop:   make(chan struct{}),
		taskCtx: func(ctx context.Context) context.Context {
			return context.Background()
		},
		logger: new(twLogger),
	}

	for _, f := range options {
		f(tw)
	}

	go tw.scheduler()

	return tw
}
