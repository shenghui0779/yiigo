package timewheel

import "time"

// Option 时间轮任务选项
type Option func(t *Task)

// WithAttempts 指定任务重试次数；默认：1
func WithAttempts(attempts uint16) Option {
	return func(t *Task) {
		if attempts > 0 {
			t.maxAttempts = attempts
		}
	}
}

// WithDelay 指定任务延迟执行时间；默认：立即执行
func WithDelay(fn func(attempts uint16) time.Duration) Option {
	return func(t *Task) {
		if fn != nil {
			t.deferFn = fn
		}
	}
}
