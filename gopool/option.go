package gopool

import "time"

// 协程池选项
type Option func(*Pool)

// WithIdleTimeout 协程闲置超时时长，默认：60s
func WithIdleTimeout(duration time.Duration) Option {
	return func(p *Pool) {
		if duration > 0 {
			p.idleTimeout = duration
		}
	}
}

// WithPrefill 预填充协程数量
func WithPrefill(n int) Option {
	return func(p *Pool) {
		if n > 0 {
			p.prefill = n
		}
	}
}

// WithNonBlock 非阻塞模式，任务会被缓存到本地链表
func WithNonBlock() Option {
	return func(p *Pool) {
		p.nonBlock = true
	}
}

// WithQueueCap 任务队列缓冲容量，默认无缓冲
func WithQueueCap(cap int) Option {
	return func(p *Pool) {
		if cap > 0 {
			p.queueCap = cap
		}
	}
}

// WithPanicHandler 任务Panic处理方法
func WithPanicHandler(fn PanicFn) Option {
	return func(p *Pool) {
		p.panicFn = fn
	}
}
