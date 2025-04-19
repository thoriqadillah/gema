package main

import (
	"embed"
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/riverqueue/river"
	"github.com/thoriqadillah/gema"
	"go.uber.org/fx"
)

//go:embed templates
var template embed.FS

func httpServer() *echo.Echo {
	e := echo.New()
	e.Use(middleware.Gzip())
	e.Use(middleware.Recover())

	return e
}

func main() {
	godotenv.Load()

	gema.RegisterValidator(map[string]validator.Func{
		"password": func(fl validator.FieldLevel) bool {
			password := fl.Field().String()
			if len(password) < 6 || len(password) > 16 {
				return false
			}
			hasNumber := false
			hasAlphabet := false
			hasSpecial := false
			for _, char := range password {
				switch {
				case char >= '0' && char <= '9':
					hasNumber = true
				case (char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z'):
					hasAlphabet = true
				case (char >= '!' && char <= '/') || (char >= ':' && char <= '@') || (char >= '[' && char <= '`') || (char >= '{' && char <= '~'):
					hasSpecial = true
				}
			}
			return hasNumber && hasAlphabet && hasSpecial
		},
	})

	app := fx.New(
		fx.NopLogger,
		gema.LoggerModule(APP_ENV),
		fx.Provide(httpServer),
		gema.DatabaseModule(DB_URL),
		gema.TransactionalCls(),
		gema.RiverQueueModule(&river.Config{
			Queues: map[string]river.QueueConfig{
				river.QueueDefault: {
					MaxWorkers: river.QueueNumWorkersMax,
				},
				"notification": {
					MaxWorkers: 100,
				},
			},
		}),
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
