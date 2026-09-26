package transactions

import "time"

type Transaction struct {
	ID              string
	CategoryID      string
	WalletID        string
	Type            string
	Amount          string
	Description     *string
	TransactionDate time.Time
	CreatedAt       time.Time
	UpdatedAt       time.Time
}
