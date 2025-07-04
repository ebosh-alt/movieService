package postgres

import (
	"context"
	_ "google.golang.org/protobuf/types/known/emptypb"
	"movieService/internal/entities"
)

type InterfaceRepository interface {
	ListMovies(ctx context.Context, request *entities.ListMoviesRequest) (*entities.ListMoviesResponse, error)
	GetMovie(ctx context.Context, movieID int) (*entities.Movie, error)
	CreateMovie(ctx context.Context, movie *entities.Movie, genreIDs []int) (*entities.Movie, error)
	DeleteMovie(ctx context.Context, movie *entities.Movie) error

	ListRatings(ctx context.Context, request *entities.ListRatingsRequest) (*entities.ListRatingsResponse, error)
	GetRating(ctx context.Context, MovieID int, RatingID int) (*entities.Rating, error)
	CreateRating(ctx context.Context, rating *entities.Rating) (*entities.Rating, error)
	DeleteRating(ctx context.Context, rating *entities.Rating) error

	ListComments(ctx context.Context, request *entities.ListCommentsRequest) (*entities.ListCommentsResponse, error)
	GetComment(ctx context.Context, MovieID int, CommentID int) (*entities.Comment, error)
	CreateComment(ctx context.Context, comment *entities.Comment) (*entities.Comment, error)
	DeleteComment(ctx context.Context, comment *entities.Comment) error
}
