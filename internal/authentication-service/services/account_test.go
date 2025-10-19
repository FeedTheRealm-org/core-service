package services_test

import (
	"testing"

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
	password := "hashedpassword123"

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
