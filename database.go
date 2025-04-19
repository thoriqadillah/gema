package gema

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/labstack/echo/v4"
	"github.com/spf13/cobra"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"go.uber.org/fx"
)

// DB is a wrapper for bun.DB to provide request scoped db transaction
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
		func(lc fx.Lifecycle) (*pgxpool.Pool, *DB) {
			fmt.Println("[Gema] Registering database module")

			pool, err := pgxpool.NewWithConfig(context.Background(), config)
			if err != nil {
				log.Fatal(err)
			}

			sql := stdlib.OpenDBFromPool(pool)
			bundb := bun.NewDB(sql, pgdialect.New())
			gemaDB := &DB{bundb}

			lc.Append(fx.Hook{
				OnStart: pool.Ping,
				OnStop: func(ctx context.Context) error {
					gemaDB.Close()
					pool.Close()

					fmt.Println("[Gema] Database connection closed")
					return nil
				},
			})
			return pool, gemaDB
		},
	))
}

func TransactionalCls() fx.Option {
	return fx.Invoke(func(db *DB, e *echo.Echo) {
		e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
			return func(c echo.Context) error {
				ctx := c.Request().Context()
				ctx = context.WithValue(ctx, "db", db.DB)

				c.SetRequest(c.Request().WithContext(ctx))
				return next(c)
			}
		})
	})
}

type TxFunc = func(ctx context.Context) error

// Transactional will propagate request scoped db transaction. If any error happens
// inside the transaction, it will rollback the the entire transaction
// If you are familiar with Nest js transactional cls, then this is kinda similar to that.
// Use `gema.DB.HostDB(ctx)` to get the propagated db instance.
func Transactional(ctx context.Context, txFunc TxFunc, options ...*sql.TxOptions) error {
	db := ctx.Value("db").(*bun.DB)

	option := &sql.TxOptions{}
	if len(options) > 0 {
		option = options[0]
	}

	tx, err := db.BeginTx(ctx, option)
	if err != nil {
		return err
	}

	ctx = context.WithValue(ctx, "tx", &tx)
	if err := txFunc(ctx); err != nil {
		return tx.Rollback()
	}

	return tx.Commit()
}

type Seeder interface {
	Seed(ctx context.Context, tx *bun.Tx) error
}

func SeederCommand(seeders ...Seeder) CommandConstructor {
	return func(db *DB) *cobra.Command {
		return &cobra.Command{
			Use:   "seed",
			Short: "Run the database seeder",
			RunE: func(cmd *cobra.Command, args []string) error {
				ctx := cmd.Context()

				tx, err := db.Begin()
				if err != nil {
					return err
				}

				for _, seeder := range seeders {
					if err := seeder.Seed(ctx, &tx); err != nil {
						return tx.Rollback()
					}
				}

				return tx.Commit()
			},
		}
	}
}
