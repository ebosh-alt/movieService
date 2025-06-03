package usecase

import (
	"context"
	"fmt"
	"google.golang.org/protobuf/types/known/timestamppb"
	"movieService/internal/config"
	"movieService/internal/entities"
	"movieService/internal/repository/postgres"
	"time"

	"go.uber.org/zap"
	JWT "movieService/pkg/jwt"
	protos "movieService/pkg/proto/gen/go"
)

var _ postgres.InterfaceRepository = (*postgres.Repository)(nil)
var _ JWT.InterfaceJWT = (*JWT.ServiceJWT)(nil)

type Usecase struct {
	cfg  *config.Config
	log  *zap.Logger
	repo postgres.InterfaceRepository
	ctx  context.Context
	jwt  JWT.InterfaceJWT
}

func NewUsecase(logger *zap.Logger, repo postgres.InterfaceRepository, cfg *config.Config, ctx context.Context, jwt JWT.InterfaceJWT,

) (*Usecase, error) {
	return &Usecase{
		cfg:  cfg,
		log:  logger,
		repo: repo,
		ctx:  ctx,
		jwt:  jwt,
	}, nil
}

func (uc *Usecase) OnStart(_ context.Context) error { return nil }
func (uc *Usecase) OnStop(_ context.Context) error  { return nil }

// Login реализует movieService.Login
func (uc *Usecase) Login(ctx context.Context, req *protos.LoginRequest) (*protos.LoginResponse, error) {
	// Собираем сущность для репозитория: Email + "PasswordHash" = сырой пароль
	user := &entities.User{
		Email:    req.GetEmail(),
		Password: req.GetPassword(),
	}

	// 1) Проверяем credentials и создаём refresh‐токен в БД
	token, err := uc.repo.Login(ctx, user)
	if err != nil {
		uc.log.Error("Login usecase failed", zap.Error(err), zap.String("email", req.Email))
		return nil, err
	}

	// 2) Генерируем access‐токен
	accessToken, err := uc.jwt.Generate(int32(token.UserID))
	if err != nil {
		uc.log.Error("create access token failed", zap.Error(err))
		return nil, err
	}

	// 3) Считаем expires_in в секундах
	expiresIn := token.ExpiresAt.Sub(time.Now()).Seconds()
	if expiresIn < 0 {
		expiresIn = 0
	}

	return &protos.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: token.Token,
		ExpiresIn:    int64(expiresIn),
	}, nil
}

// Refresh обновляет access и refresh токены по старому refresh-токену.
func (uc *Usecase) Refresh(ctx context.Context, req *protos.RefreshRequest) (*protos.RefreshResponse, error) {
	uc.log.Info("UseCase.Refresh: входной запрос", zap.String("old_refresh", req.GetRefreshToken()))

	// Сначала получаем ID пользователя по переданному refresh-токену
	userID, err := uc.repo.GetUserIDByToken(ctx, req.GetRefreshToken())
	if err != nil {
		uc.log.Error("UseCase.Refresh: не найден пользователь по refresh-токену", zap.Error(err))
		return nil, err
	}

	// Создаём новую запись RefreshToken для пользователя
	token, err := uc.repo.CreateUserToken(ctx, &entities.RefreshToken{UserID: userID})
	if err != nil {
		uc.log.Error("UseCase.Refresh: ошибка создания нового refresh-токена", zap.Error(err))
		return nil, err
	}

	// Генерируем новый access-токен
	accessToken, err := uc.jwt.Generate(int32(token.UserID))
	if err != nil {
		uc.log.Error("UseCase.Refresh: ошибка генерации access-токена", zap.Error(err))
		return nil, err
	}

	return &protos.RefreshResponse{
		AccessToken:  accessToken,
		RefreshToken: token.Token,
		ExpiresIn:    int64(time.Hour.Seconds()),
	}, nil
}

// Logout аннулирует переданный refresh-токен.
func (uc *Usecase) Logout(ctx context.Context, req *protos.LogoutRequest) error {
	uc.log.Info("UseCase.Logout: входной запрос", zap.String("refresh", req.GetRefreshToken()))

	// Маппим токен в сущность
	token := &entities.RefreshToken{
		Token: req.GetRefreshToken(),
	}

	if err := uc.repo.RevokeToken(ctx, token); err != nil {
		uc.log.Error("UseCase.Logout: ошибка реверсии токена", zap.Error(err))
		return err
	}
	return nil
}

// GetUser возвращает детальную информацию по одному пользователю.
func (uc *Usecase) GetUser(ctx context.Context, req *protos.GetUserRequest) (*protos.GetUserResponse, error) {
	uc.log.Info("UseCase.GetUser: входной запрос", zap.Int("id", int(req.GetId())))

	// 1. Маппим Protobuf → Entity
	user := &entities.User{
		ID: int(req.GetId()),
	}

	// 2. Получаем из репозитория все данные о пользователе
	user, err := uc.repo.GetUser(ctx, user)
	if err != nil {
		uc.log.Error("UseCase.GetUser: ошибка получения пользователя", zap.Error(err))
		return nil, err
	}

	// 3. Маппинг строковых полей role/status в enum-ы Protobuf
	var role protos.Role
	switch user.Role {
	case "user":
		role = protos.Role_USER
	case "admin":
		role = protos.Role_ADMIN
	default:
		role = protos.Role_ROLE_UNSPECIFIED
	}

	var status protos.Status
	switch user.Status {
	case "active":
		status = protos.Status_ACTIVE
	case "blocked":
		status = protos.Status_BLOCKED
	default:
		status = protos.Status_STATUS_UNSPECIFIED
	}

	// 4. Формируем Protobuf-ответ
	return &protos.GetUserResponse{
		User: &protos.User{
			Id:        int32(user.ID),
			Email:     user.Email,
			FullName:  user.FullName,
			Role:      role,
			Status:    status,
			CreatedAt: timestamppb.New(user.CreatedAt),
			UpdatedAt: timestamppb.New(user.UpdatedAt),
		},
	}, nil
}

// CreateUser создаёт нового пользователя или администратора.
func (uc *Usecase) CreateUser(ctx context.Context, req *protos.CreateUserRequest) (*protos.CreateUserResponse, error) {
	uc.log.Info("UseCase.CreateUser: входной запрос", zap.String("email", req.GetEmail()))
	var role string
	switch req.GetRole() {
	case protos.Role_ADMIN:
		role = "admin"
	case protos.Role_USER:
		role = "user"
	default:
		return nil, fmt.Errorf("не указана роль пользователя или роль не поддерживается: %v", req.GetRole())
	}
	// Маппим Protobuf → Entity
	userEntity := &entities.User{
		Email:    req.GetEmail(),
		Password: req.GetPassword(),
		FullName: req.GetFullName(),
		Role:     role,
	}

	user, err := uc.repo.CreateUser(ctx, userEntity)
	if err != nil {
		uc.log.Error("UseCase.CreateUser: ошибка создания пользователя", zap.Error(err))
		return nil, err
	}
	// 3. Маппинг строковых полей role/status в enum-ы Protobuf
	var roleUser protos.Role
	switch user.Role {
	case "user":
		roleUser = protos.Role_USER
	case "admin":
		roleUser = protos.Role_ADMIN
	default:
		roleUser = protos.Role_ROLE_UNSPECIFIED
	}

	var status protos.Status
	switch user.Status {
	case "active":
		status = protos.Status_ACTIVE
	case "blocked":
		status = protos.Status_BLOCKED
	default:
		status = protos.Status_STATUS_UNSPECIFIED
	}
	// Возвращаем созданного пользователя в Protobuf-формате
	return &protos.CreateUserResponse{
		User: &protos.User{
			Id:       int32(user.ID),
			Email:    user.Email,
			FullName: user.FullName,
			Role:     roleUser,
			Status:   status,
		},
	}, nil
}
