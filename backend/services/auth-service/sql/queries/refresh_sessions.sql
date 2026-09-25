-- name: CreateRefreshSession :exec
INSERT INTO refresh_sessions (token_hash, user_id, expires_at)
VALUES (sqlc.arg(token_hash), sqlc.arg(user_id)::uuid, now() + interval '30 days');

-- name: RotateRefreshSession :one
UPDATE refresh_sessions
SET token_hash = sqlc.arg(new_token_hash),
    expires_at = now() + interval '30 days'
WHERE token_hash = sqlc.arg(old_token_hash)
  AND expires_at > now()
RETURNING user_id::text;

-- name: DeleteRefreshSession :exec
DELETE FROM refresh_sessions
WHERE token_hash = sqlc.arg(token_hash);
