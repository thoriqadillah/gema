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

func StartQueue(queueConfig map[string]river.QueueConfig) fx.Option {
	return fx.Module("start_queue",
		fx.Supply(queueConfig),
		fx.Supply(river.NewWorkers()),
		fx.Provide(fx.Private, newServer),
		fx.Invoke(func(lc fx.Lifecycle, river *river.Client[pgx.Tx]) {
			ctx, cancel := context.WithCancel(context.Background())
			lc.Append(fx.Hook{
				OnStart: func(_ context.Context) error {
					return river.Start(ctx)
				},
				OnStop: func(stopCtx context.Context) error {
					cancel()
					return river.Stop(stopCtx)
				},
			})
		}),
	)
}
