package main

import (
	"embed"
	"fmt"
	"html/template"

	"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/riverqueue/river"
	"github.com/thoriqadillah/gema"
	"go.uber.org/fx"
)

//go:embed templates
var templateFs embed.FS

func emailTemplate() *template.Template {
	tmpl, err := template.ParseFS(templateFs, "templates/*.html")
	if err != nil {
		panic(err)
	}

	return tmpl
}

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
		gema.FxLogger,
		gema.LoggerModule(APP_ENV),
		fx.Provide(httpServer),
		gema.DatabaseModule(DB_URL),
		gema.TransactionalClsModule,
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
			gema.EmailerProvider(&gema.EmailerOption{
				Env:      APP_ENV,
				Template: emailTemplate(),
			}),
			gema.RiveredEmailProvider(&gema.EmailerOption{
				Env:      APP_ENV,
				Template: emailTemplate(),
			}),
		),
		gema.StorageModule(
			gema.LocalStorageProvider(&gema.LocalStorageOption{
				TempDir:       "./storage",
				FullRoutePath: fmt.Sprintf("http://localhost%s/storage", PORT),
			}),
		),
		exampleModule,
		gema.Start(PORT),
	)

	app.Run()
}
