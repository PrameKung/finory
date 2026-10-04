package wallets

import "time"

type Wallet struct {
	ID           string
	Name         string
	Type         string
	Balance      string
	CurrencyCode string
	IsDefault    bool
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
