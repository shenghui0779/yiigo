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

	tw := NewTimeWheel(time.Second, 7)

	for i := 0; i < 10; i++ {
		n := i + 1

		tw.AddTask(context.Background(), "task#"+strconv.Itoa(n), func(ctx context.Context, taskID string) error {
			ch <- fmt.Sprintf("%s run after %ds", taskID, int64(time.Since(TaskAddedAt(ctx)).Seconds()))
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
		"task#1 run after 1s",
		"task#2 run after 3s",
		"task#3 run after 5s",
		"task#4 run after 7s",
		"task#5 run after 9s",
		"task#6 run after 11s",
		"task#7 run after 13s",
		"task#8 run after 15s",
		"task#9 run after 17s",
		"task#10 run after 19s",
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

		tw.AddTask(context.Background(), "task#"+strconv.Itoa(n), func(ctx context.Context, taskID string) error {
			ch <- fmt.Sprintf("%s run after %ds", taskID, int64(time.Since(TaskAddedAt(ctx)).Seconds()))
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
		"task#1 run after 1s",
		"task#2 run after 3s",
		"task#3 run after 5s",
		"task#4 run after 7s",
		"task#5 run after 9s",
		"task#6 run after 11s",
		"task#7 run after 13s",
		"task#8 run after 15s",
		"task#9 run after 17s",
		"task#10 run after 19s",
	}, ret)
}
