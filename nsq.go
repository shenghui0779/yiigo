package yiigo

import (
	"time"

	"github.com/nsqio/go-nsq"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

var producer *nsq.Producer

// NSQLogger NSQ logger
type NSQLogger struct{}

// Output 日志输出
func (l *NSQLogger) Output(calldepth int, s string) error {
	Logger().Error(s, zap.Int("call_depth", calldepth))

	return nil
}

func initProducer(nsqd string) error {
	p, err := nsq.NewProducer(nsqd, nsq.NewConfig())

	if err != nil {
		Logger().Error("init producer error", zap.Error(err))

		return err
	}

	p.SetLogger(&NSQLogger{}, nsq.LogLevelError)

	producer = p

	return nil
}

// NSQMessage NSQ message
type NSQMessage interface {
	Bytes() ([]byte, error)
	Do() error
}

// NSQPublish synchronously publishes a message body to the specified topic.
func NSQPublish(topic string, msg NSQMessage) error {
	b, err := msg.Bytes()

	if err != nil {
		return err
	}

	return producer.Publish(topic, b)
}

// DeferredPublish synchronously publishes a message body to the specified topic
// where the message will queue at the channel level until the timeout expires.
func NSQDeferredPublish(topic string, msg NSQMessage, duration time.Duration) error {
	b, err := msg.Bytes()

	if err != nil {
		return err
	}

	return producer.DeferredPublish(topic, duration, b)
}

// NSQConsumer NSQ consumer
type NSQConsumer interface {
	nsq.Handler
	Topic() string
	Channel() string
	AttemptCount() uint16
}

func setConsumers(lookupd []string, consumers ...NSQConsumer) error {
	for _, c := range consumers {
		cfg := nsq.NewConfig()

		cfg.LookupdPollInterval = time.Second
		cfg.RDYRedistributeInterval = time.Second
		cfg.MaxInFlight = 1000

		// set attempt acount, default: 5
		if c.AttemptCount() > 0 {
			if err := cfg.Set("max_attempts", c.AttemptCount()); err != nil {
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

type nsqConfig struct {
	Lookupd []string `toml:"lookupd"`
	Nsqd    string   `toml:"nsqd"`
}

// StartNSQ starts NSQ
func StartNSQ(consumers ...NSQConsumer) error {
	cfg := new(nsqConfig)

	if err := Env("nsq").Unmarshal(cfg); err != nil {
		return err
	}

	// init producer
	if err := initProducer(cfg.Nsqd); err != nil {
		return errors.Wrap(err, "yiigo: init nsq producer error")
	}

	// set consumers
	if err := setConsumers(cfg.Lookupd, consumers...); err != nil {
		return errors.Wrap(err, "yiigo: set nsq consumers error")
	}

	Logger().Info("yiigo: nsq is OK.")

	return nil
}

// NextAttemptDuration helper for attempt duration
func NextAttemptDuration(attempts uint16) time.Duration {
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
	case 11:
		d = 2 * time.Hour
	case 12:
		d = 5 * time.Hour
	case 13:
		d = 10 * time.Hour
	default:
		d = 10 * time.Hour
	}

	return d
}
