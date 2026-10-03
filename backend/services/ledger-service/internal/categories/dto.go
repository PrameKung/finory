package categories

type CreateInput struct {
	Name  string `json:"name"`
	Type  string `json:"type"`
	Icon  string `json:"icon"`
	Color string `json:"color"`
}

type UpdateInput struct {
	Name  *string `json:"name"`
	Type  *string `json:"type"`
	Icon  *string `json:"icon"`
	Color *string `json:"color"`
}

type categoryResponse struct {
	ID        string  `json:"id"`
	Name      string  `json:"name"`
	Type      string  `json:"type"`
	Icon      *string `json:"icon"`
	Color     *string `json:"color"`
	IsDefault bool    `json:"isDefault"`
	CreatedAt string  `json:"createdAt"`
	UpdatedAt string  `json:"updatedAt"`
}
