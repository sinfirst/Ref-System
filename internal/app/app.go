package app

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/sinfirst/Ref-System/internal/config"
	"github.com/sinfirst/Ref-System/internal/functions"
	"github.com/sinfirst/Ref-System/internal/middleware/auth"
	"github.com/sinfirst/Ref-System/internal/middleware/logging"
	"github.com/sinfirst/Ref-System/internal/models"
	"github.com/sinfirst/Ref-System/internal/storage/pg"
	"github.com/sinfirst/Ref-System/internal/worker"
	"golang.org/x/crypto/bcrypt"
)

type App struct {
	storage pg.PGDB
	config  config.Config
	logger  logging.Logger
}

func NewApp(storage pg.PGDB, config config.Config, logger logging.Logger) *App {
	app := &App{storage: storage, config: config, logger: logger}
	return app
}

func (a *App) Register(w http.ResponseWriter, r *http.Request) {
	var user models.User
	err := json.NewDecoder(r.Body).Decode(&user)

	if err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	if exist := a.storage.CheckUsernameLogin(r.Context(), user.Username); exist {
		http.Error(w, "username already used, try choose another", http.StatusConflict)
		return
	}

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
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
	w.Write([]byte("successful registration"))
}

func (a *App) Login(w http.ResponseWriter, r *http.Request) {
	var user models.User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	password := a.storage.GetUserPassword(r.Context(), user.Username)
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

	cookie, _ := r.Cookie("token")
	user := auth.GetUsername(cookie.Value)
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

	worker.PollOrderStatus(r.Context(), string(body), user, a.config.AccurualSystemAddress, a.storage)
	w.WriteHeader(http.StatusAccepted)
}
func (a *App) OrdersInfo(w http.ResponseWriter, r *http.Request) {

}
func (a *App) GetBalance(w http.ResponseWriter, r *http.Request) {

}
func (a *App) WithDraw(w http.ResponseWriter, r *http.Request) {

}
func (a *App) WithDrawInfo(w http.ResponseWriter, r *http.Request) {

}
