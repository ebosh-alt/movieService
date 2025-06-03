package repository

import (
	"go.uber.org/fx"
	"go.uber.org/zap"
	"movieService/internal/repository/postgres"
)

func New() fx.Option {
	return fx.Module("repository",
		fx.Provide(
			postgres.NewRepository,
		),
		fx.Invoke(
			func(lc fx.Lifecycle, pg *postgres.Repository) {
				lc.Append(fx.Hook{
					OnStart: pg.OnStart,
					OnStop:  pg.OnStop,
				})
			},
		),
		fx.Decorate(func(log *zap.Logger) *zap.Logger {
			return log.Named("repository")
		}),
	)
}
