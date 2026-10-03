-- name: CreateDefaultCashWallet :execrows
INSERT INTO wallets (user_id, name, type, balance, currency_code, is_default)
VALUES (sqlc.arg(user_id)::uuid, 'Cash', 'cash', 0, 'THB', true)
ON CONFLICT (user_id, lower(name)) DO NOTHING;

-- name: ListWallets :many
SELECT id::text AS id, name, type, balance::text AS balance, currency_code,
       is_default, created_at, updated_at
FROM wallets
WHERE user_id = sqlc.arg(user_id)::uuid
ORDER BY is_default DESC, lower(name), id;

-- name: CreateWallet :one
INSERT INTO wallets (user_id, name, type, balance, currency_code)
VALUES (
    sqlc.arg(user_id)::uuid,
    sqlc.arg(name),
    sqlc.arg(type),
    sqlc.arg(balance)::text::numeric,
    sqlc.arg(currency_code)
)
RETURNING id::text AS id, name, type, balance::text AS balance, currency_code,
          is_default, created_at, updated_at;

-- name: UpdateWallet :one
UPDATE wallets
SET
    name = COALESCE(sqlc.narg(name)::text, name),
    type = COALESCE(sqlc.narg(type)::text, type),
    balance = COALESCE(sqlc.narg(balance)::text::numeric, balance),
    currency_code = COALESCE(sqlc.narg(currency_code)::text, currency_code),
    updated_at = now()
WHERE id = sqlc.arg(id)::uuid
  AND user_id = sqlc.arg(user_id)::uuid
  AND is_default = false
RETURNING id::text AS id, name, type, balance::text AS balance, currency_code,
          is_default, created_at, updated_at;

-- name: DeleteWallet :execrows
DELETE FROM wallets
WHERE id = sqlc.arg(id)::uuid
  AND user_id = sqlc.arg(user_id)::uuid
  AND is_default = false;
