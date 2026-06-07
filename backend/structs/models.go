package structs

type Subscription struct {
	ID          int64  `json:"id"`
	UserID      int64  `json:"user_id"`
	Name        string `json:"name"`
	Category    string `json:"category"`
	Price       int64  `json:"price"`
	Period      int64  `json:"period"`
	NextPayment string `json:"next_payment"`
	Link        string `json:"link"`
	Status      bool   `json:"status"`
	Comment     string `json:"comment"`
	PlanType    string `json:"plan_type"`
}

type User struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	Password  string `json:"-"`
	TgID      string `json:"tg_id"`
	CreatedAt string `json:"created_at,omitempty"`
}

type AuthRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type AnalyticsResponse struct {
	ActiveCount    int64             `json:"active_count"`
	ExpiringCount  int64             `json:"expiring_count"`
	MonthlyTotal   int64             `json:"monthly_total"`
	YearlyEstimate int64             `json:"yearly_estimate"`
	ByCategory     []CategorySummary `json:"by_category"`
	Bars           []MonthAmount     `json:"bars"`
}

type CategorySummary struct {
	Category string `json:"category"`
	Total    int64  `json:"total"`
}

type MonthAmount struct {
	Label  string `json:"label"`
	Amount int64  `json:"amount"`
}

type ProfileResponse struct {
	User  User         `json:"user"`
	Stats ProfileStats `json:"stats"`
}

type ProfileStats struct {
	ActiveSubscriptions int64 `json:"active_subscriptions"`
	GroupCount          int64 `json:"group_count"`
	MonthlySpend        int64 `json:"monthly_spend"`
}

type ProfilePreference struct {
	Title   string `json:"title"`
	Text    string `json:"text"`
	Enabled bool   `json:"enabled"`
}

type ConnectedApp struct {
	Name   string `json:"name"`
	Status string `json:"status"`
	Style  string `json:"style"`
}

type Group struct {
	ID              int64         `json:"id"`
	OwnerID         int64         `json:"owner_id"`
	Name            string        `json:"name"`
	Type            string        `json:"type"`
	Price           int64         `json:"price"`
	Period          int64         `json:"period"`
	InviteURL       string        `json:"invite_url"`
	Notes           string        `json:"notes"`
	Members         []GroupMember `json:"members"`
	Services        []string      `json:"services"`
	SubscriptionIDs []int64       `json:"subscription_ids,omitempty"`
}

type GroupMember struct {
	UserID int64  `json:"user_id"`
	Name   string `json:"name"`
	Owner  bool   `json:"owner"`
}

type JoinGroupRequest struct {
	UserID    int64  `json:"user_id"`
	InviteURL string `json:"invite_url"`
}

type PaymentHistorySeed struct {
	UserID         int64
	SubscriptionID int64
	Price          int64
	Period         int64
	NextPaymentRaw string
	Status         bool
}
