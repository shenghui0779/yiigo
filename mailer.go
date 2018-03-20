package yiigo

import (
	"sync"

	gomail "gopkg.in/gomail.v2"
)

type emailConfig struct {
	Title    string `toml:"title"`
	Host     string `toml:"host"`
	Port     int    `toml:"port"`
	Username string `toml:"username"`
	Password string `toml:"password"`
}

// Mailer email
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

var (
	emailDialer *gomail.Dialer
	emailMutex  sync.Mutex
)

func emailDial() {
	emailMutex.Lock()
	defer emailMutex.Unlock()

	if emailDialer != nil {
		return
	}

	conf := &emailConfig{}
	Env.Unmarshal("email", conf)

	emailDialer = gomail.NewDialer(conf.Host, conf.Port, conf.Username, conf.Password)
}

// Send send an email.
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

	if emailDialer == nil {
		emailDial()
	}

	// Send the email
	err := emailDialer.DialAndSend(msg)

	return err
}
