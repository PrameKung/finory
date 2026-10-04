-- +goose Up
CREATE TABLE categories (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    name TEXT NOT NULL,
    type TEXT NOT NULL,
    icon TEXT,
    color TEXT,
    is_default BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT categories_name_not_blank CHECK (length(btrim(name)) > 0),
    CONSTRAINT categories_name_length CHECK (char_length(name) <= 100),
    CONSTRAINT categories_type_valid CHECK (type IN ('income', 'expense')),
    CONSTRAINT categories_icon_not_blank CHECK (icon IS NULL OR length(btrim(icon)) > 0),
    CONSTRAINT categories_color_not_blank CHECK (color IS NULL OR length(btrim(color)) > 0)
);

CREATE UNIQUE INDEX categories_user_type_name_unique
    ON categories (user_id, type, lower(name));

CREATE INDEX categories_user_id_idx ON categories (user_id);

CREATE TABLE wallets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    name TEXT NOT NULL,
    type TEXT NOT NULL,
    balance NUMERIC(19, 4) NOT NULL DEFAULT 0,
    currency_code TEXT NOT NULL,
    is_default BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT wallets_name_not_blank CHECK (length(btrim(name)) > 0),
    CONSTRAINT wallets_name_length CHECK (char_length(name) <= 100),
    CONSTRAINT wallets_type_valid CHECK (type IN ('cash', 'bank', 'e_wallet', 'other')),
    CONSTRAINT wallets_currency_code_valid CHECK (currency_code ~ '^[A-Z]{3}$')
);

CREATE UNIQUE INDEX wallets_user_name_unique
    ON wallets (user_id, lower(name));

CREATE INDEX wallets_user_id_idx ON wallets (user_id);

-- +goose Down
DROP TABLE wallets;
DROP TABLE categories;
