package yiigo

import "gopkg.in/gomail.v2"

type emailConfig struct {
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
	conf := &emailConfig{}

	Env.Unmarshal("email", conf)

	d := gomail.NewDialer(conf.Host, conf.Port, conf.Username, conf.Password)

	return d
}
