package app

import (
	"net/http"

	"github.com/sinfirst/Ref-System/internal/config"
	"github.com/sinfirst/Ref-System/internal/storage/pg"
	"go.uber.org/zap"
)

type App struct {
	storage  pg.PGDB
	config   config.Config
	logger   zap.SugaredLogger
	deleteCh chan string
}

func NewApp(storage pg.PGDB, config config.Config, logger zap.SugaredLogger, deleteCh chan string) *App {
	app := &App{storage: storage, config: config, logger: logger, deleteCh: deleteCh}
	return app
}

func (a *App) Register(w http.ResponseWriter, r *http.Request) {

}

func (a *App) Login(w http.ResponseWriter, r *http.Request) {

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
