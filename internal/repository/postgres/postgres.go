package postgres

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v4/pgxpool"
	"go.uber.org/zap"
	"movieService/internal/config"
)

type Repository struct {
	ctx context.Context
	log *zap.Logger
	cfg *config.Config
	DB  *pgxpool.Pool
}

func NewRepository(log *zap.Logger, cfg *config.Config, ctx context.Context) (*Repository, error) {
	return &Repository{
		ctx: ctx,
		log: log,
		cfg: cfg,
	}, nil
}

func (r *Repository) OnStart(_ context.Context) error {
	connectionUrl := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s", r.cfg.Postgres.Host, r.cfg.Postgres.Port, r.cfg.Postgres.User, r.cfg.Postgres.Password, r.cfg.Postgres.DBName, r.cfg.Postgres.SSLMode)

	r.log.Info(connectionUrl)
	pool, err := pgxpool.Connect(r.ctx, connectionUrl)
	if err != nil {
		return err
	}
	r.DB = pool
	return nil
}

func (r *Repository) OnStop(_ context.Context) error {
	r.DB.Close()
	return nil
}

const ()

//// Login проверяет email/password, генерирует новый refresh-токен и возвращает его.
//func (r *Repository) Login(ctx context.Context, user *entities.User) (*entities.RefreshToken, error) {
//	userDTO := user.ToDTO()
//	if err := r.DB.QueryRow(ctx, loginUserSQL, userDTO.Email, userDTO.Password).Scan(&userDTO.ID); err != nil {
//		r.log.Error("Login failed", zap.Error(err), zap.String("email", *userDTO.Email))
//		return nil, fmt.Errorf("invalid credentials: %w", err)
//	}
//
//	refreshToken := entities.RefreshTokenDTO{}
//	if err := r.DB.QueryRow(ctx, createTokenSQL, userDTO.ID).
//		Scan(&refreshToken.Token, &refreshToken.IssuedAt, &refreshToken.ExpiresAt); err != nil {
//		r.log.Error("Create refresh token failed", zap.Error(err), zap.Int("user_id", *userDTO.ID))
//		return nil, fmt.Errorf("create refresh token: %w", err)
//	}
//
//	return refreshToken.ToEntity(), nil
//}
//
//// CreateUserToken создаёт новый токен пользователю.
//func (r *Repository) CreateUserToken(ctx context.Context, token *entities.RefreshToken) (*entities.RefreshToken, error) {
//	tokenDTO := token.ToDTO()
//	// создаём новый токен
//	if err := r.DB.QueryRow(ctx, createTokenSQL, tokenDTO.UserID).
//		Scan(&tokenDTO.Token, &tokenDTO.IssuedAt, &tokenDTO.ExpiresAt); err != nil {
//		r.log.Error("Create new token failed", zap.Error(err), zap.Int("user_id", *tokenDTO.UserID))
//		return nil, fmt.Errorf("create new token: %w", err)
//	}
//
//	return tokenDTO.ToEntity(), nil
//}
//func (r *Repository) GetUserIDByToken(ctx context.Context, token string) (int, error) {
//	var userID int
//	if err := r.DB.QueryRow(ctx, selectUserIDSQL, token).Scan(&userID); err != nil {
//		r.log.Error("Refresh: invalid or expired token", zap.Error(err), zap.String("token", token))
//		return 0, fmt.Errorf("invalid refresh token: %w", err)
//	}
//	return userID, nil
//
//}
//
//// RevokeToken отзывает переданный refresh-токен.
//func (r *Repository) RevokeToken(ctx context.Context, token *entities.RefreshToken) error {
//	if _, err := r.DB.Exec(ctx, revokeTokenSQL, token.Token); err != nil {
//		r.log.Error("Logout failed", zap.Error(err), zap.String("token", token.Token))
//		return fmt.Errorf("logout: %w", err)
//	}
//	return nil
//}
//
//func (r *Repository) CreateUser(ctx context.Context, user *entities.User) (*entities.User, error) {
//	userDTO := user.ToDTO()
//	if err := r.DB.QueryRow(ctx, createUserSQL, userDTO.Email, userDTO.Password, userDTO.FullName, userDTO.Role, userDTO.Status).
//		Scan(&userDTO.ID); err != nil {
//		r.log.Error("Create new user failed", zap.Error(err))
//		return nil, fmt.Errorf("create new user: %w", err)
//	}
//	return userDTO.ToEntity(), nil
//}
//
//func (r *Repository) GetUser(ctx context.Context, user *entities.User) (*entities.User, error) {
//	userDTO := user.ToDTO()
//	if err := r.DB.QueryRow(ctx, getUserSQL, userDTO.ID).Scan(
//		&userDTO.Email,
//		&userDTO.FullName,
//		&userDTO.Status,
//		&userDTO.Role,
//		&userDTO.CreatedAt,
//		&userDTO.UpdatedAt,
//	); err != nil {
//		r.log.Error("Get user failed", zap.Error(err), zap.Int("user_id", *userDTO.ID))
//		return nil, fmt.Errorf(" Get user: %w", err)
//	}
//	return userDTO.ToEntity(), nil
//}
