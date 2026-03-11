package example

import (
	"context"
	"database/sql"
	"example/helpers/db"
	"fmt"
	"io"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"github.com/riverqueue/river"
	"github.com/thoriqadillah/gema"
	"go.uber.org/fx"
)

type PrintArg struct {
	Message string `json:"message"`
}

func (PrintArg) Kind() string {
	return "print"
}

type PrintWorker struct {
	river.WorkerDefaults[PrintArg]
}

func (w *PrintWorker) Work(ctx context.Context, job *river.Job[PrintArg]) error {
	time.Sleep(3 * time.Second)
	fmt.Println("Print after 3s:", job.Args.Message)
	return nil
}

func (r *PrintWorker) Register(workers *river.Workers) {
	river.AddWorker(workers, r)
}

func newWorker() gema.WorkerRegistrar {
	return &PrintWorker{}
}

func AsWorker(constructor any) any {
	return fx.Annotate(
		constructor,
		fx.As(new(Worker)),
		fx.ResultTags(`group:"workers"`),
	)
}

type ExampleService struct {
	db       *gema.DB
	store    Store
	storage  gema.Storage
	notifier gema.Notifier
	queue    *river.Client[*sql.Tx]
}

func newService(
	db *gema.DB,
	store Store,
	queue *river.Client[*sql.Tx],
	storageFactory gema.StorageFactory,
	notifierFactory gema.NotifierFactory,
) *ExampleService {
	return &ExampleService{
		db:       db,
		store:    store,
		queue:    queue,
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

func (s *ExampleService) Upload(ctx context.Context, file io.Reader, filename string) (url string, err error) {
	ext := filepath.Ext(filename)
	id := uuid.NewString()
	filename = id + ext

	url, err = s.storage.Upload(ctx, filename, file)
	if err != nil {
		return "", err
	}

	return url, nil
}

func (s *ExampleService) QueueJob(ctx context.Context) error {
	return s.db.TransactionFunc(ctx, func(ctx context.Context) error {
		tx := db.UnwrapTx(s.db.Tx(ctx))
		_, err := s.queue.InsertTx(ctx, tx, PrintArg{"hello"}, nil)
		return err
	})
}
