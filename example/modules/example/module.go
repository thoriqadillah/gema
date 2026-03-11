package example

import (
	"github.com/thoriqadillah/gema"
	"go.uber.org/fx"
)

func NewModule() fx.Option {
	return fx.Module("example",
		fx.Provide(fx.Private, newStore),
		fx.Provide(newService),
		fx.Provide(gema.AsController(newController)),
		fx.Invoke(newWorker),
	)
}
