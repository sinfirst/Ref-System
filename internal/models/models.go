package models

import (
	"time"
)

type User struct {
	Username string `json:"login"`
	Password string `json:"password,omitempty"`
}

type OrderResponse struct {
	Order   string  `json:"order"`
	Status  string  `json:"status"`
	Accrual float64 `json:"accrual"`
}

type Order struct {
	Number   string    `json:"number"`
	Status   string    `json:"status"`
	Accrual  int       `json:"accrual,omitempty"`
	UploadAt time.Time `json:"upload_at"`
}

type UserBalance struct {
	Current   float64 `json:"current"`
	Withdrawn float64 `json:"withdrawn"`
}

type OrderFloat struct {
	Number   string    `json:"number"`
	Status   string    `json:"status"`
	Accrual  float64   `json:"accrual,omitempty"`
	UploadAt time.Time `json:"upload_at"`
}

type UserWithdrawal struct {
	OrderNum    string    `json:"order"`
	Sum         float64   `json:"sum"`
	ProcessedAt time.Time `json:"processed_at"`
}

type TypeForChannel struct {
	OrderNum string
	User     string
}
