package gema

import (
	"database/sql"
	"io/fs"

	"github.com/pressly/goose/v3"
	"github.com/spf13/cobra"
	"github.com/uptrace/bun"
)

func MigrationCommand(fs fs.FS, dir string) CommandConstructor {
	goose.SetBaseFS(fs)
	goose.SetDialect("postgres")

	return func(db *bun.DB) *cobra.Command {
		sqldb := db.DB
		migrationCmd := &cobra.Command{
			Use:   "migrate",
			Short: "Run database migration",
		}

		migrationCmd.AddCommand(
			upCmd(sqldb, dir),
			downCmd(sqldb, dir),
			versionCmd(sqldb, dir),
			createCmd(sqldb, dir),
			resetCmd(sqldb, dir),
		)

		return migrationCmd
	}
}

func upCmd(sqldb *sql.DB, dir string) *cobra.Command {
	upCmd := &cobra.Command{
		Use:     "up",
		Short:   "Migrate the database migration",
		Example: "  migrate up\n" + "  migrate up <migration_name>",
		Args:    cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			if len(args) > 0 {
				version := args[0]
				return goose.UpToContext(ctx, sqldb, dir, int64(ParseString(version).Int()))
			}

			return goose.UpContext(ctx, sqldb, dir)
		},
	}

	return upCmd
}

func downCmd(sqldb *sql.DB, dir string) *cobra.Command {
	down := &cobra.Command{
		Use:     "down",
		Short:   "Rollback database migration",
		Example: "  migrate down\n" + "  migrate <migration_name> <version>",
		Args:    cobra.MaximumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			migrationName := args[0]
			version := args[1]
			if migrationName == "" {
				return goose.DownContext(ctx, sqldb, dir)
			}

			return goose.DownToContext(ctx, sqldb, dir, int64(ParseString(version).Int()))
		},
	}

	return down
}

func versionCmd(sqldb *sql.DB, dir string) *cobra.Command {
	version := &cobra.Command{
		Use:     "version",
		Short:   "Show the current migration version",
		Example: "  migrate version",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			return goose.VersionContext(ctx, sqldb, dir)
		},
	}

	return version
}

func createCmd(sqldb *sql.DB, dir string) *cobra.Command {
	create := &cobra.Command{
		Use:     "create",
		Short:   "Create a new migration",
		Example: "  migrate create <migration_name>",
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			return goose.Create(sqldb, dir, name, "sql")
		},
	}

	return create
}

func resetCmd(sqldb *sql.DB, dir string) *cobra.Command {
	create := &cobra.Command{
		Use:     "reset",
		Short:   "Rollback all migrations",
		Example: "  migrate reset",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return goose.ResetContext(cmd.Context(), sqldb, dir)
		},
	}

	return create
}
