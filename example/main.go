package main

import (
	"example/env"

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/thoriqadillah/gema"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

func httpServer(logger *zap.Logger) *echo.Echo {
	e := echo.New()
	e.Use(gema.LoggerMiddleware(logger))

	return e
}

func main() {
	godotenv.Load()

	app := fx.New(
		gema.LoggerModule(env.APP_ENV),
		fx.Provide(httpServer),
		gema.DecorateEcho(),
		gema.DatabaseModule(env.DB_URL),
		gema.StorageModule(gema.LocalStorage),
		gema.RegisterModule(
			newController,
			newStore,
		),
		gema.Start(":8001"),
	)

	app.Run()
}
