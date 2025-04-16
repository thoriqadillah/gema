package main

import (
	"log"

	"github.com/labstack/echo/v4"
	"github.com/thoriqadillah/gema"
	"github.com/uptrace/bun"
	"go.uber.org/fx"
)

type exampleController struct{}

func newExampleController(db *bun.DB) gema.Controller {
	log.Println("database instance ping:", db.Ping())
	return &exampleController{}
}

func (e *exampleController) hello(ctx echo.Context) error {
	return ctx.String(200, "Hello, World!")
}

func (e *exampleController) CreateRoutes(r *echo.Group) {
	r.GET("/", e.hello)
}

func httpServer() *echo.Echo {
	e := echo.New()

	return e
}

func main() {
	app := fx.New(
		fx.Provide(httpServer),
		gema.DatabaseModule("postgres://packform:packform@localhost:5432/packform?sslmode=disable"),
		gema.StorageModule(gema.LocalStorage),
		gema.RegisterModule(newExampleController),
		gema.Start(":8001"),
	)

	app.Run()
}
