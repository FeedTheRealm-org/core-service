package dtos

type CreateAccountRequestDTO struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}
