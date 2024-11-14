package timewheel

import (
	"context"
	"fmt"
	"runtime/debug"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/shenghui0779/yiigo/linklist"
)

type task struct {
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
	bucket []*linklist.DoublyLinkList[*task]

	ctx    context.Context
	cancel context.CancelFunc

	err chan error
}

func (tw *timewheel) Go(ctx context.Context, fn func(ctx context.Context, taskId string) error, options ...Option) string {
	id := strings.ReplaceAll(uuid.New().String(), "-", "")
	t := &task{
		ctx:         ctx,
		id:          id,
		callback:    fn,
		maxAttempts: 1,
		deferFn: func(attempts uint16) time.Duration {
			return 0
		},
	}
	for _, f := range options {
		f(t)
	}
	tw.requeue(t)
	return id
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

func (tw *timewheel) requeue(t *task) {
	select {
	case <-tw.ctx.Done(): // 时间轮已停止
		return
	case <-t.ctx.Done(): // 任务被取消
		return
	default:
	}

	// 任务已达到最大尝试次数
	if t.attempts >= t.maxAttempts {
		return
	}
	t.attempts++

	tick := tw.tick.Nanoseconds()
	delay := t.deferFn(t.attempts)
	duration := delay.Nanoseconds()
	// 圈数
	t.round = int(duration / (tick * int64(tw.size)))
	// 槽位
	slot := (int(duration/tick)%tw.size + tw.slot) % tw.size
	if slot == tw.slot {
		if t.round == 0 {
			t.remainder = delay
			go tw.do(t)
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
	go func() {
		defer func() {
			if r := recover(); r != nil {
				err := fmt.Errorf("t(%s) panic recovered: %+v\n%s", t.id, r, string(debug.Stack()))
				select {
				case tw.err <- err:
				default:
				}
			}
		}()

		if t.remainder > 0 {
			time.Sleep(t.remainder)
		}

		select {
		case <-tw.ctx.Done():
			return
		case <-t.ctx.Done(): // 任务被取消
			err := fmt.Errorf("t(%s) canceled: %w", t.id, context.Cause(t.ctx))
			select {
			case tw.err <- err:
			default:
			}
			return
		default:
		}

		if err := t.callback(t.ctx, t.id); err != nil {
			tw.requeue(t)
		}
	}()
}

// New 返回一个时间轮实例
func New(size int, tick time.Duration) TimeWheel {
	ctx, cancel := context.WithCancel(context.TODO())
	tw := &timewheel{
		size:   size,
		tick:   tick,
		bucket: make([]*linklist.DoublyLinkList[*task], size),

		ctx:    ctx,
		cancel: cancel,

		err: make(chan error),
	}
	for i := 0; i < size; i++ {
		tw.bucket[i] = linklist.New[*task]()
	}

	go tw.scheduler()

	return tw
}
