package yiigo

import (
	"context"
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestTimeWheel_1 测试时间轮 - 先添加任务再运行
func TestTimeWheel_1(t *testing.T) {
	ch := make(chan string)
	defer close(ch)

	tw := NewTimeWheel(time.Second, 60)

	for i := 0; i < 10; i++ {
		n := i + 1
		now := time.Now()

		tw.AddTask(context.Background(), "task#"+strconv.Itoa(n), func(ctx context.Context, taskID string) error {
			ch <- fmt.Sprintf("%s - %ds", taskID, int64(time.Since(now).Seconds()))

			return nil
		}, WithTaskDefer(func(attempts uint16) time.Duration {
			return time.Second * time.Duration(n+i)
		}))
	}

	tw.Run()

	ret := make([]string, 0, 10)

	for v := range ch {
		ret = append(ret, v)
		if len(ret) == 10 {
			break
		}
	}

	assert.Equal(t, []string{
		"task#1 - 1s",
		"task#2 - 3s",
		"task#3 - 5s",
		"task#4 - 7s",
		"task#5 - 9s",
		"task#6 - 11s",
		"task#7 - 13s",
		"task#8 - 15s",
		"task#9 - 17s",
		"task#10 - 19s",
	}, ret)
}

// TestTimeWheel_2 测试时间轮 - 先运行再添加任务
func TestTimeWheel_2(t *testing.T) {
	ch := make(chan string)
	defer close(ch)

	tw := NewTimeWheel(time.Second, 60)
	tw.Run()

	for i := 0; i < 10; i++ {
		n := i + 1
		now := time.Now()

		tw.AddTask(context.Background(), "task#"+strconv.Itoa(n), func(ctx context.Context, taskID string) error {
			ch <- fmt.Sprintf("%s - %ds", taskID, int64(time.Since(now).Seconds()))

			return nil
		}, WithTaskDefer(func(attempts uint16) time.Duration {
			return time.Second * time.Duration(n+i)
		}))
	}

	ret := make([]string, 0, 10)

	for v := range ch {
		ret = append(ret, v)
		if len(ret) == 10 {
			break
		}
	}

	assert.Equal(t, []string{
		"task#1 - 1s",
		"task#2 - 3s",
		"task#3 - 5s",
		"task#4 - 7s",
		"task#5 - 9s",
		"task#6 - 11s",
		"task#7 - 13s",
		"task#8 - 15s",
		"task#9 - 17s",
		"task#10 - 19s",
	}, ret)
}
