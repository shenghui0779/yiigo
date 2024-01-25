package yiigo

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDate(t *testing.T) {
	assert.Equal(t, "2016-03-19 15:03:19", TimeToStr(1458370999, time.DateTime))
}

func TestStrToTime(t *testing.T) {
	assert.Equal(t, int64(1562910319), StrToTime("2019-07-12 13:45:19", time.DateTime).Unix())
}

func TestWeekAround(t *testing.T) {
	monday, sunday := WeekAround(1562909685, time.DateOnly)

	assert.Equal(t, "2019-07-08", monday)
	assert.Equal(t, "2019-07-14", sunday)
}
