package yiigo

import (
	"time"

	"github.com/nsqio/go-nsq"
	"go.uber.org/zap"
)

var producer *nsq.Producer

// NSQLogger NSQ logger
type NSQLogger struct{}

// Output implements the NSQ logger interface
func (l *NSQLogger) Output(calldepth int, s string) error {
	logger.Error(s, zap.Int("call_depth", calldepth))

	return nil
}

func initProducer(nsqd string) error {
	p, err := nsq.NewProducer(nsqd, nsq.NewConfig())

	if err != nil {
		logger.Error("init producer error", zap.Error(err))

		return err
	}

	p.SetLogger(&NSQLogger{}, nsq.LogLevelError)

	producer = p

	return nil
}

// NSQMessage NSQ message
type NSQMessage interface {
	Bytes() ([]byte, error)
	// Do message processing
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

// NSQDeferredPublish synchronously publishes a message body to the specified topic
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

type nsqSetting struct {
	lookupdPollInterval     time.Duration
	rdyRedistributeInterval time.Duration
	maxInFlight             int
	consumers               []NSQConsumer
}

// NSQOption configures how we set up the nsq config.
type NSQOption func(s *nsqSetting)

// WithLookupdPollInterval specifies the `LookupdPollInterval` for nsq config.
func WithLookupdPollInterval(t time.Duration) NSQOption {
	return func(s *nsqSetting) {
		s.lookupdPollInterval = t
	}
}

// WithRDYRedistributeInterval specifies the `RDYRedistributeInterval` for nsq config.
func WithRDYRedistributeInterval(t time.Duration) NSQOption {
	return func(s *nsqSetting) {
		s.rdyRedistributeInterval = t
	}
}

// WithMaxInFlight specifies the `MaxInFlight` for nsq config.
func WithMaxInFlight(n int) NSQOption {
	return func(s *nsqSetting) {
		s.maxInFlight = n
	}
}

// WithNSQConsumers specifies the consumers for nsq.
func WithNSQConsumers(consumers ...NSQConsumer) NSQOption {
	return func(s *nsqSetting) {
		s.consumers = append(s.consumers, consumers...)
	}
}

func setConsumers(lookupd []string, options ...NSQOption) error {
	setting := &nsqSetting{
		lookupdPollInterval:     time.Second,
		rdyRedistributeInterval: time.Second,
		maxInFlight:             1000,
	}

	for _, f := range options {
		f(setting)
	}

	for _, c := range setting.consumers {
		cfg := nsq.NewConfig()

		cfg.LookupdPollInterval = setting.lookupdPollInterval
		cfg.RDYRedistributeInterval = setting.rdyRedistributeInterval
		cfg.MaxInFlight = setting.maxInFlight

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

func initNSQ(nsqd string, lookupd []string, options ...NSQOption) {
	// init producer
	if err := initProducer(nsqd); err != nil {
		logger.Panic("[yiigo] init nsq error", zap.Error(err))
	}

	// set consumers
	if err := setConsumers(lookupd, options...); err != nil {
		logger.Panic("[yiigo] init nsq error", zap.Error(err))
	}

	logger.Info("[yiigo] nsq is OK")
}

// NextAttemptDuration helper for attempt duration.
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
	default:
		d = 1 * time.Hour
	}

	return d
}
