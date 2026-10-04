package ledger

// These decimal-string views keep analytics calculations independent of
// database representation and avoid converting monetary values to floats.
type Transaction struct {
	ID              string  `json:"id"`
	CategoryID      string  `json:"categoryId"`
	WalletID        string  `json:"walletId"`
	Type            string  `json:"type"`
	Amount          string  `json:"amount"`
	Description     *string `json:"description"`
	TransactionDate string  `json:"transactionDate"`
	CreatedAt       string  `json:"createdAt"`
	UpdatedAt       string  `json:"updatedAt"`
}

type Category struct {
	ID        string  `json:"id"`
	Name      string  `json:"name"`
	Type      string  `json:"type"`
	Icon      *string `json:"icon"`
	Color     *string `json:"color"`
	IsDefault bool    `json:"isDefault"`
	CreatedAt string  `json:"createdAt"`
	UpdatedAt string  `json:"updatedAt"`
}

type MonthlyData struct {
	Transactions []Transaction
	Categories   []Category
}
