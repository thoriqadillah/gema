package gema

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"io/fs"

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

	// TemplateFs will be used to parse the template using html/template
	TemplateFs fs.FS
}

type emailer struct {
	opt      *EmailerOption
	mailer   *gomail.Dialer
	template *template.Template
}

func newEmailNotifier(o *EmailerOption) Notifier {
	tmpl, err := template.ParseFS(o.TemplateFs, "**/*.html")
	if err != nil {
		panic(err)
	}

	return &emailer{
		opt:      o,
		template: tmpl,
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

	var msg string
	var mimetype string

	if m.Text != "" {
		mimetype = "text/plain"
		msg = m.Html
	} else if m.Template != "" {
		if e.opt.TemplateFs == nil {
			return fmt.Errorf("template fs not provided")
		}

		mimetype = "text/html"
		var buff bytes.Buffer
		if err := e.template.ExecuteTemplate(&buff, m.Template, m.Data); err != nil {
			return err
		}
		msg = buff.String()
	} else if m.Html != "" {
		mimetype = "text/html"
		msg = m.Html
	} else {
		return fmt.Errorf("no body provided")
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
