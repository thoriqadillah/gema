package gema

import (
	"fmt"
	"time"

	"github.com/labstack/echo/v4"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
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

func LoggerMiddleware(logger *zap.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()

			err := next(c)
			if err != nil {
				c.Error(err)
			}

			req := c.Request()
			res := c.Response()

			fields := []zapcore.Field{
				zap.String("remote_ip", c.RealIP()),
				zap.String("latency", time.Since(start).String()),
				zap.String("host", req.Host),
				zap.String("request", fmt.Sprintf("%s %s", req.Method, req.RequestURI)),
				zap.Int("status", res.Status),
				zap.Int64("size", res.Size),
				zap.String("user_agent", req.UserAgent()),
			}

			id := req.Header.Get(echo.HeaderXRequestID)
			if id == "" {
				id = res.Header().Get(echo.HeaderXRequestID)
			}
			fields = append(fields, zap.String("request_id", id))

			n := res.Status
			switch {
			case n >= 500:
				logger.With(zap.Error(err)).Error("Server error", fields...)
			case n >= 400:
				logger.With(zap.Error(err)).Warn("Client error", fields...)
			case n >= 300:
				logger.Info("Redirection", fields...)
			default:
				logger.Info("Success", fields...)
			}

			return nil
		}
	}
}
