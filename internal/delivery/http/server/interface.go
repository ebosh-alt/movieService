package server

import (
	"github.com/gin-gonic/gin"
)

// InterfaceServer описывает контракт HTTP-сервера.
// Включает запуск/остановку сервера и регистрацию HTTP-обработчиков для каждого usecase.
type InterfaceServer interface {
	// CreateController регистрирует все HTTP-маршруты, соответствующие usecase-методам.
	CreateController()

	// HTTP-обработчики, в которых будет вызываться логика из usecase.
	// Каждый метод получает *gin.Context, разбирает Protobuf/JSON-запрос,
	// вызывает соответствующий UseCase и возвращает Protobuf/JSON-ответ.

	// Login обрабатывает POST /login → usecase.Login
	Login(c *gin.Context)

	// Refresh обрабатывает POST /refresh → usecase.Refresh
	Refresh(c *gin.Context)

	// Logout обрабатывает POST /logout → usecase.Logout
	Logout(c *gin.Context)

	// GetUser обрабатывает GET /user/:id → usecase.GetUser
	GetUser(c *gin.Context)

	// CreateUser обрабатывает POST /user → usecase.CreateUser
	CreateUser(c *gin.Context)
}
