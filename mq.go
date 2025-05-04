package gema

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/riverdriver/riverpgxv5"
	"go.uber.org/fx"
)

// RiverQueueModule is a module that provides a message queue using river.
// User of this module does not need to provide the workers as it will be
// automatically created by this module. You will only need to register your worker
//
// Make sure you have migrated the river schema before using this module with river cli at https://riverqueue.com/docs/migrations#running-migrations.
// This river queue module will create a new connection pool to the database. That means you will have 2 connection pools.
// One for the public schema and one for the river queue schema.
//
// This module will provide pgxpool.Pool and river.Client[pgx.Tx].
func RiverQueueModule(connConfig *pgxpool.Config, queueConfig *river.Config) fx.Option {
	var createQueue = func(lc fx.Lifecycle, workers *river.Workers) (*river.Client[pgx.Tx], *pgxpool.Pool) {
		queueConfig.Workers = workers

		pool, err := pgxpool.NewWithConfig(context.Background(), connConfig)
		if err != nil {
			panic(err)
		}

		client, err := river.NewClient(riverpgxv5.New(pool), queueConfig)
		if err != nil {
			panic(err)
		}

		lc.Append(fx.Hook{
			OnStart: func(ctx context.Context) error {
				if err := pool.Ping(ctx); err != nil {
					return err
				}

				return client.Start(context.Background())
			},
			OnStop: func(ctx context.Context) error {
				if err := client.Stop(ctx); err != nil {
					return err
				}

				pool.Close()
				return nil
			},
		})

		return client, pool
	}

	return fx.Module("messagequeue",
		fx.Provide(river.NewWorkers),
		fx.Provide(createQueue),
	)
}
