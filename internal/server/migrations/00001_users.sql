-- +goose Up
CREATE TABLE IF NOT EXISTS users (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    login      VARCHAR(255) NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    created_at BIGINT NOT NULL DEFAULT extract(epoch FROM now())
);

CREATE INDEX idx_users_login ON users (login);

-- +goose Down
DROP TABLE IF EXISTS users;
