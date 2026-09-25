-- +goose Up
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    google_subject TEXT NOT NULL UNIQUE,
    email TEXT NOT NULL,
    display_name TEXT,
    avatar_url TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT users_google_subject_not_blank CHECK (length(btrim(google_subject)) > 0),
    CONSTRAINT users_email_not_blank CHECK (length(btrim(email)) > 0)
);

-- +goose Down
DROP TABLE users;
