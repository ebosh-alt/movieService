package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
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
	pool, err := pgxpool.New(r.ctx, connectionUrl)
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
	listMoviesSQL = `
SELECT
  m.id, m.title, m.video_url, m.cover_url, m.description,
  m.release_date, m.duration_min, m.created_at, m.updated_at,
  -- массив ID жанров
  COALESCE(array_agg(mg.genre_id ORDER BY mg.genre_id) FILTER (WHERE mg.genre_id IS NOT NULL), '{}')   AS genre_ids,
  -- массив имён жанров в том же порядке
  COALESCE(array_agg(g.name    ORDER BY mg.genre_id) FILTER (WHERE g.name    IS NOT NULL), '{}')   AS genre_names
FROM movies m
LEFT JOIN movie_genres mg ON m.id = mg.movie_id
LEFT JOIN genres        g  ON mg.genre_id = g.id
GROUP BY m.id
ORDER BY m.id
LIMIT $1 OFFSET $2;
`

	listMoviesByGenresSQL = `
SELECT
  m.id, m.title, m.video_url, m.cover_url, m.description,
  m.release_date, m.duration_min, m.created_at, m.updated_at,
  COALESCE(array_agg(mg2.genre_id ORDER BY mg2.genre_id) FILTER (WHERE mg2.genre_id IS NOT NULL), '{}') AS genre_ids,
  COALESCE(array_agg(g.name        ORDER BY mg2.genre_id) FILTER (WHERE g.name        IS NOT NULL), '{}') AS genre_names
FROM movies m
JOIN movie_genres mg ON m.id = mg.movie_id
LEFT JOIN movie_genres mg2 ON m.id = mg2.movie_id
LEFT JOIN genres        g   ON mg2.genre_id = g.id
WHERE m.id IN (SELECT movie_id FROM movie_genres WHERE genre_id = ANY($1))
GROUP BY m.id
ORDER BY m.id
LIMIT $2 OFFSET $3;
`
	countMoviesSQL         = `SELECT COUNT(*) FROM movies`
	countMoviesByGenresSQL = `SELECT COUNT(DISTINCT m.id) FROM movies m JOIN movie_genres mg ON m.id = mg.movie_id WHERE mg.genre_id = ANY($1)`
	getMovieSQL            = `
SELECT
  m.id, m.title, m.video_url, m.cover_url, m.description,
  m.release_date, m.duration_min, m.created_at, m.updated_at,
  COALESCE(array_agg(mg.genre_id ORDER BY mg.genre_id) FILTER (WHERE mg.genre_id IS NOT NULL), '{}')   AS genre_ids,
  COALESCE(array_agg(g.name       ORDER BY mg.genre_id) FILTER (WHERE g.name        IS NOT NULL), '{}')   AS genre_names
FROM movies m
LEFT JOIN movie_genres mg ON m.id = mg.movie_id
LEFT JOIN genres        g  ON mg.genre_id = g.id
WHERE m.id = $1
GROUP BY m.id;
`
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
func (r *Repository) ListMovies(ctx context.Context, request *entities.ListMoviesRequest) (*entities.ListMoviesResponse, error) {
	if request.Page <= 0 {
		request.Page = 1
	}
	if request.PerPage <= 0 {
		request.PerPage = 10
	}
	offset := (request.Page - 1) * request.PerPage
	var rows pgx.Rows
	var err error
	if len(request.GenreIDs) > 0 {
		rows, err = r.DB.Query(
			ctx,
			listMoviesByGenresSQL,
			request.GenreIDs,
			request.PerPage,
			offset,
		)
	} else {
		rows, err = r.DB.Query(ctx, listMoviesSQL, request.PerPage, offset)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	movies := make([]*entities.Movie, 0)
	for rows.Next() {
		movieDTO := &entities.MovieDTO{}
		if err := rows.Scan(
			&movieDTO.ID,
			&movieDTO.Title,
			&movieDTO.VideoURL,
			&movieDTO.CoverURL,
			&movieDTO.Description,
			&movieDTO.ReleaseDate,
			&movieDTO.DurationMin,
			&movieDTO.CreatedAt,
			&movieDTO.UpdatedAt,
			&movieDTO.GenreIDs,   // сканируем массив ID
			&movieDTO.GenreNames, // сканируем массив имён
		); err != nil {
			return nil, err
		}
		movies = append(movies, movieDTO.ToEntity())
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}

	var total int
	if len(request.GenreIDs) > 0 {
		err = r.DB.QueryRow(ctx, countMoviesByGenresSQL, request.GenreIDs).Scan(&total)
	} else {
		err = r.DB.QueryRow(ctx, countMoviesSQL).Scan(&total)
	}
	if err != nil {
		return nil, err
	}
	return &entities.ListMoviesResponse{Movies: movies, Total: total}, nil
}

func (r *Repository) GetMovie(ctx context.Context, movieID int) (*entities.Movie, error) {
	dto := &entities.MovieDTO{}
	// Сканируем в DTO все поля + два массива (genre_ids и genre_names)
	if err := r.DB.QueryRow(ctx, getMovieSQL, movieID).Scan(
		&dto.ID,
		&dto.Title,
		&dto.VideoURL,
		&dto.CoverURL,
		&dto.Description,
		&dto.ReleaseDate,
		&dto.DurationMin,
		&dto.CreatedAt,
		&dto.UpdatedAt,
		&dto.GenreIDs,   // []int
		&dto.GenreNames, // []string
	); err != nil {
		return nil, err
	}
	// Преобразуем DTO → Entity (внутри склеиваются ID+Name в []Genre)
	return dto.ToEntity(), nil
}

// CreateMovie inserts new movie and related genres.
func (r *Repository) CreateMovie(ctx context.Context, movie *entities.Movie, genreIDs []int) (*entities.Movie, error) {
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

	movieDTO := movie.ToDTO(genreIDs, nil)
	err = tx.QueryRow(ctx, insertMovieSQL,
		&movieDTO.Title,
		&movieDTO.VideoURL,
		&movieDTO.CoverURL,
		&movieDTO.Description,
		&movieDTO.ReleaseDate,
		&movieDTO.DurationMin,
	).Scan(movieDTO.ID, movieDTO.CreatedAt, movieDTO.UpdatedAt)
	if err != nil {
		return nil, err
	}

	movie = movieDTO.ToEntity()

	// genres are stored separately when provided via DTO
	if dto, ok := interface{}(movie).(interface{ GetGenreIDs() []int }); ok {
		for _, gID := range dto.GetGenreIDs() {
			if _, err = tx.Exec(ctx, insertMovieGenreSQL, movie.ID, gID); err != nil {
				return nil, err
			}
		}
	}

	return movie, nil
}

// DeleteMovie removes movie and related entities.
func (r *Repository) DeleteMovie(ctx context.Context, movie *entities.Movie) error {
	movieDTO := movie.ToDTO(make([]int, 0), nil)
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

	if _, err = tx.Exec(ctx, deleteMovieGenresSQL, movieDTO.ID); err != nil {
		return err
	}
	if _, err = tx.Exec(ctx, deleteMovieRatingsSQL, movieDTO.ID); err != nil {
		return err
	}
	if _, err = tx.Exec(ctx, deleteMovieCommentsSQL, movieDTO.ID); err != nil {
		return err
	}
	if _, err = tx.Exec(ctx, deleteMovieSQL, movieDTO.ID); err != nil {
		return err
	}

	return nil
}

// ListRatings returns ratings for a movie with pagination.
func (r *Repository) ListRatings(ctx context.Context, request *entities.ListRatingsRequest) (*entities.ListRatingsResponse, error) {
	if request.Page <= 0 {
		request.Page = 1
	}
	if request.PerPage <= 0 {
		request.PerPage = 10
	}
	offset := (request.Page - 1) * request.PerPage

	rows, err := r.DB.Query(ctx, listRatingsSQL, request.MovieID, request.PerPage, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	ratings := make([]*entities.Rating, 0)
	for rows.Next() {
		ratingDTO := &entities.RatingDTO{}
		if err := rows.Scan(
			&ratingDTO.ID,
			&ratingDTO.MovieID,
			&ratingDTO.UserID,
			&ratingDTO.Score,
			&ratingDTO.CreatedAt,
			&ratingDTO.UpdatedAt,
		); err != nil {
			return nil, err
		}
		ratings = append(ratings, ratingDTO.ToEntity())
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}

	var total int
	if err := r.DB.QueryRow(ctx, countRatingsSQL, request.MovieID).Scan(&total); err != nil {
		return nil, err
	}

	return &entities.ListRatingsResponse{Ratings: ratings, Total: total}, nil
}

// GetRating returns rating by movie and rating id.
func (r *Repository) GetRating(ctx context.Context, movieID int, ratingID int) (*entities.Rating, error) {
	ratingDTO := &entities.RatingDTO{}
	if err := r.DB.QueryRow(ctx, getRatingSQL, movieID, ratingID).Scan(
		&ratingDTO.ID,
		&ratingDTO.MovieID,
		&ratingDTO.UserID,
		&ratingDTO.Score,
		&ratingDTO.CreatedAt,
		&ratingDTO.UpdatedAt,
	); err != nil {
		return nil, err
	}
	return ratingDTO.ToEntity(), nil
}

// CreateRating inserts new rating for movie.
func (r *Repository) CreateRating(ctx context.Context, rating *entities.Rating) (*entities.Rating, error) {
	ratingDTO := rating.ToDTO()

	if err := r.DB.QueryRow(ctx, insertRatingSQL, rating.MovieID, rating.UserID, rating.Score).Scan(
		&ratingDTO.ID,
		&ratingDTO.CreatedAt,
		&ratingDTO.UpdatedAt,
	); err != nil {
		return nil, err
	}

	return ratingDTO.ToEntity(), nil
}

// DeleteRating removes rating by id.
func (r *Repository) DeleteRating(ctx context.Context, rating *entities.Rating) error {
	ratingDTO := rating.ToDTO()
	_, err := r.DB.Exec(ctx, deleteRatingSQL, ratingDTO.MovieID, ratingDTO.ID)
	return err
}

// ListComments returns comments for a movie with pagination.
func (r *Repository) ListComments(ctx context.Context, request *entities.ListCommentsRequest) (*entities.ListCommentsResponse, error) {
	if request.Page <= 0 {
		request.Page = 1
	}
	if request.PerPage <= 0 {
		request.PerPage = 10
	}
	offset := (request.Page - 1) * request.PerPage

	rows, err := r.DB.Query(ctx, listCommentsSQL, request.MovieID, request.PerPage, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	comments := make([]*entities.Comment, 0)
	for rows.Next() {
		commentDTO := &entities.CommentDTO{}
		if err := rows.Scan(
			&commentDTO.ID,
			&commentDTO.MovieID,
			&commentDTO.UserID,
			&commentDTO.Text,
			&commentDTO.CreatedAt,
			&commentDTO.UpdatedAt,
		); err != nil {
			return nil, err
		}
		comments = append(comments, commentDTO.ToEntity())
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}

	var total int
	if err := r.DB.QueryRow(ctx, countCommentsSQL, request.MovieID).Scan(&total); err != nil {
		return nil, err
	}

	return &entities.ListCommentsResponse{Comments: comments, Total: total}, nil
}

// GetComment returns comment by movie and comment id.
func (r *Repository) GetComment(ctx context.Context, movieID int, commentID int) (*entities.Comment, error) {
	commentDTO := &entities.CommentDTO{
		ID:        new(int),
		MovieID:   new(int),
		UserID:    new(int),
		Text:      new(string),
		CreatedAt: new(time.Time),
		UpdatedAt: new(time.Time),
	}
	if err := r.DB.QueryRow(ctx, getCommentSQL, movieID, commentID).Scan(
		&commentDTO.ID,
		&commentDTO.MovieID,
		&commentDTO.UserID,
		&commentDTO.Text,
		&commentDTO.CreatedAt,
		&commentDTO.UpdatedAt,
	); err != nil {
		return nil, err
	}
	return commentDTO.ToEntity(), nil
}

// CreateComment inserts new comment for movie.
func (r *Repository) CreateComment(ctx context.Context, comment *entities.Comment) (*entities.Comment, error) {
	commentDTO := comment.ToDTO()

	if err := r.DB.QueryRow(ctx, insertCommentSQL, comment.MovieID, comment.UserID, comment.Text).Scan(
		&commentDTO.ID,
		&commentDTO.CreatedAt,
		&commentDTO.UpdatedAt,
	); err != nil {
		return nil, err
	}

	return commentDTO.ToEntity(), nil
}

// DeleteComment removes comment by id.
func (r *Repository) DeleteComment(ctx context.Context, comment *entities.Comment) error {
	commentDTO := comment.ToDTO()
	_, err := r.DB.Exec(ctx, deleteCommentSQL, commentDTO.MovieID, commentDTO.ID)
	return err
}

var _ InterfaceRepository = (*Repository)(nil)
