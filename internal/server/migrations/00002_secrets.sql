-- +goose Up
CREATE TABLE IF NOT EXISTS secrets (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name       VARCHAR(255) NOT NULL,
    type       INT NOT NULL,
    data       BYTEA NOT NULL,
    metadata   JSONB DEFAULT '{}',
    version    BIGINT NOT NULL DEFAULT 1,
    updated_at BIGINT NOT NULL DEFAULT extract(epoch FROM now()),
    is_deleted BOOLEAN NOT NULL DEFAULT FALSE
);

CREATE INDEX idx_secrets_user_id ON secrets (user_id);
CREATE INDEX idx_secrets_updated_at ON secrets (user_id, updated_at);

-- +goose Down
DROP TABLE IF EXISTS secrets;
