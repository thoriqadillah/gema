package gema

import (
	"context"
	"database/sql"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/labstack/echo/v4"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"go.uber.org/fx"
)

type DB struct {
	*bun.DB
}

// HostDB will return the propagated database instance.
// Can be used as a transactional if it were run inside a `Transactional` function.
// Otherwise, it will return the default database instance
func (db *DB) HostDB(ctx context.Context) bun.IDB {
	tx, ok := ctx.Value("tx").(*bun.Tx)
	if !ok {
		return db.DB
	}

	return tx
}

// DatabaseModule connect the database using bun with pgxpool
// This module will also provide the database connection to the echo context
// to propagate request based db transaction
func DatabaseModule(dsn string) fx.Option {
	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		panic(err)
	}

	return DatabaseModuleWithOption(config)
}

func DatabaseModuleWithOption(config *pgxpool.Config) fx.Option {
	return fx.Module("database", fx.Provide(
		func(e *echo.Echo) (*pgxpool.Pool, *DB) {
			pool, err := pgxpool.NewWithConfig(context.Background(), config)
			if err != nil {
				log.Fatal(err)
			}

			sql := stdlib.OpenDBFromPool(pool)
			bundb := bun.NewDB(sql, pgdialect.New())
			gemaDB := &DB{bundb}

			e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
				return func(c echo.Context) error {
					c.Set("db", bundb)
					return next(c)
				}
			})

			return pool, gemaDB
		},
	))
}

type TxFunc = func(ctx context.Context) error

// Transactional will propagate request scoped db transaction.
// If you are familiar with Nest js transactional cls, then this is kinda similar to that.
// Use `gema.DB.HostDB(ctx)` to get the propagated db instance.
func Transactional(c echo.Context, txFunc TxFunc) error {
	ctx := c.Request().Context()

	db := c.Get("db").(*bun.DB)
	tx, err := db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return err
	}

	ctx = context.WithValue(ctx, "tx", &tx)
	if err := txFunc(ctx); err != nil {
		return tx.Rollback()
	}

	return tx.Commit()
}
