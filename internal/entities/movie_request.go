package entities

// ListMoviesRequest представляет параметры запроса GET /api/v1/movies
// Параметры передаются как query params: page, per_page, genre_ids
type ListMoviesRequest struct {
	Page     int   `json:"page" form:"page"`
	PerPage  int   `json:"per_page" form:"per_page"`
	GenreIDs []int `json:"genre_ids" form:"genre_ids"`
}

// ListMoviesResponse соответствует proto-сообщению:
//
//	message ListMoviesResponse {
//	  repeated Movie movies = 1;
//	  int32 total = 2;  // общее количество фильмов, подходящих под фильтр
//	}
type ListMoviesResponse struct {
	Movies []*Movie `json:"movies"`
	Total  int      `json:"total"`
}

// ListRatingsRequest представляет параметры запроса GET /api/v1/ratings
// Параметры передаются как query params: page, per_page, genre_ids
type ListRatingsRequest struct {
	MovieID int `json:"movie_id" form:"movie_id"`
	Page    int `json:"page" form:"page"`
	PerPage int `json:"per_page" form:"per_page"`
}

type ListRatingsResponse struct {
	Ratings []*Rating `json:"ratings"`
	Total   int       `json:"total"`
}

type ListCommentsRequest struct {
	MovieID int `json:"movie_id" form:"movie_id"`
	Page    int `json:"page" form:"page"`
	PerPage int `json:"per_page" form:"per_page"`
}
type ListCommentsResponse struct {
	Comments []*Comment `json:"comments"`
	Total    int        `json:"total"`
}
