package usecase

import (
	"go.uber.org/fx"
	"go.uber.org/zap"
)

func New() fx.Option {
	return fx.Module(
		"usecase",
		fx.Provide(
			NewUsecase,
		),
		fx.Invoke(
			func(lc fx.Lifecycle, uc *Usecase) {
				lc.Append(fx.Hook{
					OnStart: uc.OnStart,
					OnStop:  uc.OnStop,
				})
			},
		),
		fx.Decorate(func(log *zap.Logger) *zap.Logger {
			return log.Named("usecase")
		}),
	)
}
