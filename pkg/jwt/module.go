package jwt

import (
	"go.uber.org/fx"
	"go.uber.org/zap"
)

func New() fx.Option {
	return fx.Module(
		"jwt",
		fx.Provide(
			NewJWT,
		),
		fx.Invoke(
			func(lc fx.Lifecycle, j *ServiceJWT) {
				lc.Append(fx.Hook{
					OnStart: j.OnStart,
					OnStop:  j.OnStop,
				})
			},
		),
		fx.Decorate(func(log *zap.Logger) *zap.Logger {
			return log.Named("jwt")
		}),
	)
}
