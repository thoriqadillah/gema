package main

import (
	"github.com/labstack/echo/v4"
	"github.com/thoriqadillah/gema"
	"go.uber.org/fx"
)

var exampleModule = fx.Module("example",
	fx.Provide(fx.Private, newStore),
	gema.RegisterController("example", newController),
)

type exampleController struct {
	store Store
}

func newController(store Store) gema.Controller {
	return &exampleController{
		store: store,
	}
}

func (e *exampleController) hello(c echo.Context) error {
	message := e.store.Hello()
	return c.String(200, message)
}

func (e *exampleController) CreateRoutes(r *echo.Group) {
	r.GET("/", e.hello)
}
