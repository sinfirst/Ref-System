package app

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/sinfirst/Ref-System/internal/config"
	"github.com/sinfirst/Ref-System/internal/functions"
	"github.com/sinfirst/Ref-System/internal/middleware/auth"
	"github.com/sinfirst/Ref-System/internal/middleware/logging"
	"github.com/sinfirst/Ref-System/internal/models"
	"golang.org/x/crypto/bcrypt"
)

type Storage interface {
	CheckUsernameExists(ctx context.Context, username string) (bool, error)
	AddUserToDB(ctx context.Context, username, password string) error
	GetUserPassword(ctx context.Context, username string) (string, error)
	GetOrderAndUser(ctx context.Context, order string) (string, string, error)
	AddOrderToDB(ctx context.Context, order string, username string) error
	UpdateStatus(ctx context.Context, newStatus, order, user string) error
	UpdateUserBalance(ctx context.Context, user string, accrual, withdrawn float64) error
	GetUserOrders(ctx context.Context, user string) ([]models.Order, error)
	GetUserBalance(ctx context.Context, user string) (models.UserBalance, error)
	SetUserWithdrawn(ctx context.Context, orderNum, user string, withdrawn float64) error
	GetUserWithdrawns(ctx context.Context, user string) ([]models.UserWithdrawal, error)
}

type App struct {
	storage Storage
	config  config.Config
	logger  *logging.Logger
	pollCh  chan models.TypeForChannel
}

func NewApp(storage Storage, config config.Config, logger *logging.Logger, pollCh chan models.TypeForChannel) *App {
	return &App{storage: storage, config: config, logger: logger, pollCh: pollCh}
}

func (a *App) Register(w http.ResponseWriter, r *http.Request) {
	var user models.User

	err := json.NewDecoder(r.Body).Decode(&user)

	if err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	exist, err := a.storage.CheckUsernameExists(r.Context(), user.Username)
	if err != nil {
		a.logger.Logger.Errorf("err: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if exist {
		http.Error(w, "username already used, try choose another", http.StatusConflict)
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		a.logger.Logger.Errorf("err: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = a.storage.AddUserToDB(r.Context(), user.Username, string(hashedPassword))
	if err != nil {
		a.logger.Logger.Errorf("err: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	token, err := auth.BuildJWTString(user.Username)
	if err != nil {
		a.logger.Logger.Errorf("err: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	cookie := &http.Cookie{
		Name:     "token",
		Value:    token,
		HttpOnly: true,
	}

	http.SetCookie(w, cookie)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}

func (a *App) Login(w http.ResponseWriter, r *http.Request) {
	var user models.User

	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	password, err := a.storage.GetUserPassword(r.Context(), user.Username)
	if err != nil {
		a.logger.Logger.Errorf("err: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	err = bcrypt.CompareHashAndPassword([]byte(password), []byte(user.Password))
	if err != nil {
		http.Error(w, "wrong login or password", http.StatusUnauthorized)
		return
	}

	token, err := auth.BuildJWTString(user.Username)
	if err != nil {
		a.logger.Logger.Errorf("err: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	cookie := &http.Cookie{
		Name:     "token",
		Value:    token,
		HttpOnly: true,
	}
	http.SetCookie(w, cookie)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}
func (a *App) OrdersIn(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "invalid order number format", http.StatusUnprocessableEntity)
		return
	}

	if len(body) == 0 {
		http.Error(w, "number of order is empty", http.StatusBadRequest)
		return
	}

	if !functions.LuhnCheck(string(body)) {
		http.Error(w, "Failed Luhn algo", http.StatusUnprocessableEntity)
		return
	}
	ctxK := models.CtxKey("userName")
	value := r.Context().Value(ctxK)
	user := fmt.Sprintf("%v", value)
	order, username, err := a.storage.GetOrderAndUser(r.Context(), string(body))
	if err == nil && order == string(body) {
		if user == username {
			http.Error(w, "order already exist", http.StatusOK)
			return
		} else {
			http.Error(w, "order upload another user", http.StatusConflict)
			return
		}
	}

	err = a.storage.AddOrderToDB(r.Context(), string(body), user)
	if err != nil {
		a.logger.Logger.Errorf("err: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	a.pollCh <- models.TypeForChannel{User: user, OrderNum: string(body)}
	w.WriteHeader(http.StatusAccepted)
}
func (a *App) OrdersInfo(w http.ResponseWriter, r *http.Request) {
	var ordersFloat []models.OrderFloat
	ctxK := models.CtxKey("userName")
	value := r.Context().Value(ctxK)
	user := fmt.Sprintf("%v", value)

	orders, err := a.storage.GetUserOrders(r.Context(), user)
	if err != nil {
		a.logger.Logger.Errorf("err: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	userBalance, err := a.storage.GetUserBalance(r.Context(), user)
	if err != nil {
		a.logger.Logger.Errorf("err: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	for _, order := range orders {
		ordersFloat = append(ordersFloat, models.OrderFloat{
			Number:   order.Number,
			Status:   order.Status,
			Accrual:  float64(userBalance.Current),
			UploadAt: order.UploadAt,
		})
	}

	if len(ordersFloat) == 0 {
		http.Error(w, "order list is empty", http.StatusNoContent)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(ordersFloat)
	if err != nil {
		a.logger.Logger.Errorf("err: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
func (a *App) GetBalance(w http.ResponseWriter, r *http.Request) {
	ctxK := models.CtxKey("userName")
	value := r.Context().Value(ctxK)
	user := fmt.Sprintf("%v", value)

	balance, err := a.storage.GetUserBalance(r.Context(), user)
	if err != nil {
		a.logger.Logger.Errorf("err: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(balance)
	if err != nil {
		a.logger.Logger.Errorf("err: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
func (a *App) Withdraw(w http.ResponseWriter, r *http.Request) {
	var data models.UserWithdrawal
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "invalid request ", http.StatusUnprocessableEntity)
		return
	}

	if len(body) == 0 {
		http.Error(w, "request is empty", http.StatusBadRequest)
		return
	}

	if err := json.Unmarshal(body, &data); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	ctxK := models.CtxKey("userName")
	value := r.Context().Value(ctxK)
	user := fmt.Sprintf("%v", value)

	userBalance, err := a.storage.GetUserBalance(r.Context(), user)

	if err != nil {
		a.logger.Logger.Errorf("err: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if userBalance.Current < data.Sum {
		http.Error(w, "Not enough balance", http.StatusPaymentRequired)
		return
	}

	finalSum := userBalance.Current - data.Sum
	err = a.storage.UpdateUserBalance(r.Context(), user, finalSum, data.Sum)
	if err != nil {
		a.logger.Logger.Errorf("err: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = a.storage.SetUserWithdrawn(r.Context(), data.OrderNum, user, data.Sum)
	if err != nil {
		a.logger.Logger.Errorf("err: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
func (a *App) WithdrawInfo(w http.ResponseWriter, r *http.Request) {
	ctxK := models.CtxKey("userName")
	value := r.Context().Value(ctxK)
	user := fmt.Sprintf("%v", value)

	withdrawns, err := a.storage.GetUserWithdrawns(r.Context(), user)

	if err != nil {
		a.logger.Logger.Errorf("err: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if len(withdrawns) == 0 {
		http.Error(w, "list withdraw is empty", http.StatusNoContent)
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(withdrawns)
	if err != nil {
		a.logger.Logger.Errorf("err: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
