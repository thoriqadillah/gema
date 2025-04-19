package gema

import (
	"github.com/spf13/cobra"
	"go.uber.org/fx"
)

type cmdParam struct {
	fx.In
	*cobra.Command `name:"root"`
}

// CommandConstructor is a function that accepts any number of providers
// and returns a *cobra.Command
type CommandConstructor interface{}

var root *cobra.Command

func CommandModule(desc string, cmds ...CommandConstructor) fx.Option {
	root = &cobra.Command{
		Use:   "cmd",
		Short: desc,
	}

	var startCmd = func(lc fx.Lifecycle) {
		lc.Append(fx.StartHook(root.Execute))
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
