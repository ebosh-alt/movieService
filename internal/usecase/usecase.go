package usecase

import (
	"context"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"movieService/internal/config"
	"movieService/internal/entities"
	"movieService/internal/repository/postgres"
	JWT "movieService/pkg/jwt"
	protos "movieService/pkg/proto/gen/go"
)

var _ postgres.InterfaceRepository = (*postgres.Repository)(nil)
var _ JWT.InterfaceJWT = (*JWT.ServiceJWT)(nil)

type Usecase struct {
	cfg  *config.Config
	log  *zap.Logger
	repo postgres.InterfaceRepository
	ctx  context.Context
	jwt  JWT.InterfaceJWT
}

func NewUsecase(logger *zap.Logger, repo postgres.InterfaceRepository, cfg *config.Config, ctx context.Context, jwt JWT.InterfaceJWT,

) (*Usecase, error) {
	return &Usecase{
		cfg:  cfg,
		log:  logger,
		repo: repo,
		ctx:  ctx,
		jwt:  jwt,
	}, nil
}

func (uc *Usecase) OnStart(_ context.Context) error { return nil }
func (uc *Usecase) OnStop(_ context.Context) error  { return nil }

// ListMovies возвращает постраничный список фильмов.
//
// Параметры:
//   - ctx: контекст выполнения, поддерживает отмену и дедлайны.
//   - req: DTO с параметрами пагинации и фильтрации по жанрам.
//
// Возвращает:
//   - ListMoviesResponse: DTO со списком фильмов и общим количеством.
//   - error: ошибку выполнения, если что-то пошло не так.
func (uc *Usecase) ListMovies(ctx context.Context, req *protos.ListMoviesRequest) (*protos.ListMoviesResponse, error) {
	// Логируем входные параметры
	uc.log.Info("Usecase.ListMovies: входной запрос",
		zap.Int32("page", req.GetPage()),
		zap.Int32("per_page", req.GetPerPage()),
		zap.Any("genre_ids", req.GetGenreIds()),
	)
	// 1. Маппим Protobuf → Entity
	// Преобразуем page/per_page и genre_ids из int32 в int
	listMoviesRq := &entities.ListMoviesRequest{
		Page:     int(req.GetPage()),
		PerPage:  int(req.GetPerPage()),
		GenreIDs: make([]int, len(req.GetGenreIds())),
	}

	for i, gid := range req.GetGenreIds() {
		listMoviesRq.GenreIDs[i] = int(gid)
	}

	// 2. Вызываем репозиторий
	listMoveRs, err := uc.repo.ListMovies(ctx, listMoviesRq)
	if err != nil {
		uc.log.Error("Usecase.ListMovies: ошибка получения списка фильмов", zap.Error(err))
		return nil, err
	}

	// 3. Маппим Entity → Protobuf
	moviesProto := make([]*protos.Movie, 0, len(listMoveRs.Movies))
	for _, m := range listMoveRs.Movies {
		protoGenres := make([]*protos.Genre, 0, len(m.Genres))
		for _, g := range m.Genres {
			protoGenres = append(protoGenres, &protos.Genre{
				Id:   int32(g.ID),
				Name: g.Name,
			})
		}

		moviesProto = append(moviesProto, &protos.Movie{
			Id:          int32(m.ID),
			Title:       m.Title,
			VideoUrl:    m.VideoURL,
			CoverUrl:    m.CoverURL,
			Description: m.Description,
			ReleaseDate: timestamppb.New(m.ReleaseDate),
			DurationMin: int32(m.DurationMin),
			Genres:      protoGenres,
			CreatedAt:   timestamppb.New(m.CreatedAt),
			UpdatedAt:   timestamppb.New(m.UpdatedAt),
		})
	}

	// 4. Формируем и возвращаем ответ
	resp := &protos.ListMoviesResponse{
		Movies: moviesProto,
		Total:  int32(listMoveRs.Total),
	}
	uc.log.Info("Usecase.ListMovies: сформирован ответ",
		zap.Int("returned", len(moviesProto)),
		zap.Int("total", listMoveRs.Total),
	)
	return resp, nil
}

// GetMovie возвращает подробную информацию о фильме по его ID.
//
// Параметры:
//   - ctx: контекст выполнения.
//   - req: DTO с идентификатором фильма.
//
// Возвращает:
//   - Movie: DTO с деталями фильма.
//   - error: ошибку, если фильм не найден или произошёл сбой БД.
func (uc *Usecase) GetMovie(ctx context.Context, req *protos.GetMovieRequest) (*protos.Movie, error) {
	uc.log.Info("Usecase.GetMovie: входной запрос", zap.Int("id", int(req.GetId())))

	// 1. Получаем сущность фильма из репозитория
	movieEntity, err := uc.repo.GetMovie(ctx, int(req.GetId()))
	if err != nil {
		uc.log.Error("Usecase.GetMovie: ошибка получения фильма", zap.Error(err), zap.Int("id", int(req.GetId())))
		return nil, err
	}
	protoGenres := make([]*protos.Genre, 0, len(movieEntity.Genres))
	for _, g := range movieEntity.Genres {
		protoGenres = append(protoGenres, &protos.Genre{
			Id:   int32(g.ID),
			Name: g.Name,
		})
	}

	// 2. Маппим Entity → Protobuf
	movieProto := &protos.Movie{
		Id:          int32(movieEntity.ID),
		Title:       movieEntity.Title,
		VideoUrl:    movieEntity.VideoURL,
		CoverUrl:    movieEntity.CoverURL,
		Description: movieEntity.Description,
		ReleaseDate: timestamppb.New(movieEntity.ReleaseDate),
		DurationMin: int32(movieEntity.DurationMin),
		Genres:      protoGenres,
		CreatedAt:   timestamppb.New(movieEntity.CreatedAt),
		UpdatedAt:   timestamppb.New(movieEntity.UpdatedAt),
	}

	uc.log.Info("Usecase.GetMovie: сформирован ответ", zap.Int32("id", movieProto.GetId()))
	return movieProto, nil
}

// CreateMovie создаёт новый фильм в системе.
//
// Параметры:
//   - ctx: контекст выполнения.
//   - req: DTO с данными для создания (название, ссылки, описание, дата, длительность, жанры).
//
// Возвращает:
//   - CreateMovieResponse: DTO с созданным фильмом.
//   - error: ошибку валидации или сбой при записи в БД.
func (uc *Usecase) CreateMovie(ctx context.Context, req *protos.CreateMovieRequest) (*protos.CreateMovieResponse, error) {
	uc.log.Info("Usecase.CreateMovie: входной запрос",
		zap.String("title", req.GetTitle()),
		zap.Any("genre_ids", req.GetGenreIds()),
	)

	// 1. Маппим Protobuf → Entity
	genreIDs := make([]int, len(req.GetGenreIds()))
	for i, gid := range req.GetGenreIds() {
		genreIDs[i] = int(gid)
	}

	movieEntity := &entities.Movie{
		Title:       req.GetTitle(),
		VideoURL:    req.GetVideoUrl(),
		CoverURL:    req.GetCoverUrl(),
		Description: req.GetDescription(),
		ReleaseDate: req.GetReleaseDate().AsTime(),
		DurationMin: int(req.GetDurationMin()),
		// предполагаемое поле для жанров в сущности
		//GenreIDs:    genreIDs,
	}

	// 2. Вызываем репозиторий для создания
	created, err := uc.repo.CreateMovie(ctx, movieEntity, genreIDs)
	if err != nil {
		uc.log.Error("Usecase.CreateMovie: ошибка создания фильма", zap.Error(err))
		return nil, err
	}

	// 3. Маппим Entity → Protobuf
	movieProto := &protos.Movie{
		Id:          int32(created.ID),
		Title:       created.Title,
		VideoUrl:    created.VideoURL,
		CoverUrl:    created.CoverURL,
		Description: created.Description,
		ReleaseDate: timestamppb.New(created.ReleaseDate),
		DurationMin: int32(created.DurationMin),
		CreatedAt:   timestamppb.New(created.CreatedAt),
		UpdatedAt:   timestamppb.New(created.UpdatedAt),
	}

	uc.log.Info("Usecase.CreateMovie: фильм успешно создан", zap.Int32("id", movieProto.GetId()))
	return &protos.CreateMovieResponse{Movie: movieProto}, nil
}

// DeleteMovie удаляет фильм по его ID.
//
// Параметры:
//   - ctx: контекст выполнения.
//   - req: DTO с идентификатором фильма для удаления.
//
// Возвращает:
//   - Empty: пустой ответ при успешном удалении.
//   - error: ошибку, если фильм не найден или сбой БД.
func (uc *Usecase) DeleteMovie(ctx context.Context, req *protos.DeleteMovieRequest) (*emptypb.Empty, error) {
	uc.log.Info("Usecase.DeleteMovie: входной запрос", zap.Int("id", int(req.GetId())))

	// 1. Маппим Protobuf → Entity
	movieEntity := &entities.Movie{
		ID: int(req.GetId()),
	}

	// 2. Вызываем репозиторий для удаления
	if err := uc.repo.DeleteMovie(ctx, movieEntity); err != nil {
		uc.log.Error("Usecase.DeleteMovie: ошибка удаления фильма", zap.Error(err), zap.Int("id", movieEntity.ID))
		return nil, err
	}

	// 3. Возвращаем пустой ответ
	uc.log.Info("Usecase.DeleteMovie: фильм успешно удалён", zap.Int("id", movieEntity.ID))
	return &emptypb.Empty{}, nil
}

// ListRatings возвращает постраничный список оценок для указанного фильма.
//
// Параметры:
//   - ctx: контекст выполнения.
//   - req: DTO с ID фильма и параметрами пагинации.
//
// Возвращает:
//   - ListRatingsResponse: DTO со списком оценок и общим количеством.
//   - error: ошибку выполнения.
func (uc *Usecase) ListRatings(ctx context.Context, req *protos.ListRatingsRequest) (*protos.ListRatingsResponse, error) {
	uc.log.Info("Usecase.ListRatings: входной запрос",
		zap.Int32("movie_id", req.GetMovieId()),
		zap.Int32("page", req.GetPage()),
		zap.Int32("per_page", req.GetPerPage()),
	)

	// 1. Маппим Protobuf → Entity
	listReq := &entities.ListRatingsRequest{
		MovieID: int(req.GetMovieId()),
		Page:    int(req.GetPage()),
		PerPage: int(req.GetPerPage()),
	}

	// 2. Вызываем репозиторий
	listRes, err := uc.repo.ListRatings(ctx, listReq)
	if err != nil {
		uc.log.Error("Usecase.ListRatings: ошибка получения оценок", zap.Error(err), zap.Int("movie_id", listReq.MovieID))
		return nil, err
	}

	// 3. Маппим Entity → Protobuf
	ratingsProto := make([]*protos.Rating, 0, len(listRes.Ratings))
	for _, r := range listRes.Ratings {
		ratingsProto = append(ratingsProto, &protos.Rating{
			Id:        int32(r.ID),
			MovieId:   int32(r.MovieID),
			UserId:    int32(r.UserID),
			Score:     int32(r.Score),
			CreatedAt: timestamppb.New(r.CreatedAt),
			UpdatedAt: timestamppb.New(r.UpdatedAt),
		})
	}

	// 4. Формируем и возвращаем ответ
	resp := &protos.ListRatingsResponse{
		Ratings: ratingsProto,
		Total:   int32(listRes.Total),
	}
	uc.log.Info("Usecase.ListRatings: сформирован ответ",
		zap.Int("returned", len(ratingsProto)),
		zap.Int("total", listRes.Total),
	)
	return resp, nil
}

// GetRating возвращает конкретную оценку по ID фильма и ID оценки.
//
// Параметры:
//   - ctx: контекст выполнения.
//   - req: DTO с идентификаторами фильма и оценки.
//
// Возвращает:
//   - Rating: DTO с деталями оценки.
//   - error: ошибку, если оценка не найдена или сбой БД.
func (uc *Usecase) GetRating(ctx context.Context, req *protos.GetRatingRequest) (*protos.Rating, error) {
	uc.log.Info("Usecase.GetRating: входной запрос",
		zap.Int32("movie_id", req.GetMovieId()),
		zap.Int32("rating_id", req.GetRatingId()),
	)

	// 1. Вызываем репозиторий для получения оценки
	ratingEntity, err := uc.repo.GetRating(ctx, int(req.GetMovieId()), int(req.GetRatingId()))
	if err != nil {
		uc.log.Error("Usecase.GetRating: ошибка получения оценки", zap.Error(err),
			zap.Int("movie_id", int(req.GetMovieId())),
			zap.Int("rating_id", int(req.GetRatingId())),
		)
		return nil, err
	}

	// 2. Маппим Entity → Protobuf
	ratingProto := &protos.Rating{
		Id:        int32(ratingEntity.ID),
		MovieId:   int32(ratingEntity.MovieID),
		UserId:    int32(ratingEntity.UserID),
		Score:     int32(ratingEntity.Score),
		CreatedAt: timestamppb.New(ratingEntity.CreatedAt),
		UpdatedAt: timestamppb.New(ratingEntity.UpdatedAt),
	}

	uc.log.Info("Usecase.GetRating: сформирован ответ", zap.Int32("id", ratingProto.GetId()))
	return ratingProto, nil
}

// CreateRating создаёт новую оценку для фильма.
//
// Параметры:
//   - ctx: контекст выполнения.
//   - req: DTO с данными для новой оценки (ID фильма, ID пользователя, значение от 1 до 10).
//
// Возвращает:
//   - CreateRatingResponse: DTO с созданной оценкой.
//   - error: ошибку валидации или записи в БД.
func (uc *Usecase) CreateRating(ctx context.Context, req *protos.CreateRatingRequest) (*protos.CreateRatingResponse, error) {
	uc.log.Info("Usecase.CreateRating: входной запрос",
		zap.Int32("movie_id", req.GetMovieId()),
		zap.Int32("user_id", req.GetUserId()),
		zap.Int32("score", req.GetScore()),
	)

	// 1. Маппим Protobuf → Entity
	ratingEntity := &entities.Rating{
		MovieID: int(req.GetMovieId()),
		UserID:  int(req.GetUserId()),
		Score:   int(req.GetScore()),
	}

	// 2. Вызываем репозиторий для создания оценки
	created, err := uc.repo.CreateRating(ctx, ratingEntity)
	if err != nil {
		uc.log.Error("Usecase.CreateRating: ошибка создания оценки", zap.Error(err))
		return nil, err
	}

	// 3. Маппим Entity → Protobuf
	ratingProto := &protos.Rating{
		Id:        int32(created.ID),
		MovieId:   int32(created.MovieID),
		UserId:    int32(created.UserID),
		Score:     int32(created.Score),
		CreatedAt: timestamppb.New(created.CreatedAt),
		UpdatedAt: timestamppb.New(created.UpdatedAt),
	}

	uc.log.Info("Usecase.CreateRating: оценка успешно создана", zap.Int32("id", ratingProto.GetId()))
	return &protos.CreateRatingResponse{Rating: ratingProto}, nil
}

// DeleteRating удаляет оценку по ID фильма и ID оценки.
//
// Параметры:
//   - ctx: контекст выполнения.
//   - req: DTO с идентификаторами фильма и оценки.
//
// Возвращает:
//   - Empty: пустой ответ при успешном удалении.
//   - error: ошибку, если оценка не найдена или сбой БД.
func (uc *Usecase) DeleteRating(ctx context.Context, req *protos.DeleteRatingRequest) (*emptypb.Empty, error) {
	uc.log.Info("Usecase.DeleteRating: входной запрос",
		zap.Int32("movie_id", req.GetMovieId()),
		zap.Int32("rating_id", req.GetRatingId()),
	)

	// 1. Маппим Protobuf → Entity
	ratingEntity := &entities.Rating{
		ID:      int(req.GetRatingId()),
		MovieID: int(req.GetMovieId()),
	}

	// 2. Вызываем репозиторий для удаления оценки
	if err := uc.repo.DeleteRating(ctx, ratingEntity); err != nil {
		uc.log.Error("Usecase.DeleteRating: ошибка удаления оценки", zap.Error(err), zap.Int("rating_id", ratingEntity.ID))
		return nil, err
	}

	uc.log.Info("Usecase.DeleteRating: оценка успешно удалена", zap.Int("rating_id", ratingEntity.ID))
	return &emptypb.Empty{}, nil
}

// ListComments возвращает постраничный список комментариев к фильму.
//
// Параметры:
//   - ctx: контекст выполнения.
//   - req: DTO с ID фильма и параметрами пагинации.
//
// Возвращает:
//   - ListCommentsResponse: DTO со списком комментариев и общим количеством.
//   - error: ошибку выполнения.
func (uc *Usecase) ListComments(ctx context.Context, req *protos.ListCommentsRequest) (*protos.ListCommentsResponse, error) {
	uc.log.Info("Usecase.ListComments: входной запрос",
		zap.Int32("movie_id", req.GetMovieId()),
		zap.Int32("page", req.GetPage()),
		zap.Int32("per_page", req.GetPerPage()),
	)

	// 1. Маппим Protobuf → Entity
	listReq := &entities.ListCommentsRequest{
		MovieID: int(req.GetMovieId()),
		Page:    int(req.GetPage()),
		PerPage: int(req.GetPerPage()),
	}

	// 2. Вызываем репозиторий
	listRes, err := uc.repo.ListComments(ctx, listReq)
	if err != nil {
		uc.log.Error("Usecase.ListComments: ошибка получения комментариев", zap.Error(err),
			zap.Int("movie_id", listReq.MovieID),
		)
		return nil, err
	}

	// 3. Маппим Entity → Protobuf
	commentsProto := make([]*protos.Comment, 0, len(listRes.Comments))
	for _, c := range listRes.Comments {
		commentsProto = append(commentsProto, &protos.Comment{
			Id:        int32(c.ID),
			MovieId:   int32(c.MovieID),
			UserId:    int32(c.UserID),
			Text:      c.Text,
			CreatedAt: timestamppb.New(c.CreatedAt),
			UpdatedAt: timestamppb.New(c.UpdatedAt),
		})
	}

	// 4. Формируем и возвращаем ответ
	resp := &protos.ListCommentsResponse{
		Comments: commentsProto,
		Total:    int32(listRes.Total),
	}
	uc.log.Info("Usecase.ListComments: сформирован ответ",
		zap.Int("returned", len(commentsProto)),
		zap.Int("total", listRes.Total),
	)
	return resp, nil
}

// GetComment возвращает конкретный комментарий по ID фильма и ID комментария.
//
// Параметры:
//   - ctx: контекст выполнения.
//   - req: DTO с идентификаторами фильма и комментария.
//
// Возвращает:
//   - Comment: DTO с текстом и метаданными комментария.
//   - error: ошибку, если комментарий не найден или сбой БД.
func (uc *Usecase) GetComment(ctx context.Context, req *protos.GetCommentRequest) (*protos.Comment, error) {
	uc.log.Info("Usecase.GetComment: входной запрос",
		zap.Int32("movie_id", req.GetMovieId()),
		zap.Int32("comment_id", req.GetCommentId()),
	)

	// 1. Вызываем репозиторий
	commentEntity, err := uc.repo.GetComment(ctx, int(req.GetMovieId()), int(req.GetCommentId()))
	if err != nil {
		uc.log.Error("Usecase.GetComment: ошибка получения комментария", zap.Error(err),
			zap.Int("movie_id", int(req.GetMovieId())),
			zap.Int("comment_id", int(req.GetCommentId())),
		)
		return nil, err
	}

	// 2. Маппим Entity → Protobuf
	commentProto := &protos.Comment{
		Id:        int32(commentEntity.ID),
		MovieId:   int32(commentEntity.MovieID),
		UserId:    int32(commentEntity.UserID),
		Text:      commentEntity.Text,
		CreatedAt: timestamppb.New(commentEntity.CreatedAt),
		UpdatedAt: timestamppb.New(commentEntity.UpdatedAt),
	}

	uc.log.Info("Usecase.GetComment: сформирован ответ", zap.Int32("id", commentProto.GetId()))
	return commentProto, nil
}

// CreateComment создаёт новый комментарий к фильму.
//
// Параметры:
//   - ctx: контекст выполнения.
//   - req: DTO с данными для нового комментария (ID фильма, ID пользователя, текст).
//
// Возвращает:
//   - CreateCommentResponse: DTO с созданным комментарием.
//   - error: ошибку валидации или записи в БД.
func (uc *Usecase) CreateComment(ctx context.Context, req *protos.CreateCommentRequest) (*protos.CreateCommentResponse, error) {
	uc.log.Info("Usecase.CreateComment: входной запрос",
		zap.Int32("movie_id", req.GetMovieId()),
		zap.Int32("user_id", req.GetUserId()),
	)

	// 1. Маппим Protobuf → Entity
	commentEntity := &entities.Comment{
		MovieID: int(req.GetMovieId()),
		UserID:  int(req.GetUserId()),
		Text:    req.GetText(),
	}

	// 2. Вызываем репозиторий для создания
	created, err := uc.repo.CreateComment(ctx, commentEntity)
	if err != nil {
		uc.log.Error("Usecase.CreateComment: ошибка создания комментария", zap.Error(err))
		return nil, err
	}

	// 3. Маппим Entity → Protobuf
	commentProto := &protos.Comment{
		Id:        int32(created.ID),
		MovieId:   int32(created.MovieID),
		UserId:    int32(created.UserID),
		Text:      created.Text,
		CreatedAt: timestamppb.New(created.CreatedAt),
		UpdatedAt: timestamppb.New(created.UpdatedAt),
	}

	uc.log.Info("Usecase.CreateComment: комментарий успешно создан", zap.Int32("id", commentProto.GetId()))
	return &protos.CreateCommentResponse{Comment: commentProto}, nil
}

// DeleteComment удаляет комментарий по ID фильма и ID комментария.
//
// Параметры:
//   - ctx: контекст выполнения.
//   - req: DTO с идентификаторами фильма и комментария.
//
// Возвращает:
//   - Empty: пустой ответ при успешном удалении.
//   - error: ошибку, если комментарий не найден или сбой БД.
func (uc *Usecase) DeleteComment(ctx context.Context, req *protos.DeleteCommentRequest) (*emptypb.Empty, error) {
	uc.log.Info("Usecase.DeleteComment: входной запрос",
		zap.Int32("movie_id", req.GetMovieId()),
		zap.Int32("comment_id", req.GetCommentId()),
	)

	// 1. Маппим Protobuf → Entity
	commentEntity := &entities.Comment{
		ID:      int(req.GetCommentId()),
		MovieID: int(req.GetMovieId()),
	}

	// 2. Вызываем репозиторий для удаления
	if err := uc.repo.DeleteComment(ctx, commentEntity); err != nil {
		uc.log.Error("Usecase.DeleteComment: ошибка удаления комментария", zap.Error(err),
			zap.Int("comment_id", commentEntity.ID),
		)
		return nil, err
	}

	uc.log.Info("Usecase.DeleteComment: комментарий успешно удалён", zap.Int("comment_id", commentEntity.ID))
	return &emptypb.Empty{}, nil
}
