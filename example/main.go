package main

import (
	"fmt"

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/thoriqadillah/gema"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

func httpServer(logger *zap.Logger) *echo.Echo {
	e := echo.New()
	e.Use(gema.LoggerMiddleware(logger))
	e.Use(middleware.Gzip())

	return e
}

func main() {
	godotenv.Load()

	app := fx.New(
		gema.LoggerModule(APP_ENV),
		fx.Provide(httpServer),
		gema.DecorateEcho(),
		gema.DatabaseModule(DB_URL),
		gema.StorageModule(
			gema.LocalStorage,
			gema.WithUrlPath(fmt.Sprintf("http://localhost%s/storage", PORT)),
		),
		exampleModule,
		gema.Start(":8001"),
	)

	app.Run()
}
