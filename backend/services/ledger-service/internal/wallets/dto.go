package wallets

type CreateInput struct {
	Name         string `json:"name"`
	Type         string `json:"type"`
	Balance      string `json:"balance"`
	CurrencyCode string `json:"currencyCode"`
}

type UpdateInput struct {
	Name         *string `json:"name"`
	Type         *string `json:"type"`
	Balance      *string `json:"balance"`
	CurrencyCode *string `json:"currencyCode"`
}

type walletResponse struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Type         string `json:"type"`
	Balance      string `json:"balance"`
	CurrencyCode string `json:"currencyCode"`
	IsDefault    bool   `json:"isDefault"`
	CreatedAt    string `json:"createdAt"`
	UpdatedAt    string `json:"updatedAt"`
}
