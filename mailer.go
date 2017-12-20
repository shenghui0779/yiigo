package yiigo

import (
	"gopkg.in/gomail.v2"
)

type Mailer struct {
	Title    string
	Subject  string
	From     string
	To       []string
	Cc       []string
	Content  string
	Attach   []string
	Charset  string
	Encoding gomail.Encoding
}

// Send send a mail
func (m *Mailer) Send() error {
	msgSettings := []gomail.MessageSetting{}

	if m.Charset != "" {
		msgSettings = append(msgSettings, gomail.SetCharset(m.Charset))
	}

	if m.Encoding != "" {
		msgSettings = append(msgSettings, gomail.SetEncoding(m.Encoding))
	}

	msg := gomail.NewMessage(msgSettings...)

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

	msg.SetBody("text/html", m.Content)

	d := m.dialer()

	// Send the email
	err := d.DialAndSend(msg)

	return err
}

func (m *Mailer) dialer() *gomail.Dialer {
	host := EnvString("email", "host", "smtp.example.com")
	port := EnvInt("email", "port", 587)
	username := EnvString("email", "username", "yiigo@example.com")
	password := EnvString("email", "password", "xxxxxxxxx")

	d := gomail.NewDialer(host, port, username, password)

	return d
}
