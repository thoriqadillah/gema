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
		fx.Provide(AsWorker(newWorker)),
		fx.Invoke(func(p workerParams) {
			for _, worker := range p.Registrar {
				worker.Register(p.Workers)
			}
		}),
	)
}
