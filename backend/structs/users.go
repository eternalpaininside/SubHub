package structs

type User struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	Password  string `json:"-"`
	CreatedAt string `json:"created_at,omitempty"`
}

type AuthRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
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
