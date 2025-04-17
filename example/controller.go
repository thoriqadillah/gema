package main

import (
	"github.com/labstack/echo/v4"
	"github.com/thoriqadillah/gema"
)

type exampleController struct {
	store Store
}

func newController(store Store) gema.Controller {
	return &exampleController{
		store: store,
	}
}

func (e *exampleController) hello(ctx echo.Context) error {
	message := e.store.Hello()
	return ctx.String(200, message)
}

func (e *exampleController) CreateRoutes(r *echo.Group) {
	r.GET("/", e.hello)
}
