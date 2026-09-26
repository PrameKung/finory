-- name: CreateDefaultCashWallet :execrows
INSERT INTO wallets (user_id, name, type, balance, currency_code, is_default)
VALUES (sqlc.arg(user_id)::uuid, 'Cash', 'cash', 0, 'THB', true)
ON CONFLICT (user_id, lower(name)) DO NOTHING;
