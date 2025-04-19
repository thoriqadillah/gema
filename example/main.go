package main

import (
	"context"
	"embed"
	"fmt"
	"sort"

	"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/riverqueue/river"
	"github.com/thoriqadillah/gema"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type SortArgs struct {
	// Strings is a slice of strings to sort.
	Strings []string `json:"strings"`
}

func (SortArgs) Kind() string { return "sort" }

type SortWorker struct {
	// An embedded WorkerDefaults sets up default methods to fulfill the rest of
	// the Worker interface:
	river.WorkerDefaults[SortArgs]
}

func (w *SortWorker) Work(ctx context.Context, job *river.Job[SortArgs]) error {
	sort.Strings(job.Args.Strings)
	fmt.Printf("Sorted strings: %+v\n", job.Args.Strings)
	return nil
}

//go:embed templates
var template embed.FS

func httpServer(logger *zap.Logger) *echo.Echo {
	e := echo.New()
	e.Use(gema.LoggerMiddleware(logger))
	e.Use(middleware.Gzip())

	return e
}

func main() {
	godotenv.Load()

	gema.RegisterValidator(map[string]validator.Func{
		"password": func(fl validator.FieldLevel) bool {
			// TODO
			return true
		},
	})

	app := fx.New(
		fx.NopLogger,
		gema.LoggerModule(APP_ENV),
		fx.Provide(httpServer),
		gema.DatabaseModule(DB_URL),
		gema.RiverQueueModule(&river.Config{
			Queues: map[string]river.QueueConfig{
				river.QueueDefault: {
					MaxWorkers: 100,
				},
			},
			Workers: gema.RegisterRiverWorker(func(w *river.Workers) {
				river.AddWorker(w, &SortWorker{})
			}),
		}),
		gema.NotifierModule(
			gema.WithMailerName("gema"),
			gema.WithMailerSender("hello@gema.com"),
			gema.WithAppEnv(APP_ENV),
			gema.WithMailerTemplateFs(template, "templates/*.html"),
		),
		gema.StorageModule(
			gema.LocalStorage,
			gema.WithStorageUrlPath(fmt.Sprintf("http://localhost%s/storage", PORT)),
		),
		exampleModule,
		gema.Start(":8001"),
	)

	app.Run()
}
