package worker

import (
	"context"
	"math"
	"testing"
	"time"
)

func TestNormal(t *testing.T) {
	ctx := context.Background()
	m := make(map[int]int)
	for i := 0; i < 4; i++ {
		m[i] = i
	}
	ch := make(chan struct{}, 2)
	defer close(ch)
	Go(ctx, func(context.Context) {
		m[1]++
		ch <- struct{}{}
		return
	})
	Go(ctx, func(context.Context) {
		m[2]++
		ch <- struct{}{}
		return
	})
	count := 0
	for count < 2 {
		<-ch
		count++
	}
	Close()
	t.Log(m)
}

func sleep1s(context.Context) {
	time.Sleep(time.Second)
}

func TestLimit(t *testing.T) {
	ctx := context.Background()
	ch := make(chan struct{}, 4)
	defer close(ch)
	// 没有并发数限制
	now := time.Now()
	for i := 0; i < 4; i++ {
		go func() {
			sleep1s(ctx)
			ch <- struct{}{}
		}()
	}
	count := 0
	for count < 4 {
		<-ch
		count++
	}
	sec := math.Round(time.Since(now).Seconds())
	if sec != 1 {
		t.FailNow()
	}
	// 限制并发数
	w := New(2, nil)
	defer w.Close()
	now = time.Now()
	for i := 0; i < 4; i++ {
		w.Go(ctx, func(ctx context.Context) {
			sleep1s(ctx)
			ch <- struct{}{}
		})
	}
	count = 0
	for count < 4 {
		<-ch
		count++
	}
	sec = math.Round(time.Since(now).Seconds())
	if sec != 2 {
		t.FailNow()
	}
}

func TestRecover(t *testing.T) {
	ch := make(chan struct{})
	defer close(ch)
	w := New(2, func(ctx context.Context, err interface{}, stack []byte) {
		t.Log("[error] job panic:", err)
		t.Log("[stack]", string(stack))
		ch <- struct{}{}
	})
	defer w.Close()
	w.Go(context.Background(), func(ctx context.Context) {
		sleep1s(ctx)
		panic("oh my god!")
	})
	<-ch
}
