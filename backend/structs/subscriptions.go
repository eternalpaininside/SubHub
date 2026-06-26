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
