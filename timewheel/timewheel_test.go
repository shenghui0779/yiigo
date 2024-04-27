package timewheel

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"testing"
	"time"
)

// TestTimeWheel_1 测试时间轮 - 先添加任务再运行
func TestTimeWheel_1(t *testing.T) {
	ch := make(chan string)
	defer close(ch)

	tw := New(time.Second, 7)
	for i := 0; i < 10; i++ {
		n := i + 1
		addedAt := time.Now()
		tw.AddTask(context.Background(), "task#"+strconv.Itoa(n), func(ctx context.Context, taskID string) error {
			ch <- fmt.Sprintf("%s run after %ds", taskID, int64(math.Round(time.Since(addedAt).Seconds())))
			return nil
		}, WithDelay(func(attempts uint16) time.Duration {
			return time.Second * time.Duration(n+i)
		}))
	}
	tw.Run()

	for i := 0; i < 10; i++ {
		t.Log(<-ch)
	}
}

// TestTimeWheel_2 测试时间轮 - 先运行再添加任务
func TestTimeWheel_2(t *testing.T) {
	ch := make(chan string)
	defer close(ch)

	tw := New(time.Second, 7)
	tw.Run()

	for i := 0; i < 10; i++ {
		time.Sleep(time.Second)

		n := i + 1
		addedAt := time.Now()
		tw.AddTask(context.Background(), "task#"+strconv.Itoa(n), func(ctx context.Context, taskID string) error {
			ch <- fmt.Sprintf("%s run after %ds", taskID, int64(math.Round(time.Since(addedAt).Seconds())))
			return nil
		}, WithDelay(func(attempts uint16) time.Duration {
			return time.Second * time.Duration(n+i)
		}))
	}

	for i := 0; i < 10; i++ {
		t.Log(<-ch)
	}
}
