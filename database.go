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
	return fx.Module("database", fx.Provide(
		func(lc fx.Lifecycle) (*pgxpool.Pool, *bun.DB) {
			pool, err := pgxpool.New(context.Background(), dsn)
			if err != nil {
				log.Fatal(err)
			}

			sql := stdlib.OpenDBFromPool(pool)
			bundb := bun.NewDB(sql, pgdialect.New())

			return pool, bundb
		},
	))
}

func DatabaseModuleWithOption(option *pgxpool.Config) fx.Option {
	return fx.Module("database", fx.Provide(
		func(lc fx.Lifecycle) (*pgxpool.Pool, *bun.DB) {
			pool, err := pgxpool.NewWithConfig(context.Background(), option)
			if err != nil {
				log.Fatal(err)
			}

			sql := stdlib.OpenDBFromPool(pool)
			bundb := bun.NewDB(sql, pgdialect.New())

			return pool, bundb
		},
	))
}
