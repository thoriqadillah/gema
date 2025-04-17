package gema

import (
	"context"
	"fmt"
	"log"

	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
	"go.uber.org/fx"
	"go.uber.org/zap"
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

// Constructor is any function that accepts any number of arguments and returns any number of values without error
type Constructor interface{}

var controllers = make([]Controller, 0)

func registerController(c Controller) {
	controllers = append(controllers, c)
}

// RegisterController will register invoke the controller constructor
// and register the controller to the echo instance as well as any other providers
// that are passed in
func RegisterController(name string, controller Constructor) fx.Option {
	return fx.Module("controller."+name,
		fx.Provide(fx.Private, controller),
		fx.Invoke(registerController),
	)
}

func Start(port string) fx.Option {
	return fx.Module("start", fx.Invoke(
		func(lc fx.Lifecycle, e *echo.Echo, pool *pgxpool.Pool, logger *zap.Logger) {
			lc.Append(fx.Hook{
				OnStart: func(ctx context.Context) error {
					for _, controller := range controllers {
						name := fmt.Sprintf("%T", controller)
						fmt.Println("[Gema] Registering controller:", name)

						r := e.Group("")
						controller.CreateRoutes(r)

						if initter, ok := controller.(Initter); ok {
							if err := initter.Init(ctx); err != nil {
								return err
							}
						}
					}

					go func() {
						if err := e.Start(port); err != nil {
							log.Fatal(err)
						}
					}()

					return pool.Ping(ctx)
				},
				OnStop: func(ctx context.Context) error {
					defer logger.Sync()

					for _, controller := range controllers {
						if closer, ok := controller.(Closer); ok {
							if err := closer.Close(ctx); err != nil {
								return err
							}
						}
					}

					pool.Close()
					fmt.Println("[Gema] Database closed")
					return e.Shutdown(ctx)
				},
			})
		},
	))
}

// DecorateEcho will decorate echo instance with the custom json serializer
// and binder + struct validator with go-playground/validator
func DecorateEcho(customValidation ...map[string]validator.Func) fx.Option {
	validationMap := make(map[string]validator.Func)
	if len(customValidation) > 0 {
		validationMap = customValidation[0]
	}

	return fx.Module("echo",
		registerValidator(validationMap),
		fx.Invoke(registerCustomBinder),
		fx.Invoke(registerCustomSerializer),
	)
}
