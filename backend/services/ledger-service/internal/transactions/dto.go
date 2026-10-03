package transactions

type CreateInput struct {
	CategoryID      string `json:"categoryId"`
	WalletID        string `json:"walletId"`
	Type            string `json:"type"`
	Amount          string `json:"amount"`
	Description     string `json:"description"`
	TransactionDate string `json:"transactionDate"`
}

type UpdateInput struct {
	CategoryID      *string `json:"categoryId"`
	WalletID        *string `json:"walletId"`
	Type            *string `json:"type"`
	Amount          *string `json:"amount"`
	Description     *string `json:"description"`
	TransactionDate *string `json:"transactionDate"`
}

type ListFilter struct {
	Month string
	Type  string
}

type transactionResponse struct {
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
