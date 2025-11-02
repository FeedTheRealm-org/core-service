package repositories

type User struct {
	Email        string
	PasswordHash string
	VerifyCode   string
}

type AccountRepository interface {
	GetAccountByEmail(email string) (*User, error)
	CreateAccount(user *User) error
	IsAccountVerified(email string) (bool, error)
	VerifyAccount(email string, code string) error
}
