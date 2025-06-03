package app

import (
	"context"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
	"go.uber.org/zap"
	_ "os"
	_ "os/signal"
	_ "syscall"

	"movieService/internal/config"
	"movieService/internal/delivery/http/middleware"
	"movieService/internal/delivery/http/server"
	"movieService/internal/repository"
	"movieService/internal/usecase"
	"movieService/pkg/jwt"
)

func New() *fx.App {
	return fx.New(
		fx.Options(
			repository.New(),
			usecase.New(),
			middleware.New(),
			server.New(),
			jwt.New(),
		),
		fx.Provide(
			context.Background,
			config.NewConfig,
			zap.NewDevelopment,
		),
		fx.WithLogger(
			func(log *zap.Logger) fxevent.Logger {
				return &fxevent.ZapLogger{Logger: log}
			},
		),
	)
}
