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

func QueueClient() fx.Option {
	return fx.Module("queue_client",
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

func QueueServer(queueConfig map[string]river.QueueConfig) fx.Option {
	return fx.Module("queue_server",
		fx.Supply(queueConfig),
		fx.Supply(river.NewWorkers()),
		fx.Provide(fx.Private, newServer),
		fx.Invoke(func(lc fx.Lifecycle, client *river.Client[pgx.Tx], workers *river.Workers) {
			ctx, cancel := context.WithCancel(context.Background())
			lc.Append(fx.Hook{
				OnStart: func(_ context.Context) error {
					return client.Start(ctx)
				},
				OnStop: func(stopCtx context.Context) error {
					cancel()
					return client.Stop(stopCtx)
				},
			})
		}),
	)
}
