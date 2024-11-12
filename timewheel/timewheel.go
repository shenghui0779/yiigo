package timewheel

import (
	"context"
	"fmt"
	"runtime/debug"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

// Task 时间轮任务
type Task struct {
	ctx         context.Context
	id          string                                         // 任务ID
	attempts    uint16                                         // 当前尝试的次数
	maxAttempts uint16                                         // 最大尝试次数
	round       int                                            // 延迟执行的轮数
	remainder   time.Duration                                  // 任务执行前的剩余延迟（小于时间轮精度）
	deferFn     func(attempts uint16) time.Duration            // 返回任务下一次延迟执行的时间
	callback    func(ctx context.Context, taskID string) error // 任务回调函数
}

// TimeWheel 单时间轮
type TimeWheel interface {
	// Go 异步一个任务并返回任务ID，到期被执行，默认仅执行一次；若指定了重试次数，则在返回`error`后重试；
	// 注意：任务是异步执行的，`ctx`一旦被取消，则任务也随之取消；如要保证任务不被取消，请使用`context.WithoutCancel`
	Go(ctx context.Context, fn func(ctx context.Context, taskId string) error, options ...Option) string
	// Stop 终止时间轮
	Stop()
	// Err 监听异常错误
	Err() <-chan error
}

type timewheel struct {
	slot   int
	size   int
	tick   time.Duration
	bucket []sync.Map

	ctx    context.Context
	cancel context.CancelFunc

	err chan error
}

func (tw *timewheel) Go(ctx context.Context, fn func(ctx context.Context, taskId string) error, options ...Option) string {
	taskId := strings.ReplaceAll(uuid.New().String(), "-", "")
	task := &Task{
		ctx:         ctx,
		id:          taskId,
		callback:    fn,
		maxAttempts: 1,
		deferFn: func(attempts uint16) time.Duration {
			return 0
		},
	}
	for _, f := range options {
		f(task)
	}
	tw.requeue(task)
	return taskId
}

func (tw *timewheel) Err() <-chan error {
	return tw.err
}

func (tw *timewheel) Stop() {
	select {
	case <-tw.ctx.Done(): // 时间轮已停止
		return
	default:
	}

	tw.cancel()
	close(tw.err)
}

func (tw *timewheel) requeue(task *Task) {
	select {
	case <-tw.ctx.Done(): // 时间轮已停止
		return
	case <-task.ctx.Done(): // 任务被取消
		return
	default:
	}

	// 任务已达到最大尝试次数
	if task.attempts >= task.maxAttempts {
		return
	}
	task.attempts++

	tick := tw.tick.Nanoseconds()
	delay := task.deferFn(task.attempts)
	duration := delay.Nanoseconds()
	// 圈数
	task.round = int(duration / (tick * int64(tw.size)))
	// 槽位
	slot := (int(duration/tick)%tw.size + tw.slot) % tw.size
	if slot == tw.slot {
		if task.round == 0 {
			task.remainder = delay
			go tw.do(task)
			return
		}
		task.round--
	}
	// 剩余延迟
	task.remainder = time.Duration(duration % tick)
	// 存储任务
	tw.bucket[slot].Store(task.id, task)
}

func (tw *timewheel) scheduler() {
	ticker := time.NewTicker(tw.tick)
	defer ticker.Stop()

	for {
		select {
		case <-tw.ctx.Done(): // 时间轮已停止
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
		case <-tw.ctx.Done(): // 时间轮已停止
			return false
		default:
		}

		task, ok := value.(*Task)
		if !ok {
			return true
		}
		if task.round > 0 {
			task.round--
			return true
		}
		tw.do(task)
		tw.bucket[slot].Delete(key)
		return true
	})
}

func (tw *timewheel) do(task *Task) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				err := fmt.Errorf("task(%s) panic recovered: %+v\n%s", task.id, r, string(debug.Stack()))
				select {
				case tw.err <- err:
				default:
				}
			}
		}()

		if task.remainder > 0 {
			time.Sleep(task.remainder)
		}

		select {
		case <-tw.ctx.Done():
			return
		case <-task.ctx.Done(): // 任务被取消
			err := fmt.Errorf("task(%s) canceled: %w", task.id, context.Cause(task.ctx))
			select {
			case tw.err <- err:
			default:
			}
			return
		default:
		}

		if err := task.callback(task.ctx, task.id); err != nil {
			tw.requeue(task)
		}
	}()
}

// New 返回一个时间轮实例
func New(size int, tick time.Duration) TimeWheel {
	ctx, cancel := context.WithCancel(context.Background())

	tw := &timewheel{
		size:   size,
		tick:   tick,
		bucket: make([]sync.Map, size),

		ctx:    ctx,
		cancel: cancel,

		err: make(chan error),
	}

	go tw.scheduler()

	return tw
}
