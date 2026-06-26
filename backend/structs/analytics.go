package structs

type AnalyticsResponse struct {
	ActiveCount    int64 `json:"active_count"`
	ExpiringCount  int64 `json:"expiring_count"`
	MonthlyTotal   int64 `json:"monthly_total"`
	YearlyEstimate int64 `json:"yearly_estimate"`

	ByCategory []CategorySummary `json:"by_category"`
	Bars       []MonthAmount     `json:"bars"`
}

type CategorySummary struct {
	Category string `json:"category"`
	Total    int64  `json:"total"`
}

type MonthAmount struct {
	Label  string `json:"label"`
	Amount int64  `json:"amount"`
}

type PaymentHistory struct {
	UserID         int64
	SubscriptionID int64

	Price          int64
	Period         int64
	NextPaymentRaw string
	Status         bool
}
