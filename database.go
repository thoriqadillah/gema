package gema

import (
	"context"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"go.uber.org/fx"
)

// DatabaseModule connect the database using bun with pgxpool
func DatabaseModule(dsn string) fx.Option {
	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		panic(err)
	}

	return DatabaseModuleWithOption(config)
}

func DatabaseModuleWithOption(config *pgxpool.Config) fx.Option {
	return fx.Module("database", fx.Provide(
		func() (*pgxpool.Pool, *bun.DB) {
			pool, err := pgxpool.NewWithConfig(context.Background(), config)
			if err != nil {
				log.Fatal(err)
			}

			sql := stdlib.OpenDBFromPool(pool)
			bundb := bun.NewDB(sql, pgdialect.New())

			return pool, bundb
		},
	))
}
