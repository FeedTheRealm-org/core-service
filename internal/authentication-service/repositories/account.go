package repositories

import (
	"context"

	"github.com/FeedTheRealm-org/core-service/config"
)

type AccountNotFoundError struct{}

func (e *AccountNotFoundError) Error() string {
	return "Account not found"
}

type accountRepository struct {
	conf *config.Config
	db   *config.DB
}

func NewAccountRepository(conf *config.Config, db *config.DB) (AccountRepository, error) {
	return &accountRepository{
		conf: conf,
		db:   db,
	}, nil
}

func (ar *accountRepository) GetAccountByEmail(email string) (*User, error) {
	var u User
	var id interface{}
	var createdAt interface{}

	row := ar.db.Conn.QueryRow(context.Background(),
		`SELECT id, email, password_hash, created_at
		 FROM accounts
		 WHERE email = $1`, email)

	if err := row.Scan(&id, &u.Email, &u.PasswordHash, &createdAt); err != nil {
		return nil, &AccountNotFoundError{}
	}

	return &u, nil
}

func (ar *accountRepository) CreateAccount(u *User) error {
	var id interface{}
	var createdAt interface{}

	row := ar.db.Conn.QueryRow(context.Background(),
		`INSERT INTO accounts (email, password_hash)
		 VALUES ($1, $2)
		 RETURNING id, created_at`, u.Email, u.PasswordHash)

	if err := row.Scan(&id, &createdAt); err != nil {
		return err
	}

	return nil
}
