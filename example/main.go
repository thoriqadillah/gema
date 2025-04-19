package main

import (
	"embed"
	"fmt"

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/thoriqadillah/gema"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

//go:embed templates
var template embed.FS

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
		gema.NotifierModule(
			gema.WithMailerName("gema"),
			gema.WithMailerSender("hello@gema.com"),
			gema.WithAppEnv(APP_ENV),
			gema.WithMailerTemplateFs(template, "templates/*.html"),
		),
		gema.StorageModule(
			gema.LocalStorage,
			gema.WithStorageUrlPath(fmt.Sprintf("http://localhost%s/storage", PORT)),
		),
		exampleModule,
		gema.Start(":8001"),
	)

	app.Run()
}
