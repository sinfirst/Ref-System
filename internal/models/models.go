package models

type User struct {
	Username string `json:"login"`
	Password string `json:"password,omitempty"`
}

type OrderResponce struct {
	Order   string  `json:"order"`
	Status  string  `json:"status"`
	Accrual float64 `json:"accrual"`
}
