package yiigo

import (
	"errors"
	"fmt"
	"time"

	"github.com/nsqio/go-nsq"
	"go.uber.org/zap"
)

var producer *nsq.Producer

// NSQLogger 实现nsq日志接口
type NSQLogger struct{}

// Output nsq错误输出
func (l *NSQLogger) Output(calldepth int, s string) error {
	logger.Error(fmt.Sprintf("err nsq: %s", s), zap.Int("call_depth", calldepth))

	return nil
}

func initNSQProducer(nsqd string, cfg *nsq.Config) error {
	if cfg == nil {
		cfg = nsq.NewConfig()
	}

	var err error

	producer, err = nsq.NewProducer(nsqd, cfg)
	if err != nil {
		return err
	}

	if err = producer.Ping(); err != nil {
		return err
	}

	producer.SetLogger(&NSQLogger{}, nsq.LogLevelError)

	return nil
}

// NSQPublish 同步推送消息到指定Topic
func NSQPublish(topic string, msg []byte) error {
	if producer == nil {
		return errors.New("nsq producer is nil (forgotten configure?)")
	}

	return producer.Publish(topic, msg)
}

// NSQDeferredPublish 同步推送延迟消息到指定Topic
func NSQDeferredPublish(topic string, msg []byte, duration time.Duration) error {
	if producer == nil {
		return errors.New("nsq producer is nil (forgotten configure?)")
	}

	return producer.DeferredPublish(topic, duration, msg)
}

// NSQConsumer nsq消费者接口
type NSQConsumer interface {
	nsq.Handler

	// Topic 指定消费的Topic
	Topic() string

	// Channel 设置消费通道
	Channel() string

	// Attempts 设置重试次数
	Attempts() uint16

	// nsq相关配置
	Config() *nsq.Config
}

func setNSQConsumers(lookupd []string, consumers ...NSQConsumer) error {
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

		nc, err := nsq.NewConsumer(c.Topic(), c.Channel(), cfg)
		if err != nil {
			return err
		}

		nc.SetLogger(&NSQLogger{}, nsq.LogLevelError)
		nc.AddHandler(c)

		if err := nc.ConnectToNSQLookupds(lookupd); err != nil {
			return err
		}
	}

	return nil
}

// NextAttemptDelay 一个帮助方法，用于返回下一次尝试的等待时间
func NextAttemptDelay(attempts uint16) time.Duration {
	var d time.Duration

	switch attempts {
	case 0, 1:
		d = 5 * time.Second
	case 2:
		d = 10 * time.Second
	case 3:
		d = 15 * time.Second
	case 4:
		d = 30 * time.Second
	case 5:
		d = time.Minute
	case 6:
		d = 2 * time.Minute
	case 7:
		d = 5 * time.Minute
	case 8:
		d = 10 * time.Minute
	case 9:
		d = 15 * time.Minute
	case 10:
		d = 30 * time.Minute
	default:
		d = time.Hour
	}

	return d
}
