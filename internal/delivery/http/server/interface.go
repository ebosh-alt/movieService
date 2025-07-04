package server

import (
	"github.com/gin-gonic/gin"
)

// InterfaceServer описывает контракт HTTP-сервера.
// Включает запуск/остановку сервера и регистрацию HTTP-обработчиков для каждого usecase.
type InterfaceServer interface {
	ListMovies(c *gin.Context)
	GetMovie(c *gin.Context)
	CreateMovie(c *gin.Context)
	DeleteMovie(c *gin.Context)
	ListRatings(c *gin.Context)
	GetRating(c *gin.Context)
	CreateRating(c *gin.Context)
	DeleteRating(c *gin.Context)
	ListComments(c *gin.Context)
	GetComment(c *gin.Context)
	CreateComment(c *gin.Context)
	DeleteComment(c *gin.Context)
}
