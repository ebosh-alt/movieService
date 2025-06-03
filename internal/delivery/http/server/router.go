package server

import (
	_ "movieService/docs"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func (s *Server) CreateController() {
	s.serv.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	api := s.serv.Group("/api")

	// Маршруты без авторизации
	api.POST("/login", s.Login)
	api.POST("/refresh", s.Refresh)
	api.POST("/logout", s.Logout)

	// Маршруты с авторизацией (middleware.Auth() проверяет JWT-токен)
	protected := api.Group("", s.middleware.Auth())
	{
		protected.GET("/user/:id", s.GetUser)
		protected.POST("/user", s.CreateUser)
	}
}
