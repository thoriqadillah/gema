package gema

import (
	"database/sql"

	"github.com/riverqueue/river"
	"github.com/riverqueue/river/riverdriver/riverdatabasesql"
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
// This module will provide pgxpool.Pool and river.Client[*sql.Tx].
func RiverQueueModule(queueConfig *river.Config) fx.Option {
	var createQueue = func(lc fx.Lifecycle, sqldb *sql.DB, workers *river.Workers) *river.Client[*sql.Tx] {
		queueConfig.Workers = workers

		client, err := river.NewClient(riverdatabasesql.New(sqldb), queueConfig)
		if err != nil {
			panic(err)
		}

		return client
	}

	return fx.Module("message_queue",
		fx.Provide(river.NewWorkers),
		fx.Provide(createQueue),
	)
}
