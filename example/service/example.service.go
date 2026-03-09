package service

import (
	"context"
	"io"
	"path/filepath"

	"github.com/google/uuid"
	"github.com/thoriqadillah/gema"
	"go.uber.org/fx"
)

type ExampleService struct {
	db       *gema.DB
	store    Store
	storage  gema.Storage
	notifier gema.Notifier
}

func newService(
	db *gema.DB,
	store Store,
	storageFactory gema.StorageFactory,
	notifierFactory gema.NotifierFactory,
) *ExampleService {
	return &ExampleService{
		db:       db,
		store:    store,
		storage:  storageFactory.Disk(gema.LocalStorage),
		notifier: notifierFactory.Create(gema.EmailNotifier),
	}
}

func (s *ExampleService) Hello(ctx context.Context) string {
	return s.store.Hello(ctx)
}

func (s *ExampleService) Notification(ctx context.Context) error {
	return s.notifier.Send(ctx, gema.Message{
		To:       []string{"hello@gema.com"},
		Subject:  "Hello World",
		Template: "example.html",
	})
}

func (s *ExampleService) Transaction(ctx context.Context) (message string, err error) {
	err = s.db.TransactionFunc(ctx, func(ctx context.Context) error {
		message = s.store.Hello(ctx)
		return s.store.Foo(ctx)
	})

	return message, err
}

func (s *ExampleService) Upload(file io.Reader, filename string) (url string, err error) {
	ext := filepath.Ext(filename)
	id := uuid.NewString()
	filename = id + ext

	url, err = s.storage.Upload(filename, file)
	if err != nil {
		return "", err
	}

	return url, nil
}

func NewExample() fx.Option {
	return fx.Module("example.service",
		fx.Provide(fx.Private, newStore),
		fx.Provide(newService),
	)
}
