package summary

type MonthlySummary struct {
	Month   string `json:"month"`
	Income  string `json:"income"`
	Expense string `json:"expense"`
}

type CategoryDistribution struct {
	Month        string         `json:"month"`
	TotalExpense string         `json:"totalExpense"`
	Categories   []CategoryRank `json:"categories"`
}

type CategoryRank struct {
	Rank       int     `json:"rank"`
	CategoryID string  `json:"categoryId"`
	Name       string  `json:"name"`
	Icon       *string `json:"icon"`
	Color      *string `json:"color"`
	Amount     string  `json:"amount"`
	Percentage string  `json:"percentage"`
}
