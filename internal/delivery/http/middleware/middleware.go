package middleware

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"movieService/internal/config"
	"movieService/internal/repository/postgres"
	JWT "movieService/pkg/jwt"
	"net/http"
	"strings"

	_ "movieService/internal/config"
)

type Middleware struct {
	cfg   *config.Config
	repo  *postgres.Repository
	log   *zap.Logger
	roles map[string]int
	jwt   JWT.InterfaceJWT
}

var _ JWT.InterfaceJWT = (*JWT.ServiceJWT)(nil)

func NewMiddleware(cfg *config.Config, log *zap.Logger, repository *postgres.Repository, jwt JWT.InterfaceJWT) *Middleware {
	return &Middleware{
		cfg:  cfg,
		log:  log,
		repo: repository,
		jwt:  jwt,
	}
}

// Auth возвращает gin.HandlerFunc, проверяющий Bearer-JWT в заголовке Authorization.
func (m *Middleware) Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. Получаем заголовок Authorization
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is missing"})
			return
		}

		// 2. Должен быть формат "Bearer <token>"
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization header format"})
			return
		}
		tokenString := parts[1]

		// 3. Валидируем токен и получаем userID
		userID, err := m.jwt.Validate(tokenString)
		if err != nil {
			m.log.Error("Auth: invalid access token", zap.Error(err))
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			return
		}

		// 4. Кладём userID в контекст Gin, чтобы контроллеры могли его прочитать
		c.Set("userID", userID)

		c.Next()
	}
}
