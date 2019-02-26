package yiigo

import (
	"gopkg.in/gomail.v2"
)

var email *gomail.Dialer

// emailOptions email options
type emailOptions struct {
	charset     string
	encoding    gomail.Encoding
	contentType string
}

// EMailOption configures how we set up the db
type EMailOption interface {
	apply(options *emailOptions)
}

// funcEMailOption implements email option
type funcEMailOption struct {
	f func(options *emailOptions)
}

func (fo *funcEMailOption) apply(o *emailOptions) {
	fo.f(o)
}

func newFuncEMailOption(f func(options *emailOptions)) *funcEMailOption {
	return &funcEMailOption{f: f}
}

// WithEMailCharset specifies the `Charset` to email.
func WithEMailCharset(s string) EMailOption {
	return newFuncEMailOption(func(o *emailOptions) {
		o.charset = s
	})
}

// WithEMailEncoding specifies the `Encoding` to email.
func WithEMailEncoding(e gomail.Encoding) EMailOption {
	return newFuncEMailOption(func(o *emailOptions) {
		o.encoding = e
	})
}

// WithEMailContentType specifies the `ContentType` to email.
func WithEMailContentType(s string) EMailOption {
	return newFuncEMailOption(func(o *emailOptions) {
		o.contentType = s
	})
}

// Mailer email
type Mailer struct {
	Title   string
	Subject string
	From    string
	To      []string
	Cc      []string
	Content string
	Attach  []string
}

// Send send an email.
func (m *Mailer) Send(options ...EMailOption) error {
	o := &emailOptions{contentType: "text/html"}

	if len(options) > 0 {
		for _, option := range options {
			option.apply(o)
		}
	}

	settings := make([]gomail.MessageSetting, 0, 2)

	if o.charset != "" {
		settings = append(settings, gomail.SetCharset(o.charset))
	}

	if o.encoding != "" {
		settings = append(settings, gomail.SetEncoding(o.encoding))
	}

	msg := gomail.NewMessage(settings...)

	msg.SetHeader("Subject", m.Subject)
	msg.SetAddressHeader("From", m.From, m.Title)
	msg.SetHeader("To", m.To...)

	if len(m.Cc) > 0 {
		msg.SetHeader("Cc", m.Cc...)
	}

	if len(m.Attach) > 0 {
		for _, v := range m.Attach {
			msg.Attach(v)
		}
	}

	msg.SetBody(o.contentType, m.Content)

	// Send the email
	err := email.DialAndSend(msg)

	return err
}

// UseEMail use email
func UseEMail(host string, port int, username, password string) {
	email = gomail.NewDialer(host, port, username, password)
}
