package gema

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/riverdriver/riverpgxv5"
	"go.uber.org/fx"
)

// RiverQueueModule is a module that provides a message queue using river
// User of this module does not need to provide the workers as it will be
// automatically created by this module. You will only need to register your worker
//
// Make sure you have migrated the river schema before using this module
func RiverQueueModule(config *river.Config) fx.Option {
	var createQueue = func(lc fx.Lifecycle, pool *pgxpool.Pool, workers *river.Workers) *river.Client[pgx.Tx] {
		config.Workers = workers
		client, err := river.NewClient(riverpgxv5.New(pool), config)
		if err != nil {
			panic(err)
		}

		lc.Append(fx.Hook{
			OnStart: func(ctx context.Context) error {
				return client.Start(context.Background())
			},
			OnStop: func(ctx context.Context) error {
				return client.Stop(ctx)
			},
		})

		return client
	}

	return fx.Module("messagequeue",
		fx.Provide(river.NewWorkers),
		fx.Provide(createQueue),
	)
}
