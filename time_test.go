package yiigo

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDate(t *testing.T) {
	tz, err := time.LoadLocation("Asia/Shanghai")
	assert.NoError(t, err)
	assert.Equal(t, "2016-03-19 15:03:19", TimeToStr(time.DateTime, 1458370999, tz))
}

func TestStrToTime(t *testing.T) {
	tz, err := time.LoadLocation("Asia/Shanghai")
	assert.NoError(t, err)
	assert.Equal(t, int64(1562910319), StrToTime(time.DateTime, "2019-07-12 13:45:19", tz).Unix())
}

func TestWeekAround(t *testing.T) {
	tz, err := time.LoadLocation("Asia/Shanghai")
	assert.NoError(t, err)
	now := time.Unix(1562909685, 0).In(tz)
	monday, sunday := WeekAround(time.DateOnly, now)
	assert.Equal(t, "2019-07-08", monday)
	assert.Equal(t, "2019-07-14", sunday)
}
