package dtos

type CreateAccountRequestDTO struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginAccountRequestDTO struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type CreateAccountResponseDTO struct {
	Email string `json:"email"`
}

type LoginAccountResponseDTO struct {
	Token string `json:"token"`
}
