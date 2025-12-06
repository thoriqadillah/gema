package controller

import (
	"example/service"

	"github.com/labstack/echo/v4"
	"github.com/thoriqadillah/gema"
)

type exampleController struct {
	svc *service.ExampleService
}

func NewController(svc *service.ExampleService) gema.Controller {
	return &exampleController{
		svc: svc,
	}
}

func (e *exampleController) hello(c echo.Context) error {
	ctx := c.Request().Context()

	message := e.svc.Hello(ctx)
	return c.String(200, message)
}

func (e *exampleController) notification(c echo.Context) error {
	ctx := c.Request().Context()
	if err := e.svc.Notification(ctx); err != nil {
		return err
	}

	return c.String(200, "email sent")
}

func (e *exampleController) transactional(c echo.Context) error {
	ctx := c.Request().Context()
	message, err := e.svc.Transaction(ctx)
	if err != nil {
		return err
	}

	return c.String(200, message)
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

	url, err := e.svc.Upload(f, file.Filename)
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
