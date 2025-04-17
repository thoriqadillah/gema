package main

import (
	"github.com/labstack/echo/v4"
	"github.com/thoriqadillah/gema"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

func httpServer(logger *zap.Logger) *echo.Echo {
	e := echo.New()

	return e
}

func main() {
	appEnv := gema.Env("APP_ENV").String()

	app := fx.New(
		gema.LoggerModule(appEnv),
		fx.Provide(httpServer),
		gema.DatabaseModule("postgres://packform:packform@localhost:5432/packform?sslmode=disable"),
		gema.StorageModule(gema.LocalStorage),
		gema.RegisterModule(
			newExampleController,
			newStore,
		),
		gema.Start(":8001"),
	)

	app.Run()
}
