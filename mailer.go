package yiigo

import (
	"sync"

	"go.uber.org/zap"
	"gopkg.in/gomail.v2"
)

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

// Send sends an email.
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

func initMailer(name, host string, port int, account, password string) {
	dialer := &EMailDialer{dialer: gomail.NewDialer(host, port, account, password)}

	if name == Default {
		defaultMailer = dialer
	}

	mailerMap.Store(name, dialer)
}

// Mailer returns an email dialer.
func Mailer(name ...string) *EMailDialer {
	if len(name) == 0 {
		if defaultMailer == nil {
			logger.Panic("yiigo: invalid email dialer", zap.String("name", Default))
		}

		return defaultMailer
	}

	v, ok := mailerMap.Load(name[0])

	if !ok {
		logger.Panic("yiigo: invalid email dialer", zap.String("name", name[0]))
	}

	return v.(*EMailDialer)
}
