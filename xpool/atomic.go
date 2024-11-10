package xpool

import (
	"sync/atomic"
	"time"
)

// AtomicInt64 is a wrapper with a simpler interface around atomic.(Add|Store|Load|CompareAndSwap)Int64 functions.
type AtomicInt64 struct {
	int64
}

// NewAtomicInt64 initializes a new AtomicInt64 with a given value.
func NewAtomicInt64(n int64) AtomicInt64 {
	return AtomicInt64{n}
}

// Add atomically adds n to the value.
func (i *AtomicInt64) Add(n int64) int64 {
	return atomic.AddInt64(&i.int64, n)
}

// Set atomically sets n as new value.
func (i *AtomicInt64) Set(n int64) {
	atomic.StoreInt64(&i.int64, n)
}

// Get atomically returns the current value.
func (i *AtomicInt64) Get() int64 {
	return atomic.LoadInt64(&i.int64)
}

// CompareAndSwap automatically swaps the old with the new value.
func (i *AtomicInt64) CompareAndSwap(oldval, newval int64) bool {
	return atomic.CompareAndSwapInt64(&i.int64, oldval, newval)
}

// AtomicDuration is a wrapper with a simpler interface around atomic.(Add|Store|Load|CompareAndSwap)Int64 functions.
type AtomicDuration struct {
	int64
}

// NewAtomicDuration initializes a new AtomicDuration with a given value.
func NewAtomicDuration(duration time.Duration) AtomicDuration {
	return AtomicDuration{int64(duration)}
}

// Add atomically adds duration to the value.
func (d *AtomicDuration) Add(duration time.Duration) time.Duration {
	return time.Duration(atomic.AddInt64(&d.int64, int64(duration)))
}

// Set atomically sets duration as new value.
func (d *AtomicDuration) Set(duration time.Duration) {
	atomic.StoreInt64(&d.int64, int64(duration))
}

// Get atomically returns the current value.
func (d *AtomicDuration) Get() time.Duration {
	return time.Duration(atomic.LoadInt64(&d.int64))
}

// CompareAndSwap automatically swaps the old with the new value.
func (d *AtomicDuration) CompareAndSwap(oldval, newval time.Duration) bool {
	return atomic.CompareAndSwapInt64(&d.int64, int64(oldval), int64(newval))
}
