package dtos

import "time"

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
	AccessToken string    `json:"access_token"`
	Id          string    `json:"id"`
	Email       string    `json:"email"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type CheckSessionResponseDTO struct {
	Message string `json:"message"`
}

type VerifyAccountRequestDTO struct {
	Email string `json:"email"`
	Code  string `json:"code"`
}

type VerifyAccountResponseDTO struct {
	Email    string `json:"email"`
	Verified bool   `json:"verified"`
}

type RefreshVerificationRequestDTO struct {
	Email string `json:"email"`
}

type RefreshVerificationResponseDTO struct {
	Email string `json:"email"`
}
