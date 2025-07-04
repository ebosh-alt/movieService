package server

import (
	"context"
	_ "fmt"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/emptypb"
	"movieService/internal/config"
	"movieService/internal/delivery/http/middleware"
	"movieService/internal/usecase"
	protos "movieService/pkg/proto/gen/go"
	"net/http"
	"strconv"
	"strings"
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

// ListMovies godoc
// @Summary      Список фильмов
// @Description  Возвращает постраничный список фильмов с опциональным фильтром по жанрам.
// @Tags         movies
// @Accept       json
// @Produce      json
// @Param        page     query     int     false  "Номер страницы"        default(1)
// @Param        per_page query     int     false  "Элементов на страницу" default(10)
// @Param        genres   query     []int   false  "Фильтр по жанрам"     collectionFormat(csv)
// @Success      200      {object}  __.ListMoviesResponse
// @Failure 	 400 {object} errorResponse
// @Failure      500      {object}  errorResponse
// @Router       /movies [get]
func (s *Server) ListMovies(c *gin.Context) {
	// parse query params
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	per, _ := strconv.Atoi(c.DefaultQuery("per_page", "10"))
	genres := make([]int32, 0)
	if gs := c.Query("genres"); gs != "" {
		for _, part := range strings.Split(gs, ",") {
			if id, err := strconv.Atoi(strings.TrimSpace(part)); err == nil {
				genres = append(genres, int32(id))
			}
		}
	}

	req := &protos.ListMoviesRequest{
		Page:     int32(page),
		PerPage:  int32(per),
		GenreIds: genres,
	}
	resp, err := s.Usecase.ListMovies(c.Request.Context(), req)
	if err != nil {
		s.log.Error("ListMovies error", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

// GetMovie godoc
// @Summary      Получить фильм
// @Description  Возвращает подробную информацию о фильме по его ID.
// @Tags         movies
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "ID фильма"
// @Success      200  {object}  __.Movie
// @Failure      400  {object}  errorResponse
// @Failure      404  {object}  errorResponse
// @Router       /movies/{id} [get]
func (s *Server) GetMovie(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid movie id"})
		return
	}
	req := &protos.GetMovieRequest{Id: int32(id)}
	resp, err := s.Usecase.GetMovie(c.Request.Context(), req)
	if err != nil {
		s.log.Error("GetMovie error", zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

// CreateMovie godoc
// @Summary      Создать фильм
// @Description  Создаёт новый фильм в системе.
// @Tags         movies
// @Accept       json
// @Produce      json
// @Param        input  body      __.CreateMovieRequest  true  "Данные фильма"
// @Success      201    {object}  __.CreateMovieResponse
// @Failure      400    {object}  errorResponse
// @Router       /movies [post]
func (s *Server) CreateMovie(c *gin.Context) {
	var req protos.CreateMovieRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		s.log.Error("CreateMovie: invalid payload", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := s.Usecase.CreateMovie(c.Request.Context(), &req)
	if err != nil {
		s.log.Error("CreateMovie error", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, resp)
}

// DeleteMovie godoc
// @Summary      Удалить фильм
// @Description  Удаляет фильм по его ID.
// @Tags         movies
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "ID фильма"
// @Success      200  {object}  emptypb.Empty
// @Failure      400  {object}  errorResponse
// @Failure      404  {object}  errorResponse
// @Router       /movies/{id} [delete]
func (s *Server) DeleteMovie(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid movie id"})
		return
	}
	req := &protos.DeleteMovieRequest{Id: int32(id)}
	if _, err := s.Usecase.DeleteMovie(c.Request.Context(), req); err != nil {
		s.log.Error("DeleteMovie error", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, &emptypb.Empty{})
}

// ListRatings godoc
// @Summary      Список оценок
// @Description  Возвращает постраничный список оценок для указанного фильма.
// @Tags         ratings
// @Accept       json
// @Produce      json
// @Param        id       path      int  true  "ID фильма"
// @Param        page     query     int  false "Номер страницы"        default(1)
// @Param        per_page query     int  false "Элементов на страницу" default(10)
// @Success      200      {object}  __.ListRatingsResponse
// @Failure      400      {object}  errorResponse
// @Failure      500      {object}  errorResponse
// @Router       /movies/{id}/ratings [get]
func (s *Server) ListRatings(c *gin.Context) {
	mid, _ := strconv.Atoi(c.Param("id"))
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	per, _ := strconv.Atoi(c.DefaultQuery("per_page", "10"))
	req := &protos.ListRatingsRequest{
		MovieId: int32(mid),
		Page:    int32(page),
		PerPage: int32(per),
	}
	resp, err := s.Usecase.ListRatings(c.Request.Context(), req)
	if err != nil {
		s.log.Error("ListRatings error", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

// GetRating godoc
// @Summary      Получить оценку
// @Description  Возвращает конкретную оценку по ID фильма и ID оценки.
// @Tags         ratings
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "ID фильма"
// @Param        rid  path      int  true  "ID оценки"
// @Success      200  {object}  __.Rating
// @Failure      400  {object}  errorResponse
// @Failure      404  {object}  errorResponse
// @Router       /movies/{id}/ratings/{rid} [get]
func (s *Server) GetRating(c *gin.Context) {
	mid, _ := strconv.Atoi(c.Param("id"))
	rid, _ := strconv.Atoi(c.Param("rid"))
	req := &protos.GetRatingRequest{
		MovieId:  int32(mid),
		RatingId: int32(rid),
	}
	resp, err := s.Usecase.GetRating(c.Request.Context(), req)
	if err != nil {
		s.log.Error("GetRating error", zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

// CreateRating godoc
// @Summary      Создать оценку
// @Description  Создаёт новую оценку для фильма.
// @Tags         ratings
// @Accept       json
// @Produce      json
// @Param        input  body      __.CreateRatingRequest  true  "Данные оценки"
// @Success      201    {object}  __.CreateRatingResponse
// @Failure      400    {object}  errorResponse
// @Router       /movies/{id}/ratings [post]
func (s *Server) CreateRating(c *gin.Context) {
	mid, _ := strconv.Atoi(c.Param("id"))
	var req protos.CreateRatingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}
	req.MovieId = int32(mid)
	resp, err := s.Usecase.CreateRating(c.Request.Context(), &req)
	if err != nil {
		s.log.Error("CreateRating error", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, resp)
}

// DeleteRating godoc
// @Summary      Удалить оценку
// @Description  Удаляет оценку по ID фильма и ID оценки.
// @Tags         ratings
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "ID фильма"
// @Param        rid  path      int  true  "ID оценки"
// @Success      200  {object}  emptypb.Empty
// @Failure      400  {object}  errorResponse
// @Failure      404  {object}  errorResponse
// @Router       /movies/{id}/ratings/{rid} [delete]
func (s *Server) DeleteRating(c *gin.Context) {
	mid, _ := strconv.Atoi(c.Param("id"))
	rid, _ := strconv.Atoi(c.Param("rid"))
	req := &protos.DeleteRatingRequest{
		MovieId:  int32(mid),
		RatingId: int32(rid),
	}
	if _, err := s.Usecase.DeleteRating(c.Request.Context(), req); err != nil {
		s.log.Error("DeleteRating error", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, &emptypb.Empty{})
}

// ListComments godoc
// @Summary      Список комментариев
// @Description  Возвращает постраничный список комментариев к фильму.
// @Tags         comments
// @Accept       json
// @Produce      json
// @Param        id       path      int  true  "ID фильма"
// @Param        page     query     int  false "Номер страницы"        default(1)
// @Param        per_page query     int  false "Элементов на страницу" default(10)
// @Success      200      {object}  __.ListCommentsResponse
// @Failure      400      {object}  errorResponse
// @Failure      500      {object}  errorResponse
// @Router       /movies/{id}/comments [get]
func (s *Server) ListComments(c *gin.Context) {
	mid, _ := strconv.Atoi(c.Param("id"))
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	per, _ := strconv.Atoi(c.DefaultQuery("per_page", "10"))
	req := &protos.ListCommentsRequest{
		MovieId: int32(mid),
		Page:    int32(page),
		PerPage: int32(per),
	}
	resp, err := s.Usecase.ListComments(c.Request.Context(), req)
	if err != nil {
		s.log.Error("ListComments error", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

// GetComment godoc
// @Summary      Получить комментарий
// @Description  Возвращает конкретный комментарий по ID фильма и ID комментария.
// @Tags         comments
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "ID фильма"
// @Param        cid  path      int  true  "ID комментария"
// @Success      200  {object}  __.Comment
// @Failure      400  {object}  errorResponse
// @Failure      404  {object}  errorResponse
// @Router       /movies/{id}/comments/{cid} [get]
func (s *Server) GetComment(c *gin.Context) {
	mid, _ := strconv.Atoi(c.Param("id"))
	cid, _ := strconv.Atoi(c.Param("cid"))
	req := &protos.GetCommentRequest{
		MovieId:   int32(mid),
		CommentId: int32(cid),
	}
	resp, err := s.Usecase.GetComment(c.Request.Context(), req)
	if err != nil {
		s.log.Error("GetComment error", zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

// CreateComment godoc
// @Summary      Создать комментарий
// @Description  Создаёт новый комментарий к фильму.
// @Tags         comments
// @Accept       json
// @Produce      json
// @Param        input  body      __.CreateCommentRequest  true  "Данные комментария"
// @Success      201    {object}  __.CreateCommentResponse
// @Failure      400    {object}  errorResponse
// @Router       /movies/{id}/comments [post]
func (s *Server) CreateComment(c *gin.Context) {
	mid, _ := strconv.Atoi(c.Param("id"))
	var req protos.CreateCommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}
	req.MovieId = int32(mid)
	resp, err := s.Usecase.CreateComment(c.Request.Context(), &req)
	if err != nil {
		s.log.Error("CreateComment error", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, resp)
}

// DeleteComment godoc
// @Summary      Удалить комментарий
// @Description Удаляет комментарий по ID фильма и ID комментария.
// @Tags        comments
// @Accept      json
// @Produce     json
// @Param       id   path      int  true  "ID фильма"
// @Param       cid  path      int  true  "ID комментария"
// @Success     200  {object}  emptypb.Empty
// @Failure     400  {object}  errorResponse
// @Failure     404  {object}  errorResponse
// @Router      /movies/{id}/comments/{cid} [delete]
func (s *Server) DeleteComment(c *gin.Context) {
	mid, _ := strconv.Atoi(c.Param("id"))
	cid, _ := strconv.Atoi(c.Param("cid"))
	req := &protos.DeleteCommentRequest{
		MovieId:   int32(mid),
		CommentId: int32(cid),
	}
	if _, err := s.Usecase.DeleteComment(c.Request.Context(), req); err != nil {
		s.log.Error("DeleteComment error", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, &emptypb.Empty{})
}
