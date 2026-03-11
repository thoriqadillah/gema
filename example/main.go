package main

import (
	"embed"
	"example/env"
	"example/modules/example"
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
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

func init() {
	godotenv.Load()
	env.Load()
}

func main() {
	registerValidation()

	app := fx.New(
		gema.FxLogger,
		gema.LoggerModule(env.APP_ENV),
		fx.Provide(httpServer),
		fx.Provide(grpcServer),
		gema.DatabaseModule(env.DB_URL),
		gema.NotifierModule(gema.EmailerProvider(emailConfig())),
		gema.StorageModule(gema.LocalStorageProvider(storageConfig())),
		gema.QueueClient(),
		gema.QueueServer(queueConfig()),
		gema.StartHTTP(fmt.Sprintf(":%d", env.PORT)),
		gema.StartGrpc("localhost", ":1234"),
		example.NewModule(),
	)

	app.Run()
}
