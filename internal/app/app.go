package app

import (
	"encoding/json"
	"net/http"

	"github.com/sinfirst/Ref-System/internal/config"
	"github.com/sinfirst/Ref-System/internal/middleware/auth"
	"github.com/sinfirst/Ref-System/internal/middleware/logging"
	"github.com/sinfirst/Ref-System/internal/models"
	"github.com/sinfirst/Ref-System/internal/storage/pg"
	"golang.org/x/crypto/bcrypt"
)

type App struct {
	storage  pg.PGDB
	config   config.Config
	logger   logging.Logger
	deleteCh chan string
}

func NewApp(storage pg.PGDB, config config.Config, logger logging.Logger, deleteCh chan string) *App {
	app := &App{storage: storage, config: config, logger: logger, deleteCh: deleteCh}
	return app
}

func (a *App) Register(w http.ResponseWriter, r *http.Request) {
	var user models.User
	err := json.NewDecoder(r.Body).Decode(&user)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if exist := a.storage.CheckUsernameLogin(r.Context(), user.Username); exist {
		w.WriteHeader(http.StatusConflict)
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
}

func (a *App) Login(w http.ResponseWriter, r *http.Request) {
	var user models.User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	password := a.storage.GetUserPassword(r.Context(), user.Username)
	err = bcrypt.CompareHashAndPassword([]byte(password), []byte(user.Password))
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
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

}
func (a *App) OrdersOut(w http.ResponseWriter, r *http.Request) {

}
func (a *App) GetBalance(w http.ResponseWriter, r *http.Request) {

}
func (a *App) WithDraw(w http.ResponseWriter, r *http.Request) {

}
func (a *App) WithDrawInfo(w http.ResponseWriter, r *http.Request) {

}
