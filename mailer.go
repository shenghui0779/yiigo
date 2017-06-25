package yiigo

import (
	"fmt"

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

/**
 * 发送邮件
 * subject string 邮件主题
 * content string 邮件主题
 * to []string 邮件接收人
 * cc ...string 邮件抄送
 */
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

	if err != nil {
		return fmt.Errorf("[Mailer] %v", err)
	}

	return nil
}

func (m *Mailer) dialer() *gomail.Dialer {
	host := GetEnvString("email", "host", "smtp.example.com")
	port := GetEnvInt("email", "port", 587)
	username := GetEnvString("email", "username", "yiigo@example.com")
	password := GetEnvString("email", "password", "xxxxxxxxx")

	d := gomail.NewDialer(host, port, username, password)

	return d
}
