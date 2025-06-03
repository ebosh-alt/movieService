package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"go.uber.org/zap"

	"movieService/internal/config"
	"movieService/internal/entities"
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

const (
	listMoviesSQL          = `SELECT id, title, video_url, cover_url, description, release_date, duration_min, created_at, updated_at FROM movies ORDER BY id LIMIT $1 OFFSET $2`
	listMoviesByGenresSQL  = `SELECT m.id, m.title, m.video_url, m.cover_url, m.description, m.release_date, m.duration_min, m.created_at, m.updated_at FROM movies m WHERE m.id IN (SELECT movie_id FROM movie_genres WHERE genre_id = ANY($1)) ORDER BY m.id LIMIT $2 OFFSET $3`
	countMoviesSQL         = `SELECT COUNT(*) FROM movies`
	countMoviesByGenresSQL = `SELECT COUNT(DISTINCT m.id) FROM movies m JOIN movie_genres mg ON m.id = mg.movie_id WHERE mg.genre_id = ANY($1)`
	getMovieSQL            = `SELECT id, title, video_url, cover_url, description, release_date, duration_min, created_at, updated_at FROM movies WHERE id=$1`
	insertMovieSQL         = `INSERT INTO movies (title, video_url, cover_url, description, release_date, duration_min) VALUES ($1,$2,$3,$4,$5,$6) RETURNING id, created_at, updated_at`
	insertMovieGenreSQL    = `INSERT INTO movie_genres (movie_id, genre_id) VALUES ($1,$2)`
	deleteMovieSQL         = `DELETE FROM movies WHERE id=$1`
	deleteMovieGenresSQL   = `DELETE FROM movie_genres WHERE movie_id=$1`
	deleteMovieRatingsSQL  = `DELETE FROM ratings WHERE movie_id=$1`
	deleteMovieCommentsSQL = `DELETE FROM comments WHERE movie_id=$1`

	listRatingsSQL  = `SELECT id, movie_id, user_id, score, created_at, updated_at FROM ratings WHERE movie_id=$1 ORDER BY id LIMIT $2 OFFSET $3`
	countRatingsSQL = `SELECT COUNT(*) FROM ratings WHERE movie_id=$1`
	getRatingSQL    = `SELECT id, movie_id, user_id, score, created_at, updated_at FROM ratings WHERE movie_id=$1 AND id=$2`
	insertRatingSQL = `INSERT INTO ratings (movie_id, user_id, score) VALUES ($1,$2,$3) RETURNING id, created_at, updated_at`
	deleteRatingSQL = `DELETE FROM ratings WHERE movie_id=$1 AND id=$2`

	listCommentsSQL  = `SELECT id, movie_id, user_id, text, created_at, updated_at FROM comments WHERE movie_id=$1 ORDER BY id LIMIT $2 OFFSET $3`
	countCommentsSQL = `SELECT COUNT(*) FROM comments WHERE movie_id=$1`
	getCommentSQL    = `SELECT id, movie_id, user_id, text, created_at, updated_at FROM comments WHERE movie_id=$1 AND id=$2`
	insertCommentSQL = `INSERT INTO comments (movie_id, user_id, text) VALUES ($1,$2,$3) RETURNING id, created_at, updated_at`
	deleteCommentSQL = `DELETE FROM comments WHERE movie_id=$1 AND id=$2`
)

// ListMovies returns a list of movies with optional filtering by genres.
func (r *Repository) ListMovies(ctx context.Context, req *entities.ListMoviesRequest) (*entities.ListMoviesResponse, error) {
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PerPage <= 0 {
		req.PerPage = 10
	}
	offset := (req.Page - 1) * req.PerPage

	var rows pgx.Rows
	var err error
	if len(req.GenreIDs) > 0 {
		rows, err = r.DB.Query(ctx, listMoviesByGenresSQL, pgx.Array(req.GenreIDs), req.PerPage, offset)
	} else {
		rows, err = r.DB.Query(ctx, listMoviesSQL, req.PerPage, offset)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	movies := make([]*entities.Movie, 0)
	for rows.Next() {
		dto := &entities.MovieDTO{
			ID:          new(int),
			Title:       new(string),
			VideoURL:    new(string),
			CoverURL:    new(string),
			Description: new(string),
			ReleaseDate: new(time.Time),
			DurationMin: new(int),
			CreatedAt:   new(time.Time),
			UpdatedAt:   new(time.Time),
		}
		if err := rows.Scan(
			dto.ID,
			dto.Title,
			dto.VideoURL,
			dto.CoverURL,
			dto.Description,
			dto.ReleaseDate,
			dto.DurationMin,
			dto.CreatedAt,
			dto.UpdatedAt,
		); err != nil {
			return nil, err
		}
		movies = append(movies, dto.ToEntity())
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}

	var total int
	if len(req.GenreIDs) > 0 {
		err = r.DB.QueryRow(ctx, countMoviesByGenresSQL, pgx.Array(req.GenreIDs)).Scan(&total)
	} else {
		err = r.DB.QueryRow(ctx, countMoviesSQL).Scan(&total)
	}
	if err != nil {
		return nil, err
	}

	return &entities.ListMoviesResponse{Movies: movies, Total: total}, nil
}

// GetMovie returns movie by id.
func (r *Repository) GetMovie(ctx context.Context, movieID int) (*entities.Movie, error) {
	dto := &entities.MovieDTO{
		ID:          new(int),
		Title:       new(string),
		VideoURL:    new(string),
		CoverURL:    new(string),
		Description: new(string),
		ReleaseDate: new(time.Time),
		DurationMin: new(int),
		CreatedAt:   new(time.Time),
		UpdatedAt:   new(time.Time),
	}
	if err := r.DB.QueryRow(ctx, getMovieSQL, movieID).Scan(
		dto.ID,
		dto.Title,
		dto.VideoURL,
		dto.CoverURL,
		dto.Description,
		dto.ReleaseDate,
		dto.DurationMin,
		dto.CreatedAt,
		dto.UpdatedAt,
	); err != nil {
		return nil, err
	}
	return dto.ToEntity(), nil
}

// CreateMovie inserts new movie and related genres.
func (r *Repository) CreateMovie(ctx context.Context, movie *entities.Movie) (*entities.Movie, error) {
	tx, err := r.DB.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		} else {
			err = tx.Commit(ctx)
		}
	}()

	createdDTO := &entities.MovieDTO{
		ID:          new(int),
		CreatedAt:   new(time.Time),
		UpdatedAt:   new(time.Time),
		Title:       &movie.Title,
		VideoURL:    &movie.VideoURL,
		CoverURL:    &movie.CoverURL,
		Description: &movie.Description,
		ReleaseDate: &movie.ReleaseDate,
		DurationMin: &movie.DurationMin,
	}

	err = tx.QueryRow(ctx, insertMovieSQL,
		movie.Title,
		movie.VideoURL,
		movie.CoverURL,
		movie.Description,
		movie.ReleaseDate,
		movie.DurationMin,
	).Scan(createdDTO.ID, createdDTO.CreatedAt, createdDTO.UpdatedAt)
	if err != nil {
		return nil, err
	}

	created := createdDTO.ToEntity()

	// genres are stored separately when provided via DTO
	if dto, ok := interface{}(movie).(interface{ GetGenreIDs() []int }); ok {
		for _, gID := range dto.GetGenreIDs() {
			if _, err = tx.Exec(ctx, insertMovieGenreSQL, created.ID, gID); err != nil {
				return nil, err
			}
		}
	}

	return created, nil
}

// DeleteMovie removes movie and related entities.
func (r *Repository) DeleteMovie(ctx context.Context, movieID int) error {
	tx, err := r.DB.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		} else {
			err = tx.Commit(ctx)
		}
	}()

	if _, err = tx.Exec(ctx, deleteMovieGenresSQL, movieID); err != nil {
		return err
	}
	if _, err = tx.Exec(ctx, deleteMovieRatingsSQL, movieID); err != nil {
		return err
	}
	if _, err = tx.Exec(ctx, deleteMovieCommentsSQL, movieID); err != nil {
		return err
	}
	if _, err = tx.Exec(ctx, deleteMovieSQL, movieID); err != nil {
		return err
	}

	return nil
}

// ListRatings returns ratings for a movie with pagination.
func (r *Repository) ListRatings(ctx context.Context, req *entities.ListRatingsRequest) (*entities.ListRatingsResponse, error) {
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PerPage <= 0 {
		req.PerPage = 10
	}
	offset := (req.Page - 1) * req.PerPage

	rows, err := r.DB.Query(ctx, listRatingsSQL, req.MovieID, req.PerPage, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	ratings := make([]*entities.Rating, 0)
	for rows.Next() {
		dto := &entities.RatingDTO{
			ID:        new(int),
			MovieID:   new(int),
			UserID:    new(int),
			Score:     new(int),
			CreatedAt: new(time.Time),
			UpdatedAt: new(time.Time),
		}
		if err := rows.Scan(
			dto.ID,
			dto.MovieID,
			dto.UserID,
			dto.Score,
			dto.CreatedAt,
			dto.UpdatedAt,
		); err != nil {
			return nil, err
		}
		ratings = append(ratings, dto.ToEntity())
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}

	var total int
	if err := r.DB.QueryRow(ctx, countRatingsSQL, req.MovieID).Scan(&total); err != nil {
		return nil, err
	}

	return &entities.ListRatingsResponse{Ratings: ratings, Total: total}, nil
}

// GetRating returns rating by movie and rating id.
func (r *Repository) GetRating(ctx context.Context, movieID int, ratingID int) (*entities.Rating, error) {
	dto := &entities.RatingDTO{
		ID:        new(int),
		MovieID:   new(int),
		UserID:    new(int),
		Score:     new(int),
		CreatedAt: new(time.Time),
		UpdatedAt: new(time.Time),
	}
	if err := r.DB.QueryRow(ctx, getRatingSQL, movieID, ratingID).Scan(
		dto.ID,
		dto.MovieID,
		dto.UserID,
		dto.Score,
		dto.CreatedAt,
		dto.UpdatedAt,
	); err != nil {
		return nil, err
	}
	return dto.ToEntity(), nil
}

// CreateRating inserts new rating for movie.
func (r *Repository) CreateRating(ctx context.Context, rating *entities.Rating) (*entities.Rating, error) {
	dto := &entities.RatingDTO{
		ID:        new(int),
		CreatedAt: new(time.Time),
		UpdatedAt: new(time.Time),
		MovieID:   &rating.MovieID,
		UserID:    &rating.UserID,
		Score:     &rating.Score,
	}

	if err := r.DB.QueryRow(ctx, insertRatingSQL, rating.MovieID, rating.UserID, rating.Score).Scan(
		dto.ID,
		dto.CreatedAt,
		dto.UpdatedAt,
	); err != nil {
		return nil, err
	}

	return dto.ToEntity(), nil
}

// DeleteRating removes rating by id.
func (r *Repository) DeleteRating(ctx context.Context, movieID int, ratingID int) error {
	_, err := r.DB.Exec(ctx, deleteRatingSQL, movieID, ratingID)
	return err
}

// ListComments returns comments for a movie with pagination.
func (r *Repository) ListComments(ctx context.Context, req *entities.ListCommentsRequest) (*entities.ListCommentsResponse, error) {
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PerPage <= 0 {
		req.PerPage = 10
	}
	offset := (req.Page - 1) * req.PerPage

	rows, err := r.DB.Query(ctx, listCommentsSQL, req.MovieID, req.PerPage, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	comments := make([]*entities.Comment, 0)
	for rows.Next() {
		dto := &entities.CommentDTO{
			ID:        new(int),
			MovieID:   new(int),
			UserID:    new(int),
			Text:      new(string),
			CreatedAt: new(time.Time),
			UpdatedAt: new(time.Time),
		}
		if err := rows.Scan(
			dto.ID,
			dto.MovieID,
			dto.UserID,
			dto.Text,
			dto.CreatedAt,
			dto.UpdatedAt,
		); err != nil {
			return nil, err
		}
		comments = append(comments, dto.ToEntity())
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}

	var total int
	if err := r.DB.QueryRow(ctx, countCommentsSQL, req.MovieID).Scan(&total); err != nil {
		return nil, err
	}

	return &entities.ListCommentsResponse{Comments: comments, Total: total}, nil
}

// GetComment returns comment by movie and comment id.
func (r *Repository) GetComment(ctx context.Context, movieID int, commentID int) (*entities.Comment, error) {
	dto := &entities.CommentDTO{
		ID:        new(int),
		MovieID:   new(int),
		UserID:    new(int),
		Text:      new(string),
		CreatedAt: new(time.Time),
		UpdatedAt: new(time.Time),
	}
	if err := r.DB.QueryRow(ctx, getCommentSQL, movieID, commentID).Scan(
		dto.ID,
		dto.MovieID,
		dto.UserID,
		dto.Text,
		dto.CreatedAt,
		dto.UpdatedAt,
	); err != nil {
		return nil, err
	}
	return dto.ToEntity(), nil
}

// CreateComment inserts new comment for movie.
func (r *Repository) CreateComment(ctx context.Context, comment *entities.Comment) (*entities.Comment, error) {
	dto := &entities.CommentDTO{
		ID:        new(int),
		CreatedAt: new(time.Time),
		UpdatedAt: new(time.Time),
		MovieID:   &comment.MovieID,
		UserID:    &comment.UserID,
		Text:      &comment.Text,
	}

	if err := r.DB.QueryRow(ctx, insertCommentSQL, comment.MovieID, comment.UserID, comment.Text).Scan(
		dto.ID,
		dto.CreatedAt,
		dto.UpdatedAt,
	); err != nil {
		return nil, err
	}

	return dto.ToEntity(), nil
}

// DeleteComment removes comment by id.
func (r *Repository) DeleteComment(ctx context.Context, movieID int, commentID int) error {
	_, err := r.DB.Exec(ctx, deleteCommentSQL, movieID, commentID)
	return err
}

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
