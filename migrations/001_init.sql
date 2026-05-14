-- Create users table
CREATE TABLE IF NOT EXISTS users (
    id         TEXT        PRIMARY KEY,
    username   TEXT        NOT NULL UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Create boards table
CREATE TABLE IF NOT EXISTS boards (
    id                       SERIAL      PRIMARY KEY,
    name                     TEXT        NOT NULL,
    description              TEXT        NOT NULL DEFAULT '',
    schedule_type            TEXT,
    schedule_interval_seconds INT,
    created_at               TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at               TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Create scores table
-- Scores are also mirrored in Redis ZSET for fast ranked reads.
-- This table acts as the persistent source of truth.
CREATE TABLE IF NOT EXISTS scores (
    id         SERIAL      PRIMARY KEY,
    board_id   INT         NOT NULL REFERENCES boards(id) ON DELETE CASCADE,
    user_id    TEXT        NOT NULL REFERENCES users(id)  ON DELETE CASCADE,
    score      FLOAT       NOT NULL DEFAULT 0,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (board_id, user_id)
);

