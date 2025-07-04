package server

import (
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	_ "movieService/docs" // ← swagger.json и docs.go
)

func (s *Server) CreateController() {
	//url := ginSwagger.URL("/docs/swagger.json") // путь до вашего swagger.json
	s.serv.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	api := s.serv.Group("/api")
	{
		api.GET("/movies", s.ListMovies)
		api.GET("/movies/:id", s.GetMovie)
		api.POST("/movies", s.CreateMovie)
		api.DELETE("/movies/:id", s.DeleteMovie)

		api.GET("/movies/:id/ratings", s.ListRatings)
		api.GET("/movies/:id/ratings/:rid", s.GetRating)
		api.POST("/movies/:id/ratings", s.CreateRating)
		api.DELETE("/movies/:id/ratings/:rid", s.DeleteRating)

		api.GET("/movies/:id/comments", s.ListComments)
		api.GET("/movies/:id/comments/:cid", s.GetComment)
		api.POST("/movies/:id/comments", s.CreateComment)
		api.DELETE("/movies/:id/comments/:cid", s.DeleteComment)
	}
}
