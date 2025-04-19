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

type WorkerFactory func(w *river.Workers)

func RegisterRiverWorker(factory WorkerFactory) {
	factory(workers)
}

// RiverQueueModule is a module that provides a message queue using river
// User of this module does not need to provide the workers as it will be
// automatically created by this module. You will only need to register your worker
//
// Make sure you have migrated the river schema before using this module
func RiverQueueModule(config *river.Config) fx.Option {
	config.Workers = workers
	var createQueue = func(lc fx.Lifecycle, pool *pgxpool.Pool) *river.Client[pgx.Tx] {
		fmt.Println("[Gema] Registering river queue module")
		client, err := river.NewClient(riverpgxv5.New(pool), config)
		if err != nil {
			panic(err)
		}

		lc.Append(fx.Hook{
			OnStart: func(ctx context.Context) error {
				return client.Start(context.Background())
			},
			OnStop: func(ctx context.Context) error {
				err := client.Stop(ctx)
				fmt.Println("[Gema] River queue stoped")

				return err
			},
		})

		return client
	}

	return fx.Module("messagequeue", fx.Provide(createQueue))
}
