package yiigo

import (
	"sync"

	"go.uber.org/zap"
	"gopkg.in/gomail.v2"
)

type emailConfig struct {
	Host     string `toml:"host"`
	Port     int    `toml:"port"`
	Username string `toml:"username"`
	Password string `toml:"password"`
}

// EMail email
type EMail struct {
	Title       string
	Subject     string
	From        string
	To          []string
	Cc          []string
	Body        string
	ContentType string
	Attach      []string
}

// EMailDialer email dialer
type EMailDialer struct {
	dialer *gomail.Dialer
}

// Send send an email.
func (m *EMailDialer) Send(email *EMail, settings ...gomail.MessageSetting) error {
	msg := gomail.NewMessage(settings...)

	msg.SetHeader("Subject", email.Subject)
	msg.SetAddressHeader("From", email.From, email.Title)
	msg.SetHeader("To", email.To...)

	if len(email.Cc) != 0 {
		msg.SetHeader("Cc", email.Cc...)
	}

	if len(email.Attach) != 0 {
		for _, v := range email.Attach {
			msg.Attach(v)
		}
	}

	contentType := "text/html"

	if len(email.ContentType) != 0 {
		contentType = email.ContentType
	}

	msg.SetBody(contentType, email.Body)

	// Send the email
	err := m.dialer.DialAndSend(msg)

	return err
}

var (
	defaultMailer *EMailDialer
	mailerMap     sync.Map
)

func initMailer() {
	configs := make(map[string]*emailConfig, 0)

	if err := env.Get("email").Unmarshal(&configs); err != nil {
		logger.Panic("yiigo: email dialer init error", zap.Error(err))
	}

	if len(configs) == 0 {
		return
	}

	for name, cfg := range configs {
		dialer := &EMailDialer{dialer: gomail.NewDialer(cfg.Host, cfg.Port, cfg.Username, cfg.Password)}

		if name == defalutConn {
			defaultMailer = dialer
		}

		mailerMap.Store(name, dialer)
	}
}

// Mailer returns an email dialer.
func Mailer(name ...string) *EMailDialer {
	if len(name) == 0 {
		if defaultMailer == nil {
			logger.Panic("yiigo: invalid email dialer", zap.String("name", defalutConn))
		}

		return defaultMailer
	}

	v, ok := mailerMap.Load(name[0])

	if !ok {
		logger.Panic("yiigo: invalid email dialer", zap.String("name", name[0]))
	}

	return v.(*EMailDialer)
}
