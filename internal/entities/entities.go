package entities

import (
	"time"
)

// Genre -------------------------
// Сущность Genre <-> DTO
// -------------------------
type Genre struct {
	ID   int    `json:"id" db:"id"`
	Name string `json:"name" db:"name"`
}

type GenreDTO struct {
	ID   *int    `json:"id,omitempty"`
	Name *string `json:"name,omitempty"`
}

func (g *Genre) ToDTO() *GenreDTO {
	return &GenreDTO{
		ID:   &g.ID,
		Name: &g.Name,
	}
}

func (d *GenreDTO) ToEntity() *Genre {
	g := &Genre{}
	if d.ID != nil {
		g.ID = *d.ID
	}
	if d.Name != nil {
		g.Name = *d.Name
	}
	return g
}

// Movie ---------------------------------------
// Сущность Movie <-> DTO
// Таблица movies:
//
//	id           SERIAL PRIMARY KEY,
//	title        VARCHAR(255) NOT NULL,
//	video_url    TEXT         NOT NULL,
//	cover_url    TEXT         NOT NULL,
//	description  TEXT         NOT NULL,
//	release_date DATE         NOT NULL,
//	duration_min INTEGER      NOT NULL,
//	created_at   TIMESTAMPTZ  NOT NULL DEFAULT now(),
//	updated_at   TIMESTAMPTZ  NOT NULL DEFAULT now()
//
// ---------------------------------------
type Movie struct {
	ID          int       `json:"id" db:"id"`
	Title       string    `json:"title" db:"title"`
	VideoURL    string    `json:"video_url" db:"video_url"`
	CoverURL    string    `json:"cover_url" db:"cover_url"`
	Description string    `json:"description" db:"description"`
	ReleaseDate time.Time `json:"release_date" db:"release_date"`
	DurationMin int       `json:"duration_min" db:"duration_min"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

type MovieDTO struct {
	ID          *int       `json:"id,omitempty"`
	Title       *string    `json:"title,omitempty"`
	VideoURL    *string    `json:"video_url,omitempty"`
	CoverURL    *string    `json:"cover_url,omitempty"`
	Description *string    `json:"description,omitempty"`
	ReleaseDate *time.Time `json:"release_date,omitempty"`
	DurationMin *int       `json:"duration_min,omitempty"`
	CreatedAt   *time.Time `json:"created_at,omitempty"`
	UpdatedAt   *time.Time `json:"updated_at,omitempty"`
	// Для передачи связей «movie ↔ genre» по ID:
	GenreIDs *[]int `json:"genre_ids,omitempty"`
}

func (m *Movie) ToDTO(genreIDs []int) *MovieDTO {
	// genreIDs передаётся «снаружи» – т.е. тот slice int, который вы получили при JOIN (через movie_genres).
	return &MovieDTO{
		ID:          &m.ID,
		Title:       &m.Title,
		VideoURL:    &m.VideoURL,
		CoverURL:    &m.CoverURL,
		Description: &m.Description,
		ReleaseDate: &m.ReleaseDate,
		DurationMin: &m.DurationMin,
		CreatedAt:   &m.CreatedAt,
		UpdatedAt:   &m.UpdatedAt,
		GenreIDs:    &genreIDs,
	}
}

func (d *MovieDTO) ToEntity() *Movie {
	m := &Movie{}
	if d.ID != nil {
		m.ID = *d.ID
	}
	if d.Title != nil {
		m.Title = *d.Title
	}
	if d.VideoURL != nil {
		m.VideoURL = *d.VideoURL
	}
	if d.CoverURL != nil {
		m.CoverURL = *d.CoverURL
	}
	if d.Description != nil {
		m.Description = *d.Description
	}
	if d.ReleaseDate != nil {
		m.ReleaseDate = *d.ReleaseDate
	}
	if d.DurationMin != nil {
		m.DurationMin = *d.DurationMin
	}
	if d.CreatedAt != nil {
		m.CreatedAt = *d.CreatedAt
	}
	if d.UpdatedAt != nil {
		m.UpdatedAt = *d.UpdatedAt
	}
	// GenreIDs не "кладём" внутрь Movie, с ними работает слой репозитория (INSERT в movie_genres).
	return m
}

// MovieGenre ----------------------------------------------------------
// Сущность MovieGenre <-> DTO (связующая таблица)
// CREATE TABLE movie_genres ( movie_id INT, genre_id INT, PRIMARY KEY(movie_id,genre_id) );
// ----------------------------------------------------------
type MovieGenre struct {
	MovieID int `json:"movie_id" db:"movie_id"`
	GenreID int `json:"genre_id" db:"genre_id"`
}

type MovieGenreDTO struct {
	MovieID *int `json:"movie_id,omitempty"`
	GenreID *int `json:"genre_id,omitempty"`
}

func (mg *MovieGenre) ToDTO() *MovieGenreDTO {
	return &MovieGenreDTO{
		MovieID: &mg.MovieID,
		GenreID: &mg.GenreID,
	}
}

func (d *MovieGenreDTO) ToEntity() *MovieGenre {
	mg := &MovieGenre{}
	if d.MovieID != nil {
		mg.MovieID = *d.MovieID
	}
	if d.GenreID != nil {
		mg.GenreID = *d.GenreID
	}
	return mg
}

// Rating ----------------------------------------------------------
// Сущность Rating <-> DTO
// Таблица ratings:
//
//	id         SERIAL PRIMARY KEY,
//	movie_id   INTEGER     NOT NULL REFERENCES movies(id),
//	user_id    INTEGER     NOT NULL,
//	score      SMALLINT    NOT NULL CHECK(score BETWEEN 1 AND 10),
//	created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
//	updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
//
// ----------------------------------------------------------
type Rating struct {
	ID        int       `json:"id" db:"id"`
	MovieID   int       `json:"movie_id" db:"movie_id"`
	UserID    int       `json:"user_id" db:"user_id"`
	Score     int       `json:"score" db:"score"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

type RatingDTO struct {
	ID        *int       `json:"id,omitempty"`
	MovieID   *int       `json:"movie_id,omitempty"`
	UserID    *int       `json:"user_id,omitempty"`
	Score     *int       `json:"score,omitempty"`
	CreatedAt *time.Time `json:"created_at,omitempty"`
	UpdatedAt *time.Time `json:"updated_at,omitempty"`
}

func (r *Rating) ToDTO() *RatingDTO {
	return &RatingDTO{
		ID:        &r.ID,
		MovieID:   &r.MovieID,
		UserID:    &r.UserID,
		Score:     &r.Score,
		CreatedAt: &r.CreatedAt,
		UpdatedAt: &r.UpdatedAt,
	}
}

func (d *RatingDTO) ToEntity() *Rating {
	r := &Rating{}
	if d.ID != nil {
		r.ID = *d.ID
	}
	if d.MovieID != nil {
		r.MovieID = *d.MovieID
	}
	if d.UserID != nil {
		r.UserID = *d.UserID
	}
	if d.Score != nil {
		r.Score = *d.Score
	}
	if d.CreatedAt != nil {
		r.CreatedAt = *d.CreatedAt
	}
	if d.UpdatedAt != nil {
		r.UpdatedAt = *d.UpdatedAt
	}
	return r
}

// Comment ----------------------------------------------------------
// Сущность Comment <-> DTO
// Таблица comments:
//
//	id         SERIAL PRIMARY KEY,
//	movie_id   INTEGER     NOT NULL REFERENCES movies(id),
//	user_id    INTEGER     NOT NULL,
//	text       TEXT        NOT NULL,
//	created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
//	updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
//
// ----------------------------------------------------------
type Comment struct {
	ID        int       `json:"id" db:"id"`
	MovieID   int       `json:"movie_id" db:"movie_id"`
	UserID    int       `json:"user_id" db:"user_id"`
	Text      string    `json:"text" db:"text"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

type CommentDTO struct {
	ID        *int       `json:"id,omitempty"`
	MovieID   *int       `json:"movie_id,omitempty"`
	UserID    *int       `json:"user_id,omitempty"`
	Text      *string    `json:"text,omitempty"`
	CreatedAt *time.Time `json:"created_at,omitempty"`
	UpdatedAt *time.Time `json:"updated_at,omitempty"`
}

func (c *Comment) ToDTO() *CommentDTO {
	return &CommentDTO{
		ID:        &c.ID,
		MovieID:   &c.MovieID,
		UserID:    &c.UserID,
		Text:      &c.Text,
		CreatedAt: &c.CreatedAt,
		UpdatedAt: &c.UpdatedAt,
	}
}

func (d *CommentDTO) ToEntity() *Comment {
	c := &Comment{}
	if d.ID != nil {
		c.ID = *d.ID
	}
	if d.MovieID != nil {
		c.MovieID = *d.MovieID
	}
	if d.UserID != nil {
		c.UserID = *d.UserID
	}
	if d.Text != nil {
		c.Text = *d.Text
	}
	if d.CreatedAt != nil {
		c.CreatedAt = *d.CreatedAt
	}
	if d.UpdatedAt != nil {
		c.UpdatedAt = *d.UpdatedAt
	}
	return c
}
