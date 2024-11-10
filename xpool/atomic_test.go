package xpool

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestAtomicInt64(t *testing.T) {
	i := NewAtomicInt64(1)
	assert.Equal(t, int64(1), i.Get())

	i.Set(2)
	assert.Equal(t, int64(2), i.Get())

	i.Add(1)
	assert.Equal(t, int64(3), i.Get())

	i.CompareAndSwap(3, 4)
	assert.Equal(t, int64(4), i.Get())

	i.CompareAndSwap(3, 5)
	assert.Equal(t, int64(4), i.Get())
}

func TestAtomicDuration(t *testing.T) {
	d := NewAtomicDuration(time.Second)
	assert.Equal(t, time.Second, d.Get())

	d.Set(time.Second * 2)
	assert.Equal(t, time.Second*2, d.Get())

	d.Add(time.Second)
	assert.Equal(t, time.Second*3, d.Get())

	d.CompareAndSwap(time.Second*3, time.Second*4)
	assert.Equal(t, time.Second*4, d.Get())

	d.CompareAndSwap(time.Second*3, time.Second*5)
	assert.Equal(t, time.Second*4, d.Get())
}
