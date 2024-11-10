package xpool

import (
	"testing"
	"time"
)

func TestSemaNoTimeout(t *testing.T) {
	s := NewSemaphore(1, 0)
	s.Acquire()
	released := false
	go func() {
		time.Sleep(10 * time.Millisecond)
		released = true
		s.Release()
	}()
	s.Acquire()
	if !released {
		t.Errorf("release: false, want true")
	}
}

func TestSemaTimeout(t *testing.T) {
	s := NewSemaphore(1, 5*time.Millisecond)
	s.Acquire()
	go func() {
		time.Sleep(10 * time.Millisecond)
		s.Release()
	}()
	if s.Acquire() {
		t.Errorf("Acquire: true, want false")
	}
	time.Sleep(10 * time.Millisecond)
	if !s.Acquire() {
		t.Errorf("Acquire: false, want true")
	}
}

func TestSemaTryAcquire(t *testing.T) {
	s := NewSemaphore(1, 0)
	if !s.TryAcquire() {
		t.Errorf("TryAcquire: false, want true")
	}
	if s.TryAcquire() {
		t.Errorf("TryAcquire: true, want false")
	}
	s.Release()
	if !s.TryAcquire() {
		t.Errorf("TryAcquire: false, want true")
	}
}
