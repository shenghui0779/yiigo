package yiigo

import (
	"errors"
	"fmt"
	"time"

	"github.com/nsqio/go-nsq"
	"go.uber.org/zap"
)

var producer *nsq.Producer

// NSQLogger NSQ logger
type NSQLogger struct{}

// Output implements the NSQ logger interface
func (l *NSQLogger) Output(calldepth int, s string) error {
	logger.Error(fmt.Sprintf("[NSQ] %s", s), zap.Int("call_depth", calldepth))

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
		logger.Panic("[yiigo] err nsq ping", zap.String("nsqd", nsqd), zap.Error(err))
	}

	producer.SetLogger(&NSQLogger{}, nsq.LogLevelError)

	return nil
}

// NSQPublish synchronously publishes a message body to the specified topic.
func NSQPublish(topic string, msg []byte) error {
	if producer == nil {
		return errors.New("producer is nil (forgotten configure?)")
	}

	return producer.Publish(topic, msg)
}

// NSQDeferredPublish synchronously publishes a message body to the specified topic
// where the message will queue at the channel level until the timeout expires.
func NSQDeferredPublish(topic string, msg []byte, duration time.Duration) error {
	if producer == nil {
		return errors.New("producer is nil (forgotten configure?)")
	}

	return producer.DeferredPublish(topic, duration, msg)
}

// NSQConsumer NSQ consumer
type NSQConsumer interface {
	nsq.Handler
	Topic() string
	Channel() string
	Attempts() uint16
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
			if err := cfg.Set("max_attempts", c.Attempts()); err != nil {
				return err
			}
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

// NextAttemptDelay returns the delay time for next attempt.
func NextAttemptDelay(attempts uint16) time.Duration {
	var d time.Duration

	switch attempts {
	case 0:
		d = 5 * time.Second
	case 1:
		d = 10 * time.Second
	case 2:
		d = 15 * time.Second
	case 3:
		d = 30 * time.Second
	case 4:
		d = 1 * time.Minute
	case 5:
		d = 2 * time.Minute
	case 6:
		d = 5 * time.Minute
	case 7:
		d = 10 * time.Minute
	case 8:
		d = 15 * time.Minute
	case 9:
		d = 30 * time.Minute
	case 10:
		d = 1 * time.Hour
	default:
		d = 1 * time.Hour
	}

	return d
}
