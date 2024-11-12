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
	ch := make(chan string)
	defer close(ch)

	tw := New(7, time.Second)
	for i := 0; i < 10; i++ {
		n := i + 1
		addedAt := time.Now()
		go func() {
			tw.Go(context.Background(), func(ctx context.Context, taskId string) error {
				ch <- fmt.Sprintf("%s[%d] run after %ds", taskId, n, int64(math.Round(time.Since(addedAt).Seconds())))
				return nil
			}, WithDelay(func(attempts uint16) time.Duration {
				return time.Second * time.Duration(n)
			}))
		}()
	}

	for i := 0; i < 10; i++ {
		t.Log(<-ch)
	}
}
