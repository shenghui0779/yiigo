package timewheel

import (
	"context"
	"runtime/debug"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/shenghui0779/yiigo/linklist"
)

type (
	// TaskFn 任务方法，返回下一次执行的延迟时间，若返回0，则表示不再执行
	TaskFn func(ctx context.Context, taskId string, attempts int64) time.Duration

	// CtxErrFn 任务 context「取消/超时」的处理方法
	CtxErrFn func(ctx context.Context, taskId string, err error)

	// PanicFn 任务发生Panic的处理方法
	PanicFn func(ctx context.Context, taskId string, err any, stack []byte)
)

// TimeWheel 单层时间轮
type TimeWheel interface {
	// Go 异步一个任务并返回任务ID；
	// 注意：任务是异步执行的，`ctx`一旦被取消/超时，则任务也随之取消；
	// 如要保证任务不被取消，请使用`context.WithoutCancel`
	Go(ctx context.Context, taskFn TaskFn, delay time.Duration) string

	// Stop 终止时间轮
	Stop()
}

type task struct {
	id string // 任务ID

	callback TaskFn // 任务执行函数
	attempts int64  // 当前任务执行的次数

	round     int           // 延迟执行的轮数
	remainder time.Duration // 任务执行前的剩余延迟（小于时间轮精度）

	ctx context.Context
}

type timewheel struct {
	slot   int
	size   int
	tick   time.Duration
	bucket []*linklist.DoublyLinkList[*task]

	ctx    context.Context
	cancel context.CancelFunc

	ctxErrFn CtxErrFn // Ctx Done 处理函数
	panicFn  PanicFn  // Panic处理函数
}

func (tw *timewheel) Go(ctx context.Context, taskFn TaskFn, delay time.Duration) string {
	id := strings.ReplaceAll(uuid.New().String(), "-", "")
	t := &task{
		id:       id,
		callback: taskFn,
		ctx:      ctx,
	}
	tw.requeue(t, delay)
	return id
}

func (tw *timewheel) Stop() {
	select {
	case <-tw.ctx.Done(): // 时间轮已停止
		return
	default:
	}
	tw.cancel()
}

func (tw *timewheel) requeue(t *task, delay time.Duration) {
	select {
	case <-tw.ctx.Done(): // 时间轮已停止
		return
	default:
	}

	t.attempts++

	tick := tw.tick.Nanoseconds()
	duration := delay.Nanoseconds()
	// 圈数
	t.round = int(duration / (tick * int64(tw.size)))
	// 槽位
	slot := (int(duration/tick)%tw.size + tw.slot) % tw.size
	if slot == tw.slot {
		if t.round == 0 {
			t.remainder = delay
			tw.do(t)
			return
		}
		t.round--
	}
	// 剩余延迟
	t.remainder = time.Duration(duration % tick)
	// 存储任务
	tw.bucket[slot].Append(t)
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
			tw.process(tw.slot)
		}
	}
}

func (tw *timewheel) process(slot int) {
	tasks := tw.bucket[slot].Filter(func(index int, value *task) bool {
		if value.round > 0 {
			value.round--
			return false
		}
		return true
	})
	for _, t := range tasks {
		tw.do(t)
	}
}

func (tw *timewheel) do(t *task) {
	go func(t *task) {
		defer func() {
			if rerr := recover(); rerr != nil {
				if tw.panicFn != nil {
					tw.panicFn(t.ctx, t.id, rerr, debug.Stack())
				}
			}
		}()

		if t.remainder > 0 {
			time.Sleep(t.remainder)
		}

		select {
		case <-tw.ctx.Done(): // 时间轮停止
			return
		case <-t.ctx.Done(): // 任务被取消
			if tw.ctxErrFn != nil {
				tw.ctxErrFn(t.ctx, t.id, t.ctx.Err())
			}
			return
		default:
		}

		delay := t.callback(t.ctx, t.id, t.attempts)
		if delay > 0 {
			tw.requeue(t, delay)
		}
	}(t)
}

// New 返回一个时间轮实例
func New(size int, tick time.Duration, opts ...Option) TimeWheel {
	ctx, cancel := context.WithCancel(context.TODO())
	tw := &timewheel{
		size:   size,
		tick:   tick,
		bucket: make([]*linklist.DoublyLinkList[*task], size),

		ctx:    ctx,
		cancel: cancel,
	}
	for _, fn := range opts {
		fn(tw)
	}
	for i := 0; i < size; i++ {
		tw.bucket[i] = linklist.New[*task]()
	}

	go tw.scheduler()

	return tw
}
