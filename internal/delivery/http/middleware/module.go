package middleware

import (
	"go.uber.org/fx"
	"go.uber.org/zap"
)

func New() fx.Option {
	return fx.Module("Middleware",
		fx.Provide(
			NewMiddleware,
		),
		fx.Decorate(func(log *zap.Logger) *zap.Logger {
			return log.Named("middleware")
		}),
	)
}
