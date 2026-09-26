-- +goose Up
CREATE UNIQUE INDEX categories_id_user_id_unique
    ON categories (id, user_id);

CREATE UNIQUE INDEX wallets_id_user_id_unique
    ON wallets (id, user_id);

CREATE TABLE transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    category_id UUID NOT NULL,
    wallet_id UUID NOT NULL,
    type TEXT NOT NULL,
    amount NUMERIC(19, 4) NOT NULL,
    description TEXT,
    transaction_date DATE NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT transactions_type_valid CHECK (type IN ('income', 'expense')),
    CONSTRAINT transactions_amount_positive CHECK (amount > 0),
    CONSTRAINT transactions_description_not_blank CHECK (
        description IS NULL OR length(btrim(description)) > 0
    ),
    CONSTRAINT transactions_description_length CHECK (
        description IS NULL OR char_length(description) <= 500
    ),
    CONSTRAINT transactions_category_owned_by_user_fk
        FOREIGN KEY (category_id, user_id)
        REFERENCES categories (id, user_id),
    CONSTRAINT transactions_wallet_owned_by_user_fk
        FOREIGN KEY (wallet_id, user_id)
        REFERENCES wallets (id, user_id)
);

CREATE INDEX transactions_user_date_idx
    ON transactions (user_id, transaction_date DESC);

CREATE INDEX transactions_user_type_date_idx
    ON transactions (user_id, type, transaction_date DESC);

CREATE INDEX transactions_user_category_idx
    ON transactions (user_id, category_id);

CREATE INDEX transactions_user_wallet_idx
    ON transactions (user_id, wallet_id);

-- +goose Down
DROP TABLE transactions;
DROP INDEX wallets_id_user_id_unique;
DROP INDEX categories_id_user_id_unique;
