package budgets

import "time"

type Budget struct {
	ID         string
	CategoryID string
	Amount     string
	MonthStart time.Time
	CreatedAt  time.Time
	UpdatedAt  time.Time
}
