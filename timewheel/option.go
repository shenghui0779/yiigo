package timewheel

// Option 时间轮选项
type Option func(tw *timewheel)

func WithCtxErrFn(fn CtxErrFn) Option {
	return func(tw *timewheel) {
		tw.ctxErrFn = fn
	}
}

// WithPanicFn 指定任务执行Panic的处理方法
func WithPanicFn(fn PanicFn) Option {
	return func(tw *timewheel) {
		tw.panicFn = fn
	}
}
