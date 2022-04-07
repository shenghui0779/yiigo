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

// TWTask timing wheel task.
type TWTask struct {
	ctx       context.Context
	round     int
	addedAt   time.Time
	remainder time.Duration
	callback  TWHandler
}

// TWOption timing wheel option.
type TWOption func(tw *TimingWheel)

// WithTaskCtx clones context for executing tasks asynchronously, the default is `context.Background()`.
func WithTaskCtx(fn func(ctx context.Context) context.Context) TWOption {
	return func(tw *TimingWheel) {
		tw.taskCtx = fn
	}
}

// WithTWLogger specifies the logger for timing wheel.
func WithTWLogger(l CtxLogger) TWOption {
	return func(tw *TimingWheel) {
		tw.logger = l
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

// TimingWheel a simple single timing wheel, and the accuracy is 1 second.
type TimingWheel struct {
	slot    int
	tick    time.Duration
	size    int
	bucket  []sync.Map
	stop    chan struct{}
	taskCtx func(ctx context.Context) context.Context
	logger  CtxLogger
}

// NewTimingWheel returns a new timing wheel.
func NewTimingWheel(tick time.Duration, size int, options ...TWOption) *TimingWheel {
	tw := &TimingWheel{
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

// AddTask adds task to timing wheel.
func (tw *TimingWheel) AddTask(ctx context.Context, taskID string, callback TWHandler, delay time.Duration) error {
	select {
	case <-tw.stop:
		return errors.New("TimingWheel has stoped")
	default:
	}

	task := &TWTask{
		ctx:      tw.taskCtx(ctx),
		addedAt:  time.Now(),
		callback: callback,
	}

	slot := tw.calcSlot(task, delay)

	if delay < tw.tick {
		go tw.run(taskID, task)

		return nil
	}

	tw.bucket[slot].Store(taskID, task)

	return nil
}

func (tw *TimingWheel) Stop() {
	select {
	case <-tw.stop:
		tw.logger.Warn(context.Background(), "TimingWheel has stoped")

		return
	default:
		close(tw.stop)
	}
}

func (tw *TimingWheel) calcSlot(task *TWTask, delay time.Duration) int {
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

func (tw *TimingWheel) scheduler() {
	ticker := time.NewTicker(tw.tick)
	defer ticker.Stop()

	for {
		select {
		case <-tw.stop:
			tw.logger.Info(context.Background(), fmt.Sprintf("TimingWheel stoped at: %s", time.Now().String()))

			return
		case <-ticker.C:
			tw.slot = (tw.slot + 1) % tw.size
			go tw.process(tw.slot)
		}
	}
}

func (tw *TimingWheel) process(slot int) {
	tw.bucket[slot].Range(func(key, value interface{}) bool {
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

func (tw *TimingWheel) run(taskID string, task *TWTask) {
	defer func() {
		if err := recover(); err != nil {
			tw.logger.Err(task.ctx, fmt.Sprintf("task(%s) run panic", taskID), zap.Any("error", err), zap.ByteString("stack", debug.Stack()))
		}
	}()

	if task.remainder > 0 {
		time.Sleep(task.remainder)
	}

	delay := time.Since(task.addedAt).String()

	if err := task.callback(task.ctx, taskID); err != nil {
		tw.logger.Err(task.ctx, fmt.Sprintf("task(%s) run error", taskID), zap.Error(err), zap.String("delay", delay))

		return
	}

	tw.logger.Info(task.ctx, fmt.Sprintf("task(%s) run ok", taskID), zap.String("delay", delay))
}
