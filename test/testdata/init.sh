#!/bin/bash
set -e

psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" <<-EOSQL
    CREATE TABLE IF NOT EXISTS users
    (
        id            SERIAL PRIMARY KEY,
        telegram_id   BIGINT UNIQUE NOT NULL,
        first_name    VARCHAR(255),
        access_token  VARCHAR(255),
        refresh_token VARCHAR(255),
        created_at    TIMESTAMP DEFAULT NOW(),
        updated_at    TIMESTAMP DEFAULT NOW()
    );
EOSQL
