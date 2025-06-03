-- roles
create table roles(
                      id INTEGER PRIMARY KEY ,
                      name TEXT
)

-- users
CREATE TABLE users (
                       id INTEGER PRIMARY KEY,
                       email TEXT UNIQUE NOT NULL,
                       password_hash TEXT NOT NULL,
                       name TEXT NOT NULL,
                       role_id INTEGER REFERENCES roles(id),
                       created_at TIMESTAMP DEFAULT now(),
                       updated_at TIMESTAMP DEFAULT now()
);

-- refresh_tokens
CREATE TABLE refresh_tokens (
                                id INTEGER PRIMARY KEY,
                                user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
                                token TEXT UNIQUE NOT NULL,
                                expires_at TIMESTAMP NOT NULL,
                                revoked BOOLEAN DEFAULT FALSE
);