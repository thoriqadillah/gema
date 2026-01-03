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

type Initter interface {
	Init(ctx context.Context) error
}

type Closer interface {
	Close(ctx context.Context) error
}

// ControllerConstructor is any function that accepts any number of arguments and returns `Controller` of values without error
type ControllerConstructor any

var controllers = make([]Controller, 0)

func registerController(c ...Controller) {
	controllers = append(controllers, c...)
}

// RegisterController will register invoke the controller constructor
// and register the controller to the echo instance as well as any other providers
// that are passed in
func RegisterController(controller ...ControllerConstructor) fx.Option {
	return fx.Module("controller",
		fx.Provide(fx.Private, controller),
		fx.Invoke(registerController),
	)
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
					for _, controller := range controllers {
						r := e.Group("")
						controller.CreateRoutes(r)

						if initter, ok := controller.(Initter); ok {
							if err := initter.Init(ctx); err != nil {
								return err
							}
						}

						name := fmt.Sprintf("%T", controller)
						fmt.Println("[Gema] Controller registered:", name)
					}

					go e.Start(port)
					return nil
				},
				OnStop: func(ctx context.Context) error {
					for _, controller := range controllers {
						if closer, ok := controller.(Closer); ok {
							if err := closer.Close(ctx); err != nil {
								return err
							}
						}
					}

					return e.Shutdown(ctx)
				},
			})
		}),
	)
}
