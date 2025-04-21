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
		func(lc fx.Lifecycle) (*pgxpool.Pool, *bun.DB) {
			fmt.Println("[Gema] Registering database module")

			pool, err := pgxpool.NewWithConfig(context.Background(), config)
			if err != nil {
				log.Fatal(err)
			}

			sqldb := stdlib.OpenDBFromPool(pool)
			bundb := bun.NewDB(sqldb, pgdialect.New())

			lc.Append(fx.Hook{
				OnStart: pool.Ping,
				OnStop: func(ctx context.Context) error {
					bundb.Close()
					pool.Close()

					fmt.Println("[Gema] Database connection closed")
					return nil
				},
			})

			return pool, bundb
		},
	))
}

type TransactionalCls struct {
	db *bun.DB
}

type TxFunc = func(ctx context.Context) error

// Transactional will propagate request scoped db transaction. If any error happens
// inside the transaction, it will rollback the the entire transaction
// If you are familiar with Nest js transactional cls, then this is kinda similar to that.
// Use `txHost.Tx(ctx)` to get the propagated db instance.
func (t *TransactionalCls) Transactional(ctx context.Context, txFunc TxFunc, options ...*sql.TxOptions) error {
	option := &sql.TxOptions{}
	if len(options) > 0 {
		option = options[0]
	}

	tx, err := t.db.BeginTx(ctx, option)
	if err != nil {
		return err
	}

	ctx = context.WithValue(ctx, "tx", &tx)
	if err := txFunc(ctx); err != nil {
		return tx.Rollback()
	}

	return tx.Commit()
}

// TransactionalMiddleware will create a middleware that will propagate request scoped db transaction.
// It is the same as `Transactional` but it will create a new transaction for each request.
// Use this middleware if you want to use the transaction in a cleaner way inside the handler
func (t *TransactionalCls) TransactionalMiddleware(opts ...*sql.TxOptions) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			ctx := c.Request().Context()

			option := &sql.TxOptions{}
			if len(opts) > 0 {
				option = opts[0]
			}

			tx, err := t.db.BeginTx(ctx, option)
			if err != nil {
				return err
			}

			ctx = context.WithValue(ctx, "tx", &tx)
			c.SetRequest(c.Request().WithContext(ctx))

			if err := next(c); err != nil {
				_ = tx.Rollback()
				return err
			}

			return tx.Commit()
		}
	}
}

func transactionalCls(db *bun.DB) *TransactionalCls {
	return &TransactionalCls{
		db: db,
	}
}

type TransactionalHost struct {
	db *bun.DB
}

// Tx will return the propagated transaction instance.
// Can be used as a transactional if it were run inside a `Transactional` function.
// Otherwise, it will return the default database instance
func (t *TransactionalHost) Tx(ctx context.Context) bun.IDB {
	tx, ok := ctx.Value("tx").(*bun.Tx)
	if !ok {
		return t.db
	}

	return tx
}

func txHost(db *bun.DB) *TransactionalHost {
	return &TransactionalHost{
		db: db,
	}
}

var TransactionalClsModule = fx.Module("transactional-cls",
	fx.Provide(transactionalCls),
	fx.Provide(txHost),
)

type Seeder interface {
	Seed(ctx context.Context, tx *bun.Tx) error
}

func SeederCommand(seeders ...Seeder) CommandConstructor {
	return func(db *bun.DB) *cobra.Command {
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
