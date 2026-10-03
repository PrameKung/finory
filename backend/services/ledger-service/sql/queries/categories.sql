-- name: CreateDefaultCategories :execrows
INSERT INTO categories (user_id, name, type, icon, color, is_default)
VALUES
    (sqlc.arg(user_id)::uuid, 'Salary', 'income', 'briefcase-business', '#16A34A', true),
    (sqlc.arg(user_id)::uuid, 'Freelance', 'income', 'laptop', '#0D9488', true),
    (sqlc.arg(user_id)::uuid, 'Investment', 'income', 'chart-no-axes-combined', '#2563EB', true),
    (sqlc.arg(user_id)::uuid, 'Other Income', 'income', 'circle-plus', '#7C3AED', true),
    (sqlc.arg(user_id)::uuid, 'Food & Dining', 'expense', 'utensils', '#EA580C', true),
    (sqlc.arg(user_id)::uuid, 'Transportation', 'expense', 'car', '#0891B2', true),
    (sqlc.arg(user_id)::uuid, 'Housing', 'expense', 'house', '#9333EA', true),
    (sqlc.arg(user_id)::uuid, 'Bills & Utilities', 'expense', 'receipt-text', '#CA8A04', true),
    (sqlc.arg(user_id)::uuid, 'Shopping', 'expense', 'shopping-bag', '#DB2777', true),
    (sqlc.arg(user_id)::uuid, 'Healthcare', 'expense', 'heart-pulse', '#DC2626', true),
    (sqlc.arg(user_id)::uuid, 'Entertainment', 'expense', 'gamepad-2', '#4F46E5', true),
    (sqlc.arg(user_id)::uuid, 'Other Expense', 'expense', 'circle-minus', '#64748B', true)
ON CONFLICT (user_id, type, lower(name)) DO NOTHING;

-- name: ListCategories :many
SELECT id::text AS id, name, type, icon, color, is_default, created_at, updated_at
FROM categories
WHERE user_id = sqlc.arg(user_id)::uuid
ORDER BY type, lower(name), id;

-- name: CreateCategory :one
INSERT INTO categories (user_id, name, type, icon, color)
VALUES (
    sqlc.arg(user_id)::uuid,
    sqlc.arg(name),
    sqlc.arg(type),
    NULLIF(sqlc.arg(icon)::text, ''),
    NULLIF(sqlc.arg(color)::text, '')
)
RETURNING id::text AS id, name, type, icon, color, is_default, created_at, updated_at;

-- name: UpdateCategory :one
UPDATE categories
SET
    name = COALESCE(sqlc.narg(name)::text, name),
    type = COALESCE(sqlc.narg(type)::text, type),
    icon = COALESCE(sqlc.narg(icon)::text, icon),
    color = COALESCE(sqlc.narg(color)::text, color),
    updated_at = now()
WHERE id = sqlc.arg(id)::uuid
  AND user_id = sqlc.arg(user_id)::uuid
  AND is_default = false
RETURNING id::text AS id, name, type, icon, color, is_default, created_at, updated_at;

-- name: DeleteCategory :execrows
DELETE FROM categories
WHERE id = sqlc.arg(id)::uuid
  AND user_id = sqlc.arg(user_id)::uuid
  AND is_default = false;
