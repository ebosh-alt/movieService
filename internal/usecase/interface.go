// internal/usecase/interface.go

package usecase

import (
	"context"

	_ "google.golang.org/protobuf/types/known/emptypb"
	protos "movieService/pkg/proto/gen/go"
)

// InterfaceUsecase описывает бизнес-логику для работы с пользователями и аутентификацией.
type InterfaceUsecase interface {
	// Login выполняет аутентификацию по email и паролю.
	// Возвращает access и refresh токены с временем жизни.
	Login(ctx context.Context, req *protos.LoginRequest) (*protos.LoginResponse, error)

	// Refresh обновляет access и refresh токены по старому refresh-токену.
	Refresh(ctx context.Context, req *protos.RefreshRequest) (*protos.RefreshResponse, error)

	// Logout аннулирует переданный refresh-токен.
	Logout(ctx context.Context, req *protos.LogoutRequest) error

	// GetUser возвращает детальную информацию по одному пользователю.
	GetUser(ctx context.Context, req *protos.GetUserRequest) (*protos.GetUserResponse, error)

	// CreateUser создаёт нового пользователя или администратора.
	CreateUser(ctx context.Context, req *protos.CreateUserRequest) (*protos.CreateUserResponse, error)
}
