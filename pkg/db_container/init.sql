\c postgres
CREATE EXTENSION IF NOT EXISTS dblink;
DO
$$
    BEGIN
        IF NOT EXISTS (SELECT FROM pg_database WHERE datname = 'online_movie_movies') THEN
            PERFORM dblink_exec('dbname=postgres user=' || current_user, 'CREATE DATABASE online_movie_movies');
            RAISE NOTICE 'Created database: %', 'online_movie_movies';
        END IF;
    END;
$$;
\c online_movie_movies
