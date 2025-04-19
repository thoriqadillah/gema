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
	store  Store
	mailer gema.Notifier
}

func newController(store Store, notifier gema.NotifierFacade) gema.Controller {
	return &exampleController{
		store:  store,
		mailer: notifier.Create(gema.EmailNotifier),
	}
}

func (e *exampleController) hello(c echo.Context) error {
	ctx := c.Request().Context()

	message := e.store.Hello(ctx)
	return c.String(200, message)
}

func (e *exampleController) notification(c echo.Context) error {
	ctx := c.Request().Context()

	err := e.mailer.Send(ctx, gema.Message{
		To:      []string{"hello@gema.com"},
		Subject: "Hello World",
		Body:    "example.html",
	})

	if err != nil {
		return err
	}

	return c.String(200, "email sent")
}

func (e *exampleController) transactional(c echo.Context) error {
	ctx := c.Request().Context()

	var message string
	err := gema.Transactional(ctx, func(ctx context.Context) error {
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
	r.GET("/notification", e.notification)
}
