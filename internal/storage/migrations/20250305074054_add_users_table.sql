-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS users (
    username TEXT NOT NULL PRIMARY KEY,
    user_password TEXT NOT NULL,
    accrual INT DEFAULT 0,
    withdrawn INT DEFAULT 0
);

DO $$ BEGIN
    CREATE TYPE order_status AS ENUM (
        'NEW',
        'PROCESSING',
        'INVALID',
        'PROCESSED'
    );
EXCEPTION
    WHEN duplicate_object THEN NULL; 
END $$;

CREATE TABLE IF NOT EXISTS orders (
    number TEXT NOT NULL PRIMARY KEY,
    status order_status NOT NULL DEFAULT 'NEW',
    uploaded_at TIMESTAMPTZ NOT NULL,
    username TEXT REFERENCES users(username) ON DELETE SET NULL
);

CREATE TABLE IF NOT EXISTS withdrawals (
    orderNum TEXT NOT NULL PRIMARY KEY,
    sum INT DEFAULT 0,
    precessed_at TIMESTAMPTZ NOT NULL,
    username TEXT REFERENCES users(username) ON DELETE SET NULL
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS orders;
DROP TYPE IF EXISTS order_status;
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS withdrawals;
-- +goose StatementEnd