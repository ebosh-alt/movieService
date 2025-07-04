// internal/app/app.go

package app

import (
	"context"

	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
	"go.uber.org/zap"

	"movieService/internal/config"
	"movieService/internal/delivery/http/middleware"
	"movieService/internal/delivery/http/server"
	"movieService/internal/repository/postgres"
	"movieService/internal/usecase"
	"movieService/pkg/jwt"
)

func New() *fx.App {
	return fx.New(
		// --- Provide all dependencies ---
		fx.Provide(
			// базовые
			context.Background,
			config.NewConfig,
			zap.NewDevelopment,

			// Postgres-репозиторий и его интерфейс
			postgres.NewRepository,
			func(r *postgres.Repository) postgres.InterfaceRepository {
				return r
			},

			// JWT-сервис из конфига
			func(cfg *config.Config) jwt.InterfaceJWT {
				return jwt.NewJWT(cfg.JWT.Secret, cfg.JWT.TTL)
			},

			// Usecase и его интерфейс
			usecase.NewUsecase,
			func(u *usecase.Usecase) usecase.InterfaceUsecase {
				return u
			},

			// HTTP-мiddleware и сервер
			middleware.NewMiddleware,
			server.NewServer,
		),
		// Lifecycle: сначала поднимаем репозиторий
		fx.Invoke(func(lc fx.Lifecycle, repo *postgres.Repository) {
			lc.Append(fx.Hook{
				OnStart: repo.OnStart,
				OnStop:  repo.OnStop,
			})
		}),
		// --- Hook server lifecycle ---
		fx.Invoke(func(lc fx.Lifecycle, srv *server.Server) {
			lc.Append(fx.Hook{
				OnStart: srv.OnStart,
				OnStop:  srv.OnStop,
			})
		}),

		// --- Use Zap logger for Fx events ---
		fx.WithLogger(func(log *zap.Logger) fxevent.Logger {
			return &fxevent.ZapLogger{Logger: log}
		}),
	)
}
