package models

import (
	"context"
	"time"
)

type User struct {
	Username string `json:"login"`
	Password string `json:"password,omitempty"`
}

type OrderResponce struct {
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

type Storage interface {
	CheckUsernameExists(ctx context.Context, username string) (bool, error)
	AddUserToDB(ctx context.Context, username, password string) error
	GetUserPassword(ctx context.Context, username string) (string, error)
	GetOrderAndUser(ctx context.Context, order string) (string, string, error)
	AddOrderToDB(ctx context.Context, order string, username string) error
	UpdateStatus(ctx context.Context, newStatus, order, user string) error
	UpdateUserBalance(ctx context.Context, user string, accrual, withdrawn float64) error
	GetUserOrders(ctx context.Context, user string) ([]Order, error)
	GetUserBalance(ctx context.Context, user string) (UserBalance, error)
	SetUserWithdrawn(ctx context.Context, orderNum, user string, withdrawn float64) error
	GetUserWithdrawns(ctx context.Context, user string) ([]UserWithdrawal, error)
}
