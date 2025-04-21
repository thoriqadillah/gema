package gema

import (
	"context"
	"embed"
	"fmt"
	"text/template"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/riverqueue/river"
	"go.uber.org/fx"
)

type Notifier interface {
	Send(ctx context.Context, m Message) error
}

type NotifierName string

type Data map[string]interface{}

type Message struct {
	Subject string
	// Body can be a message string or template name.
	// Example:
	//
	// 1. Hello world -> this will set the body type to be text/plain and pass it as is
	//
	// 2. verify.html -> this will look into template folder and set the body type to be text/html and pass the html file as a template
	Body string
	Data Data
	From string
	To   []string
	Bcc  []string
	Cc   []string
}

type NotifierOption struct {
	Pool  *pgxpool.Pool
	River *river.Client[pgx.Tx]

	Template *template.Template
	Password string
	Username string
	Host     string
	Port     int
	Env      string
	From     string
	Name     string
}

type NotifierOptionFunc func(o *NotifierOption)

type NotifierFactory func(o *NotifierOption) Notifier

func withRiver(river *river.Client[pgx.Tx]) NotifierOptionFunc {
	return func(o *NotifierOption) {
		o.River = river
	}
}

func withPgPool(pool *pgxpool.Pool) NotifierOptionFunc {
	return func(o *NotifierOption) {
		o.Pool = pool
	}
}

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

var notifierProviders = map[NotifierName]NotifierFactory{}

// register will register the implementation of the notifier as the provider
func RegisterNotifier(name NotifierName, impl NotifierFactory) {
	notifierProviders[name] = impl
}

func newNotifier(name NotifierName, opt *NotifierOption) Notifier {
	provider, ok := notifierProviders[name]
	if !ok {
		panic(fmt.Sprintf("Notifier with %s provider not found", name))
	}

	return provider(opt)
}

type NotifierFacade interface {
	Create(name NotifierName) Notifier
}

type notifierFacade struct {
	option *NotifierOption
}

func newNotifierFacade(opts ...NotifierOptionFunc) NotifierFacade {
	opt := &NotifierOption{}

	for _, option := range opts {
		option(opt)
	}

	RegisterRiverWorker(func(w *river.Workers) {
		river.AddWorker(w, &emailWorker{
			emailer: newEmailNotifier(opt),
		})
	})

	return &notifierFacade{option: opt}
}

// Create will create a new notifier with the given name. Note that
// `RiveredEmailNotifier` will be usable if you have registered the river queue module
func (n *notifierFacade) Create(name NotifierName) Notifier {
	return newNotifier(name, n.option)
}

type notifierParams struct {
	fx.In

	Pool  *pgxpool.Pool         `optional:"true"`
	River *river.Client[pgx.Tx] `optional:"true"`
}

// NotifierModule will provide a notifier facade that can be used to create a notifier.
// Note that RiverdEmailNotifier will only be available
// if you have registered the river queue module
func NotifierModule(opts ...NotifierOptionFunc) fx.Option {
	return fx.Module("notifier",
		fx.Provide(func(p notifierParams) NotifierFacade {
			opts = append(opts,
				withRiver(p.River),
				withPgPool(p.Pool),
			)

			return newNotifierFacade(opts...)
		}),
	)
}
