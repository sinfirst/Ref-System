package main

import (
	"database/sql"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose"
	"github.com/sinfirst/Ref-System/internal/config"
	"github.com/sinfirst/Ref-System/internal/middleware/logging"
)

func main() {
	conf := config.NewConfig()
	logger := logging.NewLogger()

	if conf.DatabaseDsn == "" {
		logger.Logger.Fatalw("DB url is not set")
	}

	db, err := sql.Open("pgx", conf.DatabaseDsn)
	if err != nil {
		logger.Logger.Fatalw("Failed to open DB: ", err)
	}
	defer db.Close()

	if err := goose.Up(db, "internal/storage/migrations"); err != nil {
		logger.Logger.Fatalw("failed to apply migrations", err)
	}

	logger.Logger.Infow("Migrations applied successfully")
}
