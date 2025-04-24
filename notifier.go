package gema

import (
	"context"
	"log"

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

type NotifierRegistry map[NotifierName]Notifier

func (s NotifierRegistry) Register(name NotifierName, notifier Notifier) {
	s[name] = notifier
}

type NotifierProvider interface {
	// Register will be used to register your notifier implementation and returns your notifier module.
	// In the register, you will be provided with the notifier registry to register your notifier implementation.
	// Register must return your notifier module with fx.Option. But remember to make your notifier
	// implementation private. Otherwise, it will collide with other notifier implementations
	Register() fx.Option
}

type NotifierFactory interface {
	Create(driver NotifierName) Notifier
}

type notifierFactory struct {
	registry NotifierRegistry
}

func createNotifier(s NotifierRegistry) NotifierFactory {
	return &notifierFactory{s}
}

func (s *notifierFactory) Create(driver NotifierName) Notifier {
	notifier, ok := s.registry[driver]
	if !ok {
		log.Fatalf("[Gema] Notifier with %s provider not found", driver)
		return nil
	}

	return notifier
}

func NotifierModule(providers ...NotifierProvider) fx.Option {
	notifierMap := NotifierRegistry{}

	fxOptions := []fx.Option{
		fx.Provide(fx.Private, func() NotifierRegistry {
			return notifierMap
		}),
	}

	for _, provider := range providers {
		fxOptions = append(fxOptions, provider.Register())
	}

	fxOptions = append(fxOptions, fx.Provide(createNotifier))

	return fx.Module("notifier", fxOptions...)
}
