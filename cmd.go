package gema

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
	"go.uber.org/fx"
)

// CommandConstructor is a function that accepts any number of providers
// and returns a *cobra.Command
type CommandConstructor any

var root *cobra.Command

// CommandModule is a module that registers your command to the root command.
func CommandModule(desc string, cmds ...CommandConstructor) fx.Option {
	root = &cobra.Command{
		Use:   "cmd",
		Short: desc,
	}

	var startCmd = func(lc fx.Lifecycle, sh fx.Shutdowner) {
		ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGABRT, syscall.SIGTERM)
		lc.Append(fx.StopHook(func() error {
			cancel()
			return sh.Shutdown()
		}))

		lc.Append(fx.StartHook(func() error {
			return root.ExecuteContext(ctx)
		}))
	}

	var registerCmd = func(c *cobra.Command) {
		root.AddCommand(c)
	}

	fxOptions := []fx.Option{fx.Invoke(startCmd)}
	for _, cmd := range cmds {
		fxOptions = append(fxOptions,
			fx.Module("command.registry",
				fx.Provide(fx.Private, cmd),
				fx.Invoke(registerCmd),
			),
		)
	}

	return fx.Module("root", fxOptions...)
}
