package yiigo

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNSQOption(t *testing.T) {
	setting := new(nsqSetting)

	options := []NSQOption{
		WithLookupdPollInterval(time.Second),
		WithRDYRedistributeInterval(time.Second),
		WithMaxInFlight(1000),
	}

	for _, f := range options {
		f(setting)
	}

	assert.Equal(t, &nsqSetting{
		lookupdPollInterval:     time.Second,
		rdyRedistributeInterval: time.Second,
		maxInFlight:             1000,
	}, setting)
}
