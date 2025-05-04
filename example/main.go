package main

import (
	"embed"
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/riverqueue/river"
	"github.com/thoriqadillah/gema"
	"go.uber.org/fx"
)

//go:embed templates
var templateFs embed.FS

func httpServer() *echo.Echo {
	e := echo.New()
	e.Use(middleware.Gzip())
	e.Use(middleware.Recover())

	return e
}

func init() {
	godotenv.Load()
}

func main() {
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

	dbConfig, err := pgxpool.ParseConfig(DB_URL)
	if err != nil {
		panic(err)
	}

	queueDbConfig, err := pgxpool.ParseConfig(DB_QUEUE_URL)
	if err != nil {
		panic(err)
	}

	app := fx.New(
		gema.FxLogger,
		gema.LoggerModule(APP_ENV),
		fx.Provide(httpServer),
		gema.DatabaseModule(dbConfig),
		gema.TransactionalClsModule,
		gema.RiverQueueModule(queueDbConfig, &river.Config{
			Queues: map[string]river.QueueConfig{
				river.QueueDefault: {
					MaxWorkers: river.QueueNumWorkersMax,
				},
				gema.NotifierQueue: {
					MaxWorkers: 100,
				},
			},
		}),
		gema.NotifierModule(
			gema.EmailerProvider(&gema.EmailerOption{
				Env:        APP_ENV,
				TemplateFs: templateFs,
				Host:       MAILER_HOST,
				Port:       MAILER_PORT,
				Username:   MAILER_USER,
				Password:   MAILER_PASS,
				From:       MAILER_FROM,
				Name:       MAILER_NAME,
			}),
			gema.RiveredEmailProvider(&gema.EmailerOption{
				Env:        APP_ENV,
				TemplateFs: templateFs,
				Host:       MAILER_HOST,
				Port:       MAILER_PORT,
				Username:   MAILER_USER,
				Password:   MAILER_PASS,
				From:       MAILER_FROM,
				Name:       MAILER_NAME,
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
