package gema

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/riverdriver/riverdatabasesql"
	"github.com/riverqueue/river/riverdriver/riverpgxv5"
	"go.uber.org/fx"
)

func newClient(sql *sql.DB) *river.Client[*sql.Tx] {
	river, err := river.NewClient(riverdatabasesql.New(sql), &river.Config{})
	if err != nil {
		fmt.Printf("[Gema] Failed to create River client: %v", err)
		os.Exit(1)
	}

	return river
}

func QueueModule() fx.Option {
	return fx.Module("queue",
		fx.Provide(newClient),
	)
}

func newServer(queueConfig map[string]river.QueueConfig, pool *pgxpool.Pool, workers *river.Workers) *river.Client[pgx.Tx] {
	client, err := river.NewClient(riverpgxv5.New(pool), &river.Config{
		Queues:  queueConfig,
		Workers: workers,
	})

	if err != nil {
		log.Fatal(err)
	}

	return client
}

type QueueWorker interface {
	Register(workers *river.Workers)
}

func AsWorker(constructor any) any {
	return fx.Annotate(
		constructor,
		fx.As(new(QueueWorker)),
		fx.ResultTags(`group:"workers"`),
	)
}

type queueParams struct {
	fx.In

	fx.Lifecycle
	*river.Client[pgx.Tx]
	*river.Workers
	QueueWorker []QueueWorker `group:"workers"`
}

func StartQueue(queueConfig map[string]river.QueueConfig) fx.Option {
	return fx.Module("start_queue",
		fx.Supply(queueConfig),
		fx.Supply(river.NewWorkers()),
		fx.Provide(fx.Private, newServer),
		fx.Invoke(func(p queueParams) {
			for _, worker := range p.QueueWorker {
				worker.Register(p.Workers)
			}

			ctx, cancel := context.WithCancel(context.Background())
			p.Append(fx.Hook{
				OnStart: func(_ context.Context) error {
					return p.Start(ctx)
				},
				OnStop: func(stopCtx context.Context) error {
					cancel()
					return p.Stop(stopCtx)
				},
			})
		}),
	)
}
