package services_test

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
)

func TestAccount_CreateAccount(t *testing.T) {
	email := "new@example.com"
	password := "password123"

	user, err := service.CreateAccount(email, password)
	assert.Nil(t, err, "expected no error on account creation")
	assert.NotNil(t, user, "expected user to be created")
	assert.Equal(t, email, user.Email, "expected email to match")
}

func TestAccount_GetUserByEmail(t *testing.T) {
	email := "existing@example.com"
	password := "password123"

	_, err := service.CreateAccount(email, password)
	assert.Nil(t, err, "expected no error on account creation")

	user, err := service.GetUserByEmail(email)
	assert.Nil(t, err, "expected no error on getting user by email")
	assert.NotNil(t, user, "expected user to be found")
	assert.Equal(t, email, user.Email, "expected email to match")
}

func TestAccount_GetUserByEmail_NotFound(t *testing.T) {
	email := "notfound@example.com"

	user, err := service.GetUserByEmail(email)
	assert.NotNil(t, err, "expected error on getting user by email")
	assert.Nil(t, user, "expected no user to be found")
}

func TestAccount_CreateAccount_AlreadyExists(t *testing.T) {
	email := "existing@example.com"
	password := "password123"

	user, err := service.CreateAccount(email, password)
	assert.NotNil(t, err, "expected error on account creation")
	assert.Error(t, err, "Account already exists")
	assert.Nil(t, user, "expected no user to be created")
}

func TestAccount_CreateAccount_EmptyEmail(t *testing.T) {
	email := ""
	password := "password123"

	user, err := service.CreateAccount(email, password)
	assert.NotNil(t, err, "expected error when creating account with empty email")
	assert.Error(t, err, "Empty email")
	assert.Nil(t, user, "expected no user to be created with empty email")
}

func TestAccount_CreateAccount_InvalidEmail_EmptyDomain(t *testing.T) {
	email := "user@"
	password := "password123"

	user, err := service.CreateAccount(email, password)
	assert.NotNil(t, err, "expected error when creating account with empty email domain")
	assert.Error(t, err, "Invalid email")
	assert.Nil(t, user, "expected no user to be created with empty email domain")
}

func TestAccount_CreateAccount_InvalidEmail_InvalidDomain(t *testing.T) {
	email := "user@invalid_domain"
	password := "password123"

	user, err := service.CreateAccount(email, password)
	assert.NotNil(t, err, "expected error when creating account with invalid email domain")
	assert.Error(t, err, "Invalid email")
	assert.Nil(t, user, "expected no user to be created with invalid email domain")
}

func TestAccount_CreateAccount_EmptyPassword(t *testing.T) {
	email := "user@example.com"
	password := ""

	user, err := service.CreateAccount(email, password)
	assert.NotNil(t, err, "expected error when creating account with empty password")
	assert.Error(t, err, "Empty password")
	assert.Nil(t, user, "expected no user to be created with empty password")
}

func TestAccount_CreateAccount_InvalidPassword_TooShort(t *testing.T) {
	email := "user@example.com"
	password := "123"

	user, err := service.CreateAccount(email, password)
	assert.NotNil(t, err, "expected error when creating account with too short password")
	assert.Error(t, err, "Password too short")
	assert.Nil(t, user, "expected no user to be created with too short password")
}

func TestAccount_CreateAccount_InvalidPassword_NonChars(t *testing.T) {
	email := "user@example.com"
	password := "12345678"

	user, err := service.CreateAccount(email, password)
	assert.NotNil(t, err, "expected error when creating account with password non chars")
	assert.Error(t, err, "Password must contain at least one letter")
	assert.Nil(t, user, "expected no user to be created with password non chars")
}

func TestAccount_CreateAccount_InvalidPassword_NonNumbers(t *testing.T) {
	email := "user@example.com"
	password := "PasswordOnly"

	user, err := service.CreateAccount(email, password)
	assert.NotNil(t, err, "expected error when creating account with password missing numbers")
	assert.Error(t, err, "Password must contain at least one number")
	assert.Nil(t, user, "expected no user to be created with password missing numbers")
}

func TestAccount_LoginAccount(t *testing.T) {
	email := "existing@example.com"
	password := "password123"

	token, err := service.LoginAccount(email, password)
	assert.Nil(t, err, "expected no error on login")
	assert.NotEmpty(t, token, "expected token to be returned")
}

func TestAccount_LoginAccount_InvalidCredentials(t *testing.T) {
	email := "existing@example.com"
	wrongPassword := "wrongpassword123"

	token, err := service.LoginAccount(email, wrongPassword)
	assert.NotNil(t, err, "expected error on login with invalid credentials")
	assert.Empty(t, token, "expected no token to be returned")
}

func TestAccount_LoginAccount_NonExistentEmail(t *testing.T) {
	wrongEmail := "nonexistent@example.com"
	password := "password123"

	token, err := service.LoginAccount(wrongEmail, password)
	assert.NotNil(t, err, "expected error on login with non-existent email")
	assert.Empty(t, token, "expected no token to be returned")
}

func TestAccount_ValidateSessionToken(t *testing.T) {
	email := "existing@example.com"
	password := "password123"

	token, err := service.LoginAccount(email, password)
	assert.Nil(t, err, "expected no error on login")
	assert.NotEmpty(t, token, "expected token to be returned")

	err = service.ValidateSessionToken(token)
	assert.Nil(t, err, "expected no error on validating session token")
}

func TestAccount_ValidateSessionToken_InvalidToken(t *testing.T) {
	invalidToken := "invalid.token.string"

	err := service.ValidateSessionToken(invalidToken)
	assert.NotNil(t, err, "expected error on validating invalid session token")
}

func TestAccount_ValidateSessionToken_ExpiredToken(t *testing.T) {
	email := "existing@example.com"
	password := "password123"

	token, err := service.LoginAccount(email, password)
	assert.Nil(t, err, "expected no error on login")
	assert.NotEmpty(t, token, "expected token to be returned")

	expiredToken := jwt.NewWithClaims(jwt.SigningMethodHS256, &jwt.MapClaims{
		"email":      email,
		"expires_at": time.Now().Add(-time.Hour).Unix(),
		"issued_at":  time.Now().Unix(),
	})

	expiredTokenString, err := expiredToken.SignedString([]byte("test_secret_key"))
	if err != nil {
		t.Fatalf("failed to sign token: %v", err)
	}

	err = service.ValidateSessionToken(expiredTokenString)
	assert.NotNil(t, err, "expected error on validating expired session token")
}
