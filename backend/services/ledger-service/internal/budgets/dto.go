package budgets

type CreateInput struct {
	CategoryID string `json:"categoryId"`
	Amount     string `json:"amount"`
	Month      string `json:"month"`
}

type UpdateInput struct {
	CategoryID *string `json:"categoryId"`
	Amount     *string `json:"amount"`
	Month      *string `json:"month"`
}

type ListFilter struct {
	Month string
}

type budgetResponse struct {
	ID         string `json:"id"`
	CategoryID string `json:"categoryId"`
	Amount     string `json:"amount"`
	Month      string `json:"month"`
	CreatedAt  string `json:"createdAt"`
	UpdatedAt  string `json:"updatedAt"`
}
