package main

import (
	"context"
	"path/filepath"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/thoriqadillah/gema"
	"go.uber.org/fx"
)

var exampleModule = fx.Module("example",
	fx.Provide(fx.Private, newStore),
	gema.RegisterController(newController),
)

type exampleController struct {
	store   Store
	mailer  gema.Notifier
	cls     *gema.TransactionalCls
	storage gema.Storage
}

func newController(
	store Store,
	cls *gema.TransactionalCls,
	storageFactory gema.StorageFactory,
	notifierFactory gema.NotifierFactory,
) gema.Controller {
	return &exampleController{
		cls:     cls,
		store:   store,
		storage: storageFactory.Disk(gema.LocalStorage),
		mailer:  notifierFactory.Create(gema.RiveredEmailNotifier),
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
		To:       []string{"hello@gema.com"},
		Subject:  "Hello World",
		Template: "example.html",
	})

	if err != nil {
		return err
	}

	return c.String(200, "email sent")
}

func (e *exampleController) transactional(c echo.Context) error {
	ctx := c.Request().Context()

	return e.cls.Transactional(ctx, func(ctx context.Context) error {
		message := e.store.Hello(ctx)
		if err := e.store.Foo(ctx); err != nil {
			return err
		}

		return c.String(200, message)
	})
}

func (e *exampleController) validate(c echo.Context) error {
	var foo foo
	if err := c.Bind(&foo); err != nil {
		return err
	}

	return c.JSON(200, foo)
}

func (e *exampleController) upload(c echo.Context) error {
	file, err := c.FormFile("file")
	if err != nil {
		return err
	}

	f, err := file.Open()
	if err != nil {
		return err
	}

	defer f.Close()

	ext := filepath.Ext(file.Filename)
	id := uuid.NewString()
	filename := id + ext

	url, err := e.storage.Upload(filename, f)
	if err != nil {
		return err
	}

	return c.JSON(200, echo.Map{
		"url": url,
	})
}

func (e *exampleController) CreateRoutes(r *echo.Group) {
	r.GET("/", e.hello)
	r.POST("/validate", e.validate)
	r.POST("/upload", e.upload)
	r.GET("/transactional", e.transactional)
	r.GET("/notification", e.notification)
}
