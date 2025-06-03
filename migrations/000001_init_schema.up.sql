CREATE TABLE IF NOT EXISTS movies
(
    id           SERIAL PRIMARY KEY,
    title        VARCHAR(255) NOT NULL,
    video_url    TEXT         NOT NULL,
    cover_url    TEXT         NOT NULL,
    description  TEXT         NOT NULL,
    release_date DATE         NOT NULL,
    duration_min INTEGER      NOT NULL,
    created_at   TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ  NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS genres
(
    id   SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL UNIQUE
);

CREATE TABLE IF NOT EXISTS movie_genres
(
    movie_id INTEGER NOT NULL REFERENCES movies (id),
    genre_id INTEGER NOT NULL REFERENCES genres (id),
    PRIMARY KEY (
                 movie_id,
                 genre_id)
);

CREATE TABLE IF NOT EXISTS ratings
(
    id         SERIAL PRIMARY KEY,
    movie_id   INTEGER     NOT NULL REFERENCES movies (id),
    user_id    INTEGER     NOT NULL,
    score      SMALLINT    NOT NULL CHECK (score BETWEEN 1 AND 10),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS comments
(
    id         SERIAL PRIMARY KEY,
    movie_id   INTEGER     NOT NULL REFERENCES movies (id),
    user_id    INTEGER     NOT NULL,
    text       TEXT        NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_movies_title ON movies (title);
CREATE INDEX IF NOT EXISTS idx_movies_release_date ON movies (release_date);
CREATE INDEX IF NOT EXISTS idx_moviegenres_genre ON movie_genres (genre_id);
CREATE INDEX IF NOT EXISTS idx_ratings_movie ON ratings (movie_id);
CREATE INDEX IF NOT EXISTS idx_comments_movie ON comments (movie_id);