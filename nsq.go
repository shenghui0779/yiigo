package yiigo

import (
	"errors"
	"fmt"
	"time"

	"github.com/nsqio/go-nsq"
	"go.uber.org/zap"
)

var producer *nsq.Producer

// NSQLogger nsq logger
type NSQLogger struct{}

// Output implements the nsq logger interface
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
		logger.Panic("err nsq ping", zap.String("nsqd", nsqd), zap.Error(err))
	}

	producer.SetLogger(&NSQLogger{}, nsq.LogLevelError)

	return nil
}

// NSQPublish synchronously publishes a message body to the specified topic.
func NSQPublish(topic string, msg []byte) error {
	if producer == nil {
		return errors.New("nsq producer is nil (forgotten configure?)")
	}

	return producer.Publish(topic, msg)
}

// NSQDeferredPublish synchronously publishes a message body to the specified topic
// where the message will queue at the channel level until the timeout expires.
func NSQDeferredPublish(topic string, msg []byte, duration time.Duration) error {
	if producer == nil {
		return errors.New("nsq producer is nil (forgotten configure?)")
	}

	return producer.DeferredPublish(topic, duration, msg)
}

// NSQConsumer nsq consumer
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

// NextAttemptDelay returns the delay time for nsq next attempt.
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
