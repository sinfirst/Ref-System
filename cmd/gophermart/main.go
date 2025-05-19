package main

import (
	"net/http"

	"github.com/sinfirst/Ref-System/internal/app"
	"github.com/sinfirst/Ref-System/internal/config"
	"github.com/sinfirst/Ref-System/internal/middleware/logging"
	"github.com/sinfirst/Ref-System/internal/router"
	"github.com/sinfirst/Ref-System/internal/storage/pg"
)

func main() {

	logger := logging.NewLogger()
	conf := config.NewConfig()
	db := pg.NewPGDB(conf, logger)
	app := app.NewApp(db, conf, logger)
	router := router.NewRouter(app, logger)
	err := pg.InitMigrations(conf, logger)
	if err != nil {
		logger.Logger.Fatalw("Failed init migrations:", err)
	}
	server := &http.Server{Addr: conf.ServerAdress, Handler: router}

	logger.Logger.Infow("Start server", "addr: ", server.Addr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Logger.Fatalw("create server err: ", err)
	}
}
