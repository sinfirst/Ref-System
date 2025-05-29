package pg

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose"
	"github.com/sinfirst/Ref-System/internal/config"
	"github.com/sinfirst/Ref-System/internal/middleware/logging"
	"github.com/sinfirst/Ref-System/internal/models"
)

type PGDB struct {
	logger *logging.Logger
	db     *pgxpool.Pool
}

func NewPGDB(conf config.Config, logger *logging.Logger) *PGDB {
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

func (p *PGDB) CheckUsernameExists(ctx context.Context, username string) (bool, error) {
	var exists bool
	err := p.db.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM users WHERE username = $1
		)
	`, username).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("checking user existence: %w", err)
	}
	return exists, nil
} //

func (p *PGDB) AddUserToDB(ctx context.Context, username, password string) error {
	var insertedUser string

	query := `
		INSERT INTO users (username, user_password)
		VALUES ($1, $2)
		ON CONFLICT (username) DO UPDATE SET username = EXCLUDED.username
		RETURNING username
	`
	err := p.db.QueryRow(ctx, query, username, password).Scan(&insertedUser)

	if err != nil {
		return err
	}

	return nil
}

func (p *PGDB) GetUserPassword(ctx context.Context, username string) (string, error) {
	var password string

	query := `SELECT user_password FROM users WHERE username = $1`
	row := p.db.QueryRow(ctx, query, username)
	err := row.Scan(&password)
	if err != nil {
		return "", err
	}
	return password, nil
}

func (p *PGDB) GetOrderAndUser(ctx context.Context, order string) (string, string, error) {
	var userOrder string
	var username string

	query := `SELECT number, username FROM orders WHERE number = $1`
	row := p.db.QueryRow(ctx, query, order)
	err := row.Scan(&userOrder, &username)

	return userOrder, username, err
}

func (p *PGDB) AddOrderToDB(ctx context.Context, order string, username string) error {
	query := `INSERT INTO orders (number, uploaded_at, username)
				VALUES ($1, $2, $3) ON CONFLICT (number) DO NOTHING`
	_, err := p.db.Exec(ctx, query, order, time.Now(), username)

	return err
}

func (p *PGDB) UpdateStatus(ctx context.Context, newStatus, order, user string) error {
	query := `UPDATE orders SET status = $1
			WHERE number = $2`
	result, err := p.db.Exec(ctx, query, newStatus, order)

	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("0 rows affected")
	}

	return nil
}

func (p *PGDB) UpdateUserBalance(ctx context.Context, user string, accrual, withdrawn float64) error {
	query := `UPDATE users SET accrual = $1, withdrawn = $2
			WHERE username = $3`
	result, err := p.db.Exec(ctx, query, int(accrual*100), int(withdrawn*1000), user)

	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("0 rows affected")
	}

	return nil
}

func (p *PGDB) GetUserOrders(ctx context.Context, user string) ([]models.Order, error) {
	var orders []models.Order
	query := `SELECT number, status, uploaded_at 
		FROM orders WHERE username = $1`
	rows, err := p.db.Query(ctx, query, user)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var o models.Order

		err := rows.Scan(&o.Number, &o.Status, &o.UploadAt)
		if err != nil {
			return nil, err
		}

		orders = append(orders, o)
	}

	return orders, nil
}

func (p *PGDB) GetUserBalance(ctx context.Context, user string) (models.UserBalance, error) {
	var balance models.UserBalance
	query := `SELECT accrual, withdrawn 
		FROM users WHERE username = $1`
	row := p.db.QueryRow(ctx, query, user)
	err := row.Scan(&balance.Current, &balance.Withdrawn)

	if err != nil {
		return models.UserBalance{}, err
	}

	balance = models.UserBalance{
		Current:   balance.Current / 100,
		Withdrawn: balance.Withdrawn / 1000,
	}

	return balance, nil
}

func (p *PGDB) SetUserWithdrawn(ctx context.Context, orderNum, user string, withdrawn float64) error {
	query := `INSERT INTO withdrawals (orderNum, sum, precessed_at, username)
				VALUES ($1, $2, $3, $4) ON CONFLICT (orderNum) DO NOTHING`
	_, err := p.db.Exec(ctx, query, orderNum, withdrawn*100, time.Now(), user)

	return err
}

func (p *PGDB) GetUserWithdrawns(ctx context.Context, user string) ([]models.UserWithdrawal, error) {
	var UserWithdrawals []models.UserWithdrawal
	query := `SELECT orderNum, sum, precessed_at 
				FROM withdrawals WHERE username = $1`
	rows, err := p.db.Query(ctx, query, user)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var o models.UserWithdrawal

		err := rows.Scan(&o.OrderNum, &o.Sum, &o.ProcessedAt)
		if err != nil {
			return nil, err
		}

		UserWithdrawals = append(UserWithdrawals, models.UserWithdrawal{
			Sum:         o.Sum / 100,
			OrderNum:    o.OrderNum,
			ProcessedAt: o.ProcessedAt,
		})
	}

	return UserWithdrawals, nil
}

func InitMigrations(conf config.Config, logger *logging.Logger) error {
	if conf.DatabaseDsn == "" {
		return fmt.Errorf("DB url is not set")
	}

	db, err := sql.Open("pgx", conf.DatabaseDsn)
	if err != nil {
		return err
	}
	defer db.Close()

	if err := goose.Up(db, "internal/storage/migrations"); err != nil {
		return err
	}

	logger.Logger.Infow("Migrations applied successfully")
	return nil
}
