package gema

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/riverdriver/riverpgxv5"
	"go.uber.org/fx"
)

var workers = river.NewWorkers()

type WorkerFactory func(r *river.Workers)

func RegisterRiverWorker(factory WorkerFactory) {
	factory(workers)
}

// RiverQueueModule is a module that provides a message queue using river
// User of this module does not need to provide the workers as it will be
// automatically created by this module. You will only need to register your worker
//
// Make sure you have migrated the river schema before using this module
func RiverQueueModule(config *river.Config) fx.Option {
	conf := config
	conf.Workers = workers

	var createQueue = func(pool *pgxpool.Pool) *river.Client[pgx.Tx] {
		client, err := river.NewClient(riverpgxv5.New(pool), conf)
		if err != nil {
			panic(err)
		}

		return client
	}

	var invokeQueue = func(lc fx.Lifecycle, client *river.Client[pgx.Tx]) {
		lc.Append(fx.Hook{
			OnStart: func(ctx context.Context) error {
				fmt.Println("[Gema] Starting message queue")
				return client.Start(ctx)
			},
			OnStop: func(ctx context.Context) error {
				err := client.Stop(ctx)
				fmt.Println("[Gema] Message queue stoped")

				return err
			},
		})
	}

	return fx.Module("messagequeue",
		fx.Provide(createQueue),
		fx.Invoke(invokeQueue),
	)
}
