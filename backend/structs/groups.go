package structs

type Group struct {
	ID      int64 `json:"id"`
	OwnerID int64 `json:"owner_id"`

	Name      string `json:"name"`
	Type      string `json:"type"`
	Price     int64  `json:"price"`
	Period    int64  `json:"period"`
	InviteURL string `json:"invite_url"`
	Notes     string `json:"notes"`

	Members         []GroupMember `json:"members"`
	Services        []string      `json:"services"`
	SubscriptionIDs []int64       `json:"subscription_ids,omitempty"`
}

type GroupMember struct {
	UserID int64 `json:"user_id"`

	Name  string `json:"name"`
	Owner bool   `json:"owner"`
}

type JoinGroupRequest struct {
	UserID    int64  `json:"user_id"`
	InviteURL string `json:"invite_url"`
}
