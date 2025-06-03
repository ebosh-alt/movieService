package server

import (
	"context"
	_ "fmt"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"movieService/internal/config"
	"movieService/internal/delivery/http/middleware"
	"movieService/internal/usecase"
	protos "movieService/pkg/proto/gen/go"
	"net/http"
	_ "net/http"
	"strconv"
)

type Server struct {
	log        *zap.Logger
	cfg        *config.Config
	serv       *gin.Engine
	Usecase    usecase.InterfaceUsecase
	middleware *middleware.Middleware
}

var _ usecase.InterfaceUsecase = (*usecase.Usecase)(nil)
var _ InterfaceServer = (*Server)(nil)

func NewServer(logger *zap.Logger, cfg *config.Config, uc usecase.InterfaceUsecase, middleware *middleware.Middleware) (*Server, error) {
	return &Server{
		log:        logger,
		cfg:        cfg,
		serv:       gin.Default(),
		Usecase:    uc,
		middleware: middleware,
	}, nil
}

func (s *Server) OnStart(_ context.Context) error {
	s.CreateController()
	go func() {
		s.log.Debug("server started")
		if err := s.serv.Run(s.cfg.Server.Host + ":" + s.cfg.Server.Port); err != nil {
			s.log.Error("failed to server: " + err.Error())
		}
		return
	}()
	return nil
}

func (s *Server) OnStop(_ context.Context) error {
	s.log.Debug("stop server")
	return nil
}

// Login обрабатывает POST /api/login и вызывает usecase.Login.
func (s *Server) Login(c *gin.Context) {
	var req protos.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		s.log.Error("Server.Login: не удалось спарсить JSON", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request payload"})
		return
	}

	// Вызываем бизнес-логику
	resp, err := s.Usecase.Login(c.Request.Context(), &req)
	if err != nil {
		s.log.Error("Server.Login: usecase.Login вернул ошибку", zap.Error(err))
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	// Отправляем JSON-ответ (protobuf-структура будет автоматически сериализована)
	c.JSON(http.StatusOK, resp)
}

// Refresh обрабатывает POST /api/refresh и вызывает usecase.Refresh.
func (s *Server) Refresh(c *gin.Context) {
	var req protos.RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		s.log.Error("Server.Refresh: не удалось спарсить JSON", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request payload"})
		return
	}

	resp, err := s.Usecase.Refresh(c.Request.Context(), &req)
	if err != nil {
		s.log.Error("Server.Refresh: usecase.Refresh вернул ошибку", zap.Error(err))
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// Logout обрабатывает POST /api/logout и вызывает usecase.Logout.
func (s *Server) Logout(c *gin.Context) {
	var req protos.LogoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		s.log.Error("Server.Logout: не удалось спарсить JSON", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request payload"})
		return
	}

	if err := s.Usecase.Logout(c.Request.Context(), &req); err != nil {
		s.log.Error("Server.Logout: usecase.Logout вернул ошибку", zap.Error(err))
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "logged out"})
}

// GetUser обрабатывает GET /api/user/:id и вызывает usecase.GetUser.
func (s *Server) GetUser(c *gin.Context) {
	idParam := c.Param("id")
	if idParam == "" {
		s.log.Error("Server.GetUser: не указан параметр id")
		c.JSON(http.StatusBadRequest, gin.H{"error": "user id is required"})
		return
	}

	idInt, err := strconv.Atoi(idParam)
	if err != nil {
		s.log.Error("Server.GetUser: не удалось конвертировать id в int", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	req := &protos.GetUserRequest{Id: int32(idInt)}
	resp, err := s.Usecase.GetUser(c.Request.Context(), req)
	if err != nil {
		s.log.Error("Server.GetUser: usecase.GetUser вернул ошибку", zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// CreateUser обрабатывает POST /api/user и вызывает usecase.CreateUser.
func (s *Server) CreateUser(c *gin.Context) {
	var req protos.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		s.log.Error("Server.CreateUser: не удалось спарсить JSON", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request payload"})
		return
	}

	resp, err := s.Usecase.CreateUser(c.Request.Context(), &req)
	if err != nil {
		s.log.Error("Server.CreateUser: usecase.CreateUser вернул ошибку", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, resp)
}
