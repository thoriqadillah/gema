package main

import (
	"embed"
	"example/controller"
	"example/env"
	"example/service"
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/riverqueue/river"
	"github.com/thoriqadillah/gema"
	"go.uber.org/fx"
	"google.golang.org/grpc"
)

//go:embed templates
var templateFs embed.FS

func httpServer() *echo.Echo {
	e := echo.New()
	e.Use(middleware.Gzip())
	e.Use(middleware.Recover())

	return e
}

func grpcServer() *grpc.Server {
	return grpc.NewServer()
}

func registerValidation() {
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
}

func main() {
	godotenv.Load()
	env.Load()

	registerValidation()
	storageConfig := &gema.LocalStorageOption{
		TempDir:       "./storage",
		FullRoutePath: fmt.Sprintf("http://localhost%s/storage", env.PORT),
	}

	queueConfig := map[string]river.QueueConfig{
		river.QueueDefault: {
			MaxWorkers: 100,
		},
	}

	app := fx.New(
		gema.FxLogger,
		gema.LoggerModule(env.APP_ENV),
		fx.Provide(httpServer),
		fx.Provide(grpcServer),
		gema.DatabaseModule(env.DB_URL),
		gema.NotifierModule(
			gema.EmailerProvider(&gema.EmailerOption{
				Env:        env.APP_ENV,
				TemplateFs: templateFs,
				Host:       env.MAILER_HOST,
				Port:       env.MAILER_PORT,
				Username:   env.MAILER_USER,
				Password:   env.MAILER_PASS,
				From:       env.MAILER_FROM,
				Name:       env.MAILER_NAME,
			}),
		),
		gema.StorageModule(gema.LocalStorageProvider(storageConfig)),
		gema.QueueClient(),
		gema.QueueServer(queueConfig),
		gema.StartHTTP(env.PORT),
		gema.StartGRPC("localhost", ":1234"),
		service.NewExample(),
		gema.RegisterController(controller.NewController),
	)

	app.Run()
}
