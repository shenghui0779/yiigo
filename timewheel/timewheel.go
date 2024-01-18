package timewheel

import (
	"context"
	"sync"
	"time"
)

type ctxKey int

// ctxTaskAddedAt Context存储任务入队时间的Key
const ctxTaskAddedAt ctxKey = 0

// TaskAddedAt 返回任务的添加时间
func TaskAddedAt(ctx context.Context) time.Time {
	v := ctx.Value(ctxTaskAddedAt)
	if v == nil {
		return time.Time{}
	}

	t, ok := v.(time.Time)
	if !ok {
		return time.Time{}
	}

	return t
}

// Task 时间轮任务
type Task struct {
	ctx         context.Context
	uniqID      string                                         // 任务唯一标识
	round       int                                            // 延迟执行的轮数
	attempts    uint16                                         // 当前尝试的次数
	maxAttempts uint16                                         // 最大尝试次数
	remainder   time.Duration                                  // 任务执行前的剩余延迟（小于时间轮精度）
	deferFn     func(attempts uint16) time.Duration            // 返回任务下一次延迟执行的时间
	callback    func(ctx context.Context, taskID string) error // 任务回调函数
}

// TimeWheel 单时间轮
type TimeWheel interface {
	// AddTask 添加一个任务，到期被执行，默认仅执行一次；若指定了重试次数，则在发生错误后重试；
	// 注意：任务是异步执行的，`ctx`一旦被取消，则任务也随之取消，故需考虑是否应该克隆一个不带「取消」的`ctx`
	AddTask(ctx context.Context, taskID string, handler func(ctx context.Context, taskID string) error, options ...Option)
	// Run 运行时间轮
	Run()
	// Stop 终止时间轮
	Stop()
}

type timewheel struct {
	slot   int
	tick   time.Duration
	size   int
	bucket []sync.Map
	stop   chan struct{}
}

func (tw *timewheel) AddTask(ctx context.Context, taskID string, handler func(ctx context.Context, taskID string) error, options ...Option) {
	task := &Task{
		ctx:         ctx,
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

func (tw *timewheel) Run() {
	go tw.scheduler()
}

func (tw *timewheel) Stop() {
	select {
	case <-tw.stop: // 时间轮已停止
		return
	default:
	}

	close(tw.stop)
}

func (tw *timewheel) requeue(task *Task) {
	select {
	case <-tw.stop: // 时间轮已停止
		return
	default:
	}

	// 任务已达到最大尝试次数
	if task.attempts >= task.maxAttempts {
		return
	}

	task.attempts++
	task.ctx = context.WithValue(task.ctx, ctxTaskAddedAt, time.Now())

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

	tw.bucket[slot].Store(task.uniqID, task)
}

func (tw *timewheel) scheduler() {
	ticker := time.NewTicker(tw.tick)
	defer ticker.Stop()

	for {
		select {
		case <-tw.stop: // 时间轮已停止
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
		case <-tw.stop: // 时间轮已停止
			return false
		default:
		}

		task := value.(*Task)
		if task.round > 0 {
			task.round--
			return true
		}

		select {
		case <-task.ctx.Done(): // 任务被取消
		default:
			go tw.do(task)
		}

		tw.bucket[slot].Delete(key)

		return true
	})
}

func (tw *timewheel) do(task *Task) {
	defer func() {
		if recover() != nil {
			tw.requeue(task)
		}
	}()

	if task.remainder > 0 {
		time.Sleep(task.remainder)
	}

	if err := task.callback(task.ctx, task.uniqID); err != nil {
		tw.requeue(task)
	}
}

// New 返回一个时间轮实例
func New(tick time.Duration, size int) TimeWheel {
	return &timewheel{
		tick:   tick,
		size:   size,
		bucket: make([]sync.Map, size),
		stop:   make(chan struct{}),
	}
}
