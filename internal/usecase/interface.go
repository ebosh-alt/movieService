// internal/usecase/interface.go

package usecase

import (
	"context"
	"google.golang.org/protobuf/types/known/emptypb"

	_ "google.golang.org/protobuf/types/known/emptypb"
	protos "movieService/pkg/proto/gen/go"
)

// InterfaceUsecase описывает бизнес-логику для работы с пользователями и аутентификацией.
type InterfaceUsecase interface {
	// ListMovies возвращает постраничный список фильмов.
	//
	// Параметры:
	//   - ctx: контекст выполнения, поддерживает отмену и дедлайны.
	//   - req: DTO с параметрами пагинации и фильтрации по жанрам.
	//
	// Возвращает:
	//   - ListMoviesResponse: DTO со списком фильмов и общим количеством.
	//   - error: ошибку выполнения, если что-то пошло не так.
	ListMovies(ctx context.Context, req *protos.ListMoviesRequest) (*protos.ListMoviesResponse, error)

	// GetMovie возвращает подробную информацию о фильме по его ID.
	//
	// Параметры:
	//   - ctx: контекст выполнения.
	//   - req: DTO с идентификатором фильма.
	//
	// Возвращает:
	//   - Movie: DTO с деталями фильма.
	//   - error: ошибку, если фильм не найден или произошёл сбой БД.
	GetMovie(ctx context.Context, req *protos.GetMovieRequest) (*protos.Movie, error)

	// CreateMovie создаёт новый фильм в системе.
	//
	// Параметры:
	//   - ctx: контекст выполнения.
	//   - req: DTO с данными для создания (название, ссылки, описание, дата, длительность, жанры).
	//
	// Возвращает:
	//   - CreateMovieResponse: DTO с созданным фильмом.
	//   - error: ошибку валидации или сбой при записи в БД.
	CreateMovie(ctx context.Context, req *protos.CreateMovieRequest) (*protos.CreateMovieResponse, error)

	// DeleteMovie удаляет фильм по его ID.
	//
	// Параметры:
	//   - ctx: контекст выполнения.
	//   - req: DTO с идентификатором фильма для удаления.
	//
	// Возвращает:
	//   - Empty: пустой ответ при успешном удалении.
	//   - error: ошибку, если фильм не найден или сбой БД.
	DeleteMovie(ctx context.Context, req *protos.DeleteMovieRequest) (*emptypb.Empty, error)

	// --- Rating ---

	// ListRatings возвращает постраничный список оценок для указанного фильма.
	//
	// Параметры:
	//   - ctx: контекст выполнения.
	//   - req: DTO с ID фильма и параметрами пагинации.
	//
	// Возвращает:
	//   - ListRatingsResponse: DTO со списком оценок и общим количеством.
	//   - error: ошибку выполнения.
	ListRatings(ctx context.Context, req *protos.ListRatingsRequest) (*protos.ListRatingsResponse, error)

	// GetRating возвращает конкретную оценку по ID фильма и ID оценки.
	//
	// Параметры:
	//   - ctx: контекст выполнения.
	//   - req: DTO с идентификаторами фильма и оценки.
	//
	// Возвращает:
	//   - Rating: DTO с деталями оценки.
	//   - error: ошибку, если оценка не найдена или сбой БД.
	GetRating(ctx context.Context, req *protos.GetRatingRequest) (*protos.Rating, error)

	// CreateRating создаёт новую оценку для фильма.
	//
	// Параметры:
	//   - ctx: контекст выполнения.
	//   - req: DTO с данными для новой оценки (ID фильма, ID пользователя, значение от 1 до 10).
	//
	// Возвращает:
	//   - CreateRatingResponse: DTO с созданной оценкой.
	//   - error: ошибку валидации или записи в БД.
	CreateRating(ctx context.Context, req *protos.CreateRatingRequest) (*protos.CreateRatingResponse, error)

	// DeleteRating удаляет оценку по ID фильма и ID оценки.
	//
	// Параметры:
	//   - ctx: контекст выполнения.
	//   - req: DTO с идентификаторами фильма и оценки.
	//
	// Возвращает:
	//   - Empty: пустой ответ при успешном удалении.
	//   - error: ошибку, если оценка не найдена или сбой БД.
	DeleteRating(ctx context.Context, req *protos.DeleteRatingRequest) (*emptypb.Empty, error)

	// --- Comment ---

	// ListComments возвращает постраничный список комментариев к фильму.
	//
	// Параметры:
	//   - ctx: контекст выполнения.
	//   - req: DTO с ID фильма и параметрами пагинации.
	//
	// Возвращает:
	//   - ListCommentsResponse: DTO со списком комментариев и общим количеством.
	//   - error: ошибку выполнения.
	ListComments(ctx context.Context, req *protos.ListCommentsRequest) (*protos.ListCommentsResponse, error)

	// GetComment возвращает конкретный комментарий по ID фильма и ID комментария.
	//
	// Параметры:
	//   - ctx: контекст выполнения.
	//   - req: DTO с идентификаторами фильма и комментария.
	//
	// Возвращает:
	//   - Comment: DTO с текстом и метаданными комментария.
	//   - error: ошибку, если комментарий не найден или сбой БД.
	GetComment(ctx context.Context, req *protos.GetCommentRequest) (*protos.Comment, error)

	// CreateComment создаёт новый комментарий к фильму.
	//
	// Параметры:
	//   - ctx: контекст выполнения.
	//   - req: DTO с данными для нового комментария (ID фильма, ID пользователя, текст).
	//
	// Возвращает:
	//   - CreateCommentResponse: DTO с созданным комментарием.
	//   - error: ошибку валидации или записи в БД.
	CreateComment(ctx context.Context, req *protos.CreateCommentRequest) (*protos.CreateCommentResponse, error)

	// DeleteComment удаляет комментарий по ID фильма и ID комментария.
	//
	// Параметры:
	//   - ctx: контекст выполнения.
	//   - req: DTO с идентификаторами фильма и комментария.
	//
	// Возвращает:
	//   - Empty: пустой ответ при успешном удалении.
	//   - error: ошибку, если комментарий не найден или сбой БД.
	DeleteComment(ctx context.Context, req *protos.DeleteCommentRequest) (*emptypb.Empty, error)
}
