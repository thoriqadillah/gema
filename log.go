package gema

import (
	"go.uber.org/fx"
	"go.uber.org/zap"
)

// LoggerModule provides a zap logger dependency
func LoggerModule(env string, options ...zap.Option) fx.Option {
	return fx.Module("logger", fx.Provide(
		func() *zap.Logger {
			var logger *zap.Logger
			var err error

			if env == "development" {
				logger, err = zap.NewDevelopment(options...)
			}

			logger, err = zap.NewProduction(options...)
			if err != nil {
				panic(err)
			}

			return logger
		},
	))
}
