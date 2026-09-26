-- +goose Up
CREATE TABLE budgets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    category_id UUID NOT NULL,
    amount NUMERIC(19, 4) NOT NULL,
    month_start DATE NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT budgets_amount_positive CHECK (amount > 0),
    CONSTRAINT budgets_month_start_valid CHECK (
        month_start = date_trunc('month', month_start)::date
    ),
    CONSTRAINT budgets_category_owned_by_user_fk
        FOREIGN KEY (category_id, user_id)
        REFERENCES categories (id, user_id),
    CONSTRAINT budgets_user_category_month_unique
        UNIQUE (user_id, category_id, month_start)
);

CREATE INDEX budgets_user_month_idx
    ON budgets (user_id, month_start DESC);

-- +goose Down
DROP TABLE budgets;
