package pg

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sinfirst/Ref-System/internal/config"
	"github.com/sinfirst/Ref-System/internal/middleware/logging"
)

type PGDB struct {
	logger logging.Logger
	db     *pgxpool.Pool
}

func NewPGDB(conf config.Config, logger logging.Logger) *PGDB {
	db, err := pgxpool.New(context.Background(), conf.DatabaseDsn)

	if err != nil {
		logger.Logger.Errorw("Problem with connecting to db: ", err)
		return nil
	}

	err = db.Ping(context.Background())

	if err != nil {
		logger.Logger.Errorw("Problem with ping to db: ", err)
		return nil
	}

	return &PGDB{logger: logger, db: db}
}

func (p *PGDB) CheckUsernameLogin(ctx context.Context, username string) bool {
	var user string

	query := `SELECT username FROM users WHERE username = $1`
	row := p.db.QueryRow(ctx, query, username)
	row.Scan(&user)

	return user != ""
}

func (p *PGDB) AddUserToDB(ctx context.Context, username, password string) error {
	var insertedUser string
	query := `INSERT INTO users (username, user_password)
				VALUES ($1, $2) ON CONFLICT (username) DO NOTHING
				RETURNING username`

	err := p.db.QueryRow(ctx, query, username, password).Scan(&insertedUser)

	if err != nil {
		return err
	}

	return nil
}
