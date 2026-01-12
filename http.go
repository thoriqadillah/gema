package gema

import (
	"context"
	"fmt"

	"github.com/labstack/echo/v4"
	"go.uber.org/fx"
)

type Controller interface {
	CreateRoutes(r *echo.Group)
}

// ControllerConstructor is any function that accepts any number of arguments and returns `Controller` of values without error
type ControllerConstructor any

func registerController(e *echo.Echo, c Controller) {
	r := e.Group("")
	c.CreateRoutes(r)
}

// RegisterController will register invoke the controller constructor
// and register the controller to the echo instance as well as any other providers
// that are passed in
func RegisterController(constructors ...ControllerConstructor) fx.Option {
	options := make([]fx.Option, len(constructors))
	for i, c := range constructors {
		options[i] = fx.Module(
			fmt.Sprintf("controllers.%d", i),
			fx.Provide(fx.Private, c),
			fx.Invoke(registerController),
		)
	}

	return fx.Module("controller", options...)
}

// Start will start the echo server and register the controllers
// to the echo instance. It will also create custom binder for added validation
// and serializer for the echo instance
func Start(port string) fx.Option {
	return fx.Module("start",
		fx.Invoke(registerCustomBinder),
		fx.Invoke(registerCustomSerializer),
		fx.Invoke(func(lc fx.Lifecycle, e *echo.Echo) {
			lc.Append(fx.Hook{
				OnStart: func(ctx context.Context) error {
					go e.Start(port)
					return nil
				},
				OnStop: e.Shutdown,
			})
		}),
	)
}
