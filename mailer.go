package yiigo

import (
	"sync"

	"github.com/pelletier/go-toml"
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
func (m *EMailDialer) Send(e *EMail, settings ...gomail.MessageSetting) error {
	msg := gomail.NewMessage(settings...)

	msg.SetHeader("Subject", e.Subject)
	msg.SetAddressHeader("From", e.From, e.Title)
	msg.SetHeader("To", e.To...)

	if len(e.Cc) > 0 {
		msg.SetHeader("Cc", e.Cc...)
	}

	if len(e.Attach) > 0 {
		for _, v := range e.Attach {
			msg.Attach(v)
		}
	}

	contentType := "text/html"

	if len(e.ContentType) > 0 {
		contentType = e.ContentType
	}

	msg.SetBody(contentType, e.Body)

	// Send the email
	err := m.dialer.DialAndSend(msg)

	return err
}

var (
	defaultMailer *EMailDialer
	mailerMap     sync.Map
)

func initMailer() {
	tree, ok := env.get("email").(*toml.Tree)

	if !ok {
		return
	}

	keys := tree.Keys()

	if len(keys) == 0 {
		return
	}

	for _, v := range keys {
		node, ok := tree.Get(v).(*toml.Tree)

		if !ok {
			continue
		}

		cfg := new(emailConfig)

		if err := node.Unmarshal(cfg); err != nil {
			logger.Error("yiigo: email dialer init error", zap.String("name", v), zap.Error(err))
		}

		dialer := &EMailDialer{dialer: gomail.NewDialer(cfg.Host, cfg.Port, cfg.Username, cfg.Password)}

		if v == AsDefault {
			defaultMailer = dialer
		}

		mailerMap.Store(v, dialer)
	}
}

// Mailer returns an email dialer.
func Mailer(name ...string) *EMailDialer {
	if len(name) == 0 {
		if defaultMailer == nil {
			logger.Panic("yiigo: invalid email dialer", zap.String("name", AsDefault))
		}

		return defaultMailer
	}

	v, ok := mailerMap.Load(name[0])

	if !ok {
		logger.Panic("yiigo: invalid email dialer", zap.String("name", name[0]))
	}

	return v.(*EMailDialer)
}
