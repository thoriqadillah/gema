package gema

import (
	"bytes"
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"path/filepath"
	"text/template"

	"gopkg.in/gomail.v2"
)

const EmailNotifier NotifierName = "email"

// WithAppEnv sets the environment, it can be "development" or "production"
func WithAppEnv(env string) NotifierOptionFunc {
	return func(o *NotifierOption) {
		o.Env = env
	}
}

func WithMailerName(name string) NotifierOptionFunc {
	return func(o *NotifierOption) {
		o.Name = name
	}
}

func WithMailerSender(from string) NotifierOptionFunc {
	return func(o *NotifierOption) {
		o.From = from
	}
}

func WithMailerTemplateFs(templateFs embed.FS, templatePattern string) NotifierOptionFunc {
	return func(o *NotifierOption) {
		template, err := template.ParseFS(templateFs, templatePattern)
		if err != nil {
			panic(err)
		}

		o.Template = template
	}
}

func WithMailerPassword(password string) NotifierOptionFunc {
	return func(o *NotifierOption) {
		o.Password = password
	}
}

func WithMailerUsername(username string) NotifierOptionFunc {
	return func(o *NotifierOption) {
		o.Username = username
	}
}

func WithMailerHost(host string) NotifierOptionFunc {
	return func(o *NotifierOption) {
		o.Host = host
	}
}

func WithMailerPort(port int) NotifierOptionFunc {
	return func(o *NotifierOption) {
		o.Port = port
	}
}

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
