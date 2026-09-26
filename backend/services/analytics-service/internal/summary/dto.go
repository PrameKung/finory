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

type TrendSeries struct {
	Month       string       `json:"month"`
	Granularity string       `json:"granularity"`
	Points      []TrendPoint `json:"points"`
}

type TrendPoint struct {
	Date    string `json:"date"`
	Income  string `json:"income"`
	Expense string `json:"expense"`
}

type MonthlyComparison struct {
	Month         string         `json:"month"`
	PreviousMonth string         `json:"previousMonth"`
	Current       MonthlyTotals  `json:"current"`
	Previous      MonthlyTotals  `json:"previous"`
	Changes       MonthlyChanges `json:"changes"`
}

type MonthlyTotals struct {
	Income  string `json:"income"`
	Expense string `json:"expense"`
}

type MonthlyChanges struct {
	Income  AmountChange `json:"income"`
	Expense AmountChange `json:"expense"`
}

type AmountChange struct {
	Amount     string  `json:"amount"`
	Percentage *string `json:"percentage"`
}
