package main

import (
	"cmdexample/env"
	"context"
	"embed"

	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
	"github.com/thoriqadillah/gema"
	"go.uber.org/fx"
)

//go:embed migrations
var migrationFs embed.FS

func helloWorld() *cobra.Command {
	return &cobra.Command{
		Use:   "hello",
		Short: "Hello world",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Println("Hello world")
		},
	}
}

func main() {
	godotenv.Load()
	env.Load()
	ctx := context.Background()

	app := fx.New(
		fx.NopLogger,
		gema.DatabaseModule(env.DB_URL),
		gema.CommandModule("Command line application",
			helloWorld,
			gema.MigrationCommand(migrationFs, "migrations"),
			gema.SeederCommand(), // register your seeder here
		),
	)

	app.Start(ctx)
	app.Stop(ctx)
	<-app.Done()
}
