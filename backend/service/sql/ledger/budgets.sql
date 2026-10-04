-- name: CreateBudget :one
INSERT INTO budgets (user_id, category_id, amount, month_start)
SELECT
    sqlc.arg(user_id)::uuid,
    category.id,
    sqlc.arg(amount)::text::numeric,
    sqlc.arg(month_start)::date
FROM categories AS category
WHERE category.id = sqlc.arg(category_id)::uuid
  AND category.user_id = sqlc.arg(user_id)::uuid
  AND category.type = 'expense'
RETURNING id::text AS id, category_id::text AS category_id,
          amount::text AS amount, month_start, created_at, updated_at;

-- name: ListBudgets :many
SELECT id::text AS id, category_id::text AS category_id,
       amount::text AS amount, month_start, created_at, updated_at
FROM budgets
WHERE user_id = sqlc.arg(user_id)::uuid
  AND (sqlc.narg(month_start)::date IS NULL OR month_start = sqlc.narg(month_start)::date)
ORDER BY month_start DESC, created_at DESC, id DESC;

-- name: GetBudget :one
SELECT id::text AS id, category_id::text AS category_id,
       amount::text AS amount, month_start, created_at, updated_at
FROM budgets
WHERE id = sqlc.arg(id)::uuid
  AND user_id = sqlc.arg(user_id)::uuid;

-- name: UpdateBudget :one
UPDATE budgets AS budget
SET
    category_id = COALESCE(sqlc.narg(category_id)::uuid, budget.category_id),
    amount = COALESCE(sqlc.narg(amount)::text::numeric, budget.amount),
    month_start = COALESCE(sqlc.narg(month_start)::date, budget.month_start),
    updated_at = now()
WHERE budget.id = sqlc.arg(id)::uuid
  AND budget.user_id = sqlc.arg(user_id)::uuid
  AND EXISTS (
      SELECT 1
      FROM categories AS category
      WHERE category.id = COALESCE(sqlc.narg(category_id)::uuid, budget.category_id)
        AND category.user_id = budget.user_id
        AND category.type = 'expense'
  )
RETURNING budget.id::text AS id,
          budget.category_id::text AS category_id,
          budget.amount::text AS amount,
          budget.month_start,
          budget.created_at,
          budget.updated_at;

-- name: DeleteBudget :execrows
DELETE FROM budgets
WHERE id = sqlc.arg(id)::uuid
  AND user_id = sqlc.arg(user_id)::uuid;
