package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/sinfirst/Ref-System/internal/app"
	"github.com/sinfirst/Ref-System/internal/config"
	"github.com/sinfirst/Ref-System/internal/middleware/logging"
	"github.com/sinfirst/Ref-System/internal/models"
	"github.com/sinfirst/Ref-System/internal/router"
	"github.com/sinfirst/Ref-System/internal/storage"
	"github.com/sinfirst/Ref-System/internal/storage/pg"
	"github.com/sinfirst/Ref-System/internal/worker"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	pollCh := make(chan models.TypeForChannel, 6)
	logger := logging.NewLogger()
	conf := config.NewConfig()
	stg := storage.NewStorage(conf, logger)
	db := pg.NewPGDB(conf, logger)
	app := app.NewApp(stg, conf, logger, pollCh)
	router := router.NewRouter(app, logger)
	worker := worker.NewPollWorker(ctx, conf.AccurualSystemAddress, db, pollCh)
	err := pg.InitMigrations(conf, logger)
	if err != nil {
		logger.Logger.Fatalw("Failed init migrations:", err)
	}
	server := &http.Server{Addr: conf.ServerAdress, Handler: router}

	go func() {
		logger.Logger.Infow("Starting server", "addr", conf.ServerAdress)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Logger.Fatalw("create server error: ", err)
		}
	}()
	<-ctx.Done()
	if err := server.Shutdown(context.Background()); err != nil {
		logger.Logger.Errorw("Server shutdown error", err)
	}
	worker.StopWorker()
	close(pollCh)
}
