package repositories

import (
	"context"

	"github.com/FeedTheRealm-org/core-service/config"
	"github.com/jackc/pgx/v5"
)

type accountRepository struct {
	conf *config.Config
	conn *pgx.Conn
}

func NewAccountRepository(conf *config.Config) (AccountRepository, error) {
	conn, err := conf.Dbc.GetConnectionToDatabase()
	if err != nil {
		return nil, err
	}

	return &accountRepository{
		conf: conf,
		conn: conn,
	}, nil
}

func (ar *accountRepository) GetAccountByEmail(email string) (*User, error) {
	var u User
	var id interface{}
	var createdAt interface{}

	row := ar.conn.QueryRow(context.Background(),
		`SELECT id, email, password_hash, created_at
		 FROM accounts
		 WHERE email = $1`, email)

	if err := row.Scan(&id, &u.Email, &u.PasswordHash, &createdAt); err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &u, nil
}

func (ar *accountRepository) CreateAccount(u *User) error {
	var id interface{}
	var createdAt interface{}

	row := ar.conn.QueryRow(context.Background(),
		`INSERT INTO accounts (email, password_hash)
		 VALUES ($1, $2)
		 RETURNING id, created_at`, u.Email, u.PasswordHash)

	if err := row.Scan(&id, &createdAt); err != nil {
		return err
	}

	return nil
}
