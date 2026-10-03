-- name: CreateTransaction :one
INSERT INTO transactions (
    user_id,
    category_id,
    wallet_id,
    type,
    amount,
    description,
    transaction_date
)
SELECT
    sqlc.arg(user_id)::uuid,
    category.id,
    wallet.id,
    sqlc.arg(type),
    sqlc.arg(amount)::text::numeric,
    NULLIF(sqlc.arg(description)::text, ''),
    sqlc.arg(transaction_date)::date
FROM categories AS category
CROSS JOIN wallets AS wallet
WHERE category.id = sqlc.arg(category_id)::uuid
  AND category.user_id = sqlc.arg(user_id)::uuid
  AND category.type = sqlc.arg(type)
  AND wallet.id = sqlc.arg(wallet_id)::uuid
  AND wallet.user_id = sqlc.arg(user_id)::uuid
RETURNING id::text AS id, category_id::text AS category_id,
          wallet_id::text AS wallet_id, type, amount::text AS amount,
          description, transaction_date, created_at, updated_at;

-- name: ListTransactions :many
SELECT id::text AS id, category_id::text AS category_id,
       wallet_id::text AS wallet_id, type, amount::text AS amount,
       description, transaction_date, created_at, updated_at
FROM transactions
WHERE user_id = sqlc.arg(user_id)::uuid
  AND (
      sqlc.narg(month_start)::date IS NULL
      OR (
          transaction_date >= sqlc.narg(month_start)::date
          AND transaction_date < sqlc.narg(month_start)::date + INTERVAL '1 month'
      )
  )
  AND (sqlc.narg(type)::text IS NULL OR type = sqlc.narg(type)::text)
ORDER BY transaction_date DESC, created_at DESC, id DESC;

-- name: GetTransaction :one
SELECT id::text AS id, category_id::text AS category_id,
       wallet_id::text AS wallet_id, type, amount::text AS amount,
       description, transaction_date, created_at, updated_at
FROM transactions
WHERE id = sqlc.arg(id)::uuid
  AND user_id = sqlc.arg(user_id)::uuid;

-- name: UpdateTransaction :one
UPDATE transactions AS txn
SET
    category_id = COALESCE(sqlc.narg(category_id)::uuid, txn.category_id),
    wallet_id = COALESCE(sqlc.narg(wallet_id)::uuid, txn.wallet_id),
    type = COALESCE(sqlc.narg(type)::text, txn.type),
    amount = COALESCE(sqlc.narg(amount)::text::numeric, txn.amount),
    description = CASE
        WHEN sqlc.arg(set_description)::boolean
            THEN NULLIF(sqlc.narg(description)::text, '')
        ELSE txn.description
    END,
    transaction_date = COALESCE(sqlc.narg(transaction_date)::date, txn.transaction_date),
    updated_at = now()
WHERE txn.id = sqlc.arg(id)::uuid
  AND txn.user_id = sqlc.arg(user_id)::uuid
  AND EXISTS (
      SELECT 1
      FROM categories AS category
      WHERE category.id = COALESCE(sqlc.narg(category_id)::uuid, txn.category_id)
        AND category.user_id = txn.user_id
        AND category.type = COALESCE(sqlc.narg(type)::text, txn.type)
  )
  AND EXISTS (
      SELECT 1
      FROM wallets AS wallet
      WHERE wallet.id = COALESCE(sqlc.narg(wallet_id)::uuid, txn.wallet_id)
        AND wallet.user_id = txn.user_id
  )
RETURNING txn.id::text AS id,
          txn.category_id::text AS category_id,
          txn.wallet_id::text AS wallet_id,
          txn.type,
          txn.amount::text AS amount,
          txn.description,
          txn.transaction_date,
          txn.created_at,
          txn.updated_at;

-- name: DeleteTransaction :execrows
DELETE FROM transactions
WHERE id = sqlc.arg(id)::uuid
  AND user_id = sqlc.arg(user_id)::uuid;
