package repositories

import (
	"time"
	"context"

	"github.com/FeedTheRealm-org/core-service/config"
	"github.com/jackc/pgx/v5"
)

type AccountNotFoundError struct{}

type AccountNotVerifiedError struct{}

type AccountVerificationExpired struct{}

func (e *AccountNotFoundError) Error() string {
	return "Account not found"
}

func (e *AccountNotVerifiedError) Error() string {
	return "Account not verified"
}

func (e *AccountVerificationExpired) Error() string {
	return "Account verification has expired"
}

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
		return nil, &AccountNotFoundError{}
	}

	return &u, nil
}

func (ar *accountRepository) CreateAccount(u *User) error {
	var id interface{}
	var createdAt interface{}

	row := ar.conn.QueryRow(context.Background(),
		`INSERT INTO accounts (email, password_hash, verify_code, expiration_verify_code)
		 VALUES ($1, $2, $3, $4)
		 RETURNING id, created_at`, u.Email, u.PasswordHash, u.VerifyCode, u.Expiration)

	if err := row.Scan(&id, &createdAt); err != nil {
		return err
	}

	return nil
}

func (ar *accountRepository) IsAccountVerified(email string) (bool, error) {
	var verifyCode interface{}

	row := ar.conn.QueryRow(context.Background(),
		`SELECT verify_code
		 FROM accounts
		 WHERE email = $1`, email)

	if err := row.Scan(&verifyCode); err != nil {
		return false, &AccountNotFoundError{}
	}

	return verifyCode == nil, nil
}

func (ar *accountRepository) VerifyAccount(email string, code string, currentTime time.Time) error {
	var verifyCode interface{}
	var expiration interface{}

	row := ar.conn.QueryRow(context.Background(),
		`SELECT verify_code, expiration_verify_code
		 FROM accounts
		 WHERE email = $1`, email)

	if err := row.Scan(&verifyCode, &expiration); err != nil {
		return &AccountNotFoundError{}
	}

	if verifyCode != code {
		return &AccountNotVerifiedError{}
	}

	if currentTime.After(expiration.(time.Time)) {
		return &AccountVerificationExpired{}
	}

	_, err := ar.conn.Exec(context.Background(),
		`UPDATE accounts
		 SET verify_code = NULL
		 WHERE email = $1`, email)

	return err
}
