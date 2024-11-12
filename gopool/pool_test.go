package gopool

import (
	"context"
	"math"
	"sync"
	"testing"
	"time"
)

func TestNormal(t *testing.T) {
	ctx := context.Background()
	p := New(2)
	defer p.Close()
	m := make(map[int]int)
	for i := 0; i < 4; i++ {
		m[i] = i
	}
	var wg sync.WaitGroup
	wg.Add(1)
	p.Go(ctx, func(context.Context) {
		m[1]++
		wg.Done()
	})
	wg.Add(1)
	p.Go(ctx, func(context.Context) {
		m[2]++
		wg.Done()
	})
	wg.Wait()
	Close()
	t.Log(m)
}

func sleep1s(context.Context) {
	time.Sleep(time.Second)
}

func TestLimit(t *testing.T) {
	ctx := context.Background()
	var wg sync.WaitGroup
	// 没有并发数限制
	now := time.Now()
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			sleep1s(ctx)
			wg.Done()
		}()
	}
	wg.Wait()
	sec := math.Round(time.Since(now).Seconds())
	if sec != 1 {
		t.FailNow()
	}
	// 限制并发数
	p := New(2)
	defer p.Close()
	now = time.Now()
	for i := 0; i < 4; i++ {
		wg.Add(1)
		p.Go(ctx, func(ctx context.Context) {
			sleep1s(ctx)
			wg.Done()
		})
	}
	wg.Wait()
	sec = math.Round(time.Since(now).Seconds())
	if sec != 2 {
		t.FailNow()
	}
}

func TestRecover(t *testing.T) {
	ch := make(chan struct{})
	defer close(ch)
	p := New(2, WithPanicHandler(func(ctx context.Context, err interface{}, stack []byte) {
		t.Log("[error] job panic:", err)
		t.Log("[stack]", string(stack))
		ch <- struct{}{}
	}))
	defer p.Close()
	p.Go(context.Background(), func(ctx context.Context) {
		sleep1s(ctx)
		panic("oh my god!")
	})
	<-ch
}
