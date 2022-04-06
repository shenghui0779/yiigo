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
	ctx      context.Context
	round    int
	addedAt  time.Time
	callback TWHandler
}

// TWOption timing wheel option.
type TWOption func(tw *TimingWheel)

// WithTWTaskCtx clones ctx for execute tasks asynchronously.
func WithTWTaskCtx(fn func(ctx context.Context) context.Context) TWOption {
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
	buckets []sync.Map
	stop    chan struct{}
	taskCtx func(ctx context.Context) context.Context
	logger  CtxLogger
}

// NewTimingWheel returns a new timing wheel.
func NewTimingWheel(size int, options ...TWOption) (*TimingWheel, error) {
	tw := &TimingWheel{
		tick:    time.Second,
		size:    size,
		buckets: make([]sync.Map, size),
		stop:    make(chan struct{}),
		taskCtx: func(ctx context.Context) context.Context {
			return context.Background()
		},
		logger: new(twLogger),
	}

	for _, f := range options {
		f(tw)
	}

	go tw.scheduler()

	return tw, nil
}

// AddTask adds task to timing wheel.
func (tw *TimingWheel) AddTask(ctx context.Context, taskID string, callback TWHandler, delay time.Duration) error {
	select {
	case <-tw.stop:
		return errors.New("TimingWheel has stoped")
	default:
	}

	if delay < tw.tick {
		tw.logger.Warn(ctx, "task run immediately because of delay less than one second", zap.String("task_id", taskID), zap.String("delay", delay.String()))

		if err := callback(ctx, taskID); err != nil {
			tw.logger.Err(ctx, "task run error", zap.String("task_id", taskID), zap.Error(err))

			return nil
		}

		tw.logger.Info(ctx, "task run ok", zap.String("task_id", taskID), zap.String("delay", delay.String()))

		return nil
	}

	task := &TWTask{
		ctx:      tw.taskCtx(ctx),
		addedAt:  time.Now(),
		callback: callback,
	}

	slot := tw.calcSlot(task, delay)

	tw.buckets[slot].Store(taskID, task)

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
	tick := int(tw.tick.Seconds())
	total := tick * tw.size
	duration := int(delay.Seconds())

	if duration > total {
		task.round = duration / total
		duration = duration % total

		if duration == 0 {
			task.round--
		}
	}

	return (tw.slot + duration/tick) % tw.size
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
			go tw.run(tw.slot)
		}
	}
}

func (tw *TimingWheel) run(slot int) {
	tw.buckets[slot].Range(func(key, value interface{}) bool {
		taskID := key.(string)
		task := value.(*TWTask)

		if task.round > 0 {
			task.round--

			return true
		}

		go func() {
			defer func() {
				if err := recover(); err != nil {
					tw.logger.Err(task.ctx, "task run panic",
						zap.Any("error", err),
						zap.ByteString("stack", debug.Stack()),
					)
				}
			}()

			if err := task.callback(task.ctx, taskID); err != nil {
				tw.logger.Err(task.ctx, "task run error", zap.String("task_id", taskID), zap.Error(err), zap.String("delay", time.Since(task.addedAt).String()))

				return
			}

			tw.logger.Info(task.ctx, "task run ok", zap.String("task_id", taskID), zap.String("delay", time.Since(task.addedAt).String()))
		}()

		tw.buckets[slot].Delete(key)

		return true
	})
}
