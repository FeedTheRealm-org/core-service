package repositories

import (
	"time"
)

type User struct {
	Email        string
	PasswordHash string
	VerifyCode   string
	Expiration   time.Time
}

type AccountRepository interface {
	GetAccountByEmail(email string) (*User, error)
	CreateAccount(user *User) error
	IsAccountVerified(email string) (bool, error)
	VerifyAccount(email string, code string, currentTime time.Time) error
}
