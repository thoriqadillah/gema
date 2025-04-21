package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
	"github.com/thoriqadillah/gema"
	"go.uber.org/fx"
)

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
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGABRT, syscall.SIGTERM)
	defer stop()

	godotenv.Load()

	app := fx.New(
		fx.NopLogger,
		gema.DatabaseModule(DB_URL),
		gema.CommandModule("Command line application",
			helloWorld,
			gema.SeederCommand(), // register your seeder here
		),
	)
	app.Start(ctx)
}
