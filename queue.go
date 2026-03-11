package gema

import (
	"context"
	"database/sql"
	"log"

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
		panic(err)
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

func AsWorker(constructor any) any {
	return fx.Annotate(
		constructor,
		fx.As(new(river.Worker[river.JobArgs])),
		fx.ResultTags(`group:"workers"`),
	)
}

type queueParams struct {
	fx.In

	lc    fx.Lifecycle
	river *river.Client[pgx.Tx]
	w     *river.Workers

	Workers []river.Worker[river.JobArgs] `group:"workers"`
}

func StartQueue(queueConfig map[string]river.QueueConfig) fx.Option {
	return fx.Module("start_queue",
		fx.Supply(queueConfig),
		fx.Supply(river.NewWorkers()),
		fx.Provide(fx.Private, newServer),
		fx.Invoke(func(p queueParams) {
			for _, worker := range p.Workers {
				river.AddWorker(p.w, worker)
			}

			ctx, cancel := context.WithCancel(context.Background())
			p.lc.Append(fx.Hook{
				OnStart: func(_ context.Context) error {
					return p.river.Start(ctx)
				},
				OnStop: func(stopCtx context.Context) error {
					cancel()
					return p.river.Stop(stopCtx)
				},
			})
		}),
	)
}
