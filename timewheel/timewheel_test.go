package timewheel

import (
	"context"
	"fmt"
	"math"
	"testing"
	"time"
)

// TestTimeWheel 测试时间轮
func TestTimeWheel(t *testing.T) {
	ctx := context.Background()

	ch := make(chan string)
	defer close(ch)

	tw := New(7, time.Second)

	addedAt := time.Now()

	tw.Go(ctx, func(ctx context.Context, taskId string, attempts int64) time.Duration {
		ch <- fmt.Sprintf("task[%s][%d] run after %ds", taskId, attempts, int64(math.Round(time.Since(addedAt).Seconds())))
		if attempts >= 10 {
			return 0
		}
		if attempts%2 == 0 {
			return time.Second * 2
		}
		return time.Second
	}, time.Second)

	tw.Go(ctx, func(ctx context.Context, taskId string, attempts int64) time.Duration {
		ch <- fmt.Sprintf("task[%s][%d] run after %ds", taskId, attempts, int64(math.Round(time.Since(addedAt).Seconds())))
		if attempts >= 5 {
			return 0
		}
		return time.Second * 2
	}, time.Second*2)

	for i := 0; i < 15; i++ {
		t.Log(<-ch)
	}
}

// TestPanic 测试Panic
func TestPanic(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	tw := New(7, time.Second, WithPanicFn(func(ctx context.Context, taskId string, err any, stack []byte) {
		fmt.Println("[task]", taskId)
		fmt.Println("[error]", err)
		fmt.Println("[stack]", string(stack))
		cancel()
	}))

	addedAt := time.Now()

	tw.Go(ctx, func(ctx context.Context, taskId string, attempts int64) time.Duration {
		fmt.Println("task run after", time.Since(addedAt).String())
		panic("oh no!")
	}, time.Second)

	<-ctx.Done()
}
