package repositories

type User struct {
	Email        string
	PasswordHash string
}

type AccountRepository interface {
	GetAccountByEmail(email string) (*User, error)
	CreateAccount(user *User) error
}
