package storage

import (
	"github.com/sinfirst/Ref-System/internal/app"
	"github.com/sinfirst/Ref-System/internal/config"
	"github.com/sinfirst/Ref-System/internal/middleware/logging"
	"github.com/sinfirst/Ref-System/internal/storage/pg"
)

func NewStorage(conf config.Config, logger *logging.Logger) app.Storage {
	return pg.NewPGDB(conf, logger)
}
