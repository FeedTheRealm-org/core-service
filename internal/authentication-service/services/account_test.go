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
