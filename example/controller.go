package main

import (
	"context"

	"github.com/labstack/echo/v4"
	"github.com/thoriqadillah/gema"
	"go.uber.org/fx"
)

var exampleModule = fx.Module("example",
	fx.Provide(fx.Private, newStore),
	gema.RegisterController(newController),
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
	ctx := c.Request().Context()

	message := e.store.Hello(ctx)
	return c.String(200, message)
}

func (e *exampleController) transactional(c echo.Context) error {
	var message string
	err := gema.Transactional(c, func(ctx context.Context) error {
		message = e.store.Hello(ctx)
		return e.store.Foo(ctx)
	})

	if err != nil {
		return err
	}

	return c.String(200, message)
}

func (e *exampleController) CreateRoutes(r *echo.Group) {
	r.GET("/", e.hello)
	r.GET("/transactional", e.transactional)
}
