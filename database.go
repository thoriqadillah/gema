package gema

import (
	"context"
	"database/sql"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/spf13/cobra"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"go.uber.org/fx"
)

type DB struct {
	*bun.DB
}

type TxFunc = func(ctx context.Context) error

type contextKey struct{}

var txKey = contextKey{}

// TransactionFunc will propagate request scoped db transaction. If any error happens
// inside the transaction, it will rollback the the entire transaction.
// Use `db.Tx(ctx)` to get the propagated db instance.
func (t *DB) TransactionFunc(ctx context.Context, txFunc TxFunc, options ...*sql.TxOptions) error {
	option := &sql.TxOptions{}
	if len(options) > 0 {
		option = options[0]
	}

	tx, err := t.BeginTx(ctx, option)
	if err != nil {
		return err
	}

	ctx = context.WithValue(ctx, txKey, &tx)
	if err := txFunc(ctx); err != nil {
		return tx.Rollback()
	}

	return tx.Commit()
}

// Tx will return the propagated transaction instance.
// Can be used as a transaction if it were run inside a `TransactionFunc` function.
// Otherwise, it will return the default database instance
func (t *DB) Tx(ctx context.Context) bun.IDB {
	tx, ok := ctx.Value(txKey).(*bun.Tx)
	if !ok {
		return t.DB
	}

	return tx
}

// DatabaseModule connect the database using bun with pgxpool and provides the bun.DB instance.
func DatabaseModule(dbUrl string) fx.Option {
	return fx.Module("database", fx.Provide(
		func(lc fx.Lifecycle) (*pgxpool.Pool, *sql.DB, *DB) {
			pool, err := pgxpool.New(context.Background(), dbUrl)
			if err != nil {
				fmt.Println("[Gema] Failed to connect to database: ", err)
				os.Exit(1)
			}

			sqldb := stdlib.OpenDBFromPool(pool)
			bundb := bun.NewDB(sqldb, pgdialect.New())

			lc.Append(fx.Hook{
				OnStart: pool.Ping,
				OnStop: func(ctx context.Context) error {
					bundb.Close()
					pool.Close()

					return nil
				},
			})

			g := &DB{bundb}
			return pool, sqldb, g
		},
	))
}

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
