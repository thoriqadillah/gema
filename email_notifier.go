package gema

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"path/filepath"

	"go.uber.org/fx"
	"gopkg.in/gomail.v2"
)

const EmailNotifier NotifierName = "email"

type EmailerOption struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
	Name     string
	Env      string
	Template *template.Template
}

type emailer struct {
	opt    *EmailerOption
	mailer *gomail.Dialer
}

func newEmailNotifier(o *EmailerOption) Notifier {
	return &emailer{
		opt: o,
		mailer: gomail.NewDialer(
			o.Host,
			o.Port,
			o.Username,
			o.Password,
		),
	}
}

func (e *emailer) Send(ctx context.Context, m Message) error {
	if e.opt.Env == "development" {
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
		if err := e.opt.Template.ExecuteTemplate(&buff, m.Body, m.Data); err != nil {
			return err
		}

		msg = buff.String()
	}

	from := m.From
	if from == "" {
		from = fmt.Sprintf("%s <%s>", e.opt.Name, e.opt.From)
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

type emailerProvider struct {
	opt *EmailerOption
}

func EmailerProvider(opt *EmailerOption) NotifierProvider {
	return &emailerProvider{
		opt: opt,
	}
}

func (e *emailerProvider) registerEmailer(registry NotifierRegistry) {
	notifier := newEmailNotifier(e.opt)
	registry.Register(EmailNotifier, notifier)
}

func (e *emailerProvider) Register() fx.Option {
	return fx.Module("notifier.email", fx.Invoke(e.registerEmailer))
}
