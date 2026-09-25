-- name: UpsertGoogleUser :one
INSERT INTO users (google_subject, email, display_name, avatar_url)
VALUES (
    sqlc.arg(google_subject),
    sqlc.arg(email),
    NULLIF(sqlc.arg(display_name)::text, ''),
    NULLIF(sqlc.arg(avatar_url)::text, '')
)
ON CONFLICT (google_subject) DO UPDATE SET
    email = EXCLUDED.email,
    display_name = COALESCE(EXCLUDED.display_name, users.display_name),
    avatar_url = COALESCE(EXCLUDED.avatar_url, users.avatar_url),
    updated_at = now()
RETURNING id::text;

-- name: GetUserByID :one
SELECT id::text AS id, email,
       COALESCE(display_name, '')::text AS display_name,
       COALESCE(avatar_url, '')::text AS avatar_url
FROM users
WHERE id = sqlc.arg(id)::uuid;
