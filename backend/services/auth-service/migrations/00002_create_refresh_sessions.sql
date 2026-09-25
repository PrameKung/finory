-- +goose Up
CREATE TABLE refresh_sessions (
    token_hash BYTEA PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT refresh_sessions_token_hash_length CHECK (octet_length(token_hash) = 32)
);

-- +goose Down
DROP TABLE refresh_sessions;
