package gema

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"text/template"

	"gopkg.in/gomail.v2"
)

const EmailNotifier NotifierName = "email"

type Emailer struct {
	template *template.Template
	mailer   *gomail.Dialer
	from     string
	name     string
	env      string
}

func newEmailNotifier(o *NotifierOption) Notifier {
	return &Emailer{
		template: o.Template,
		env:      o.Env,
		from:     o.From,
		name:     o.Name,
		mailer: gomail.NewDialer(
			o.Host,
			o.Port,
			o.Username,
			o.Password,
		),
	}
}

func (e *Emailer) Send(ctx context.Context, m Message) error {
	if e.env == "development" {
		b, err := json.MarshalIndent(m, "", "  ")
		if err != nil {
			return err
		}

		fmt.Println(string(b))
		return nil
	}

	mimetype := "text/plain"
	msg := m.Body

	ext := filepath.Ext(m.Body)
	if ext == ".html" {
		mimetype = "text/html"
		var buff bytes.Buffer
		if err := e.template.ExecuteTemplate(&buff, m.Body, m.Data); err != nil {
			return err
		}

		msg = buff.String()
	}

	from := m.From
	if from == "" {
		from = fmt.Sprintf("%s <%s>", e.name, e.from)
	}

	message := gomail.NewMessage()
	message.SetHeader("From", from)
	message.SetHeader("To", m.To...)
	message.SetHeader("Subject", m.Subject)
	message.SetHeader("Bcc", m.Bcc...)
	message.SetHeader("Cc", m.Cc...)
	message.SetBody(mimetype, msg)

	return e.mailer.DialAndSend(message)
}

func init() {
	RegisterNotifier(EmailNotifier, newEmailNotifier)
}
