package gema

import (
	"context"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"go.uber.org/fx"
)

type Controller interface {
	CreateRoutes(r *echo.Group)
}

func AsController(constructor any) any {
	return fx.Annotate(
		constructor,
		fx.As(new(Controller)),
		fx.ResultTags(`group:"controllers"`),
	)
}

type httpParams struct {
	fx.In

	*echo.Echo
	fx.Lifecycle
	Controllers []Controller `group:"controllers"`
}

// StartHTTP will start the echo server and register the controllers
// to the echo instance. It will also create custom binder for added validation
// and serializer for the echo instance
func StartHTTP(address string) fx.Option {
	return fx.Module("start_http",
		fx.Invoke(registerCustomBinder),
		fx.Invoke(registerCustomSerializer),
		fx.Invoke(func(p httpParams) {
			for _, controller := range p.Controllers {
				controller.CreateRoutes(p.Group(""))
			}

			p.Append(fx.Hook{
				OnStart: func(ctx context.Context) error {
					go func() {
						if err := p.Start(address); err != nil && err != http.ErrServerClosed {
							fmt.Println("[Gema] Http server stopped with error: ", err)
						}
					}()

					return nil
				},
				OnStop: p.Shutdown,
			})
		}),
	)
}
