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

type VerifyAccountRequestDTO struct {
	Email string `json:"email"`
	Code  string `json:"code"`
}

type VerifyAccountResponseDTO struct {
	Email    string `json:"email"`
	Verified bool   `json:"verified"`
}
