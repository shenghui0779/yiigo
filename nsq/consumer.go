package nsq

import (
	"time"

	"github.com/nsqio/go-nsq"
)

// Consumer nsq消费者接口
type Consumer interface {
	nsq.Handler
	// Topic 指定消费的Topic
	Topic() string
	// Channel 设置消费通道
	Channel() string
	// Attempts 设置重试次数
	Attempts() uint16
	// Config nsq相关配置
	Config() *nsq.Config
}

func consumerSet(lookupd []string, consumers ...Consumer) error {
	for _, c := range consumers {
		cfg := c.Config()
		if cfg == nil {
			cfg = nsq.NewConfig()
			cfg.LookupdPollInterval = time.Second
			cfg.RDYRedistributeInterval = time.Second
			cfg.MaxInFlight = 1000
		}
		// set attempt acount, default: 5
		if c.Attempts() > 0 {
			cfg.MaxAttempts = c.Attempts()
		}
		// add consumer
		nc, err := nsq.NewConsumer(c.Topic(), c.Channel(), cfg)
		if err != nil {
			return err
		}
		nc.SetLogger(&Logger{}, nsq.LogLevelError)
		nc.AddHandler(c)
		if err := nc.ConnectToNSQLookupds(lookupd); err != nil {
			return err
		}
	}
	return nil
}
