package credential_validation

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsValidEmail_ValidEmail(t *testing.T) {
	err := IsValidEmail("user@example.com")
	assert.Nil(t, err)
	assert.NoError(t, err)
}

func TestIsValidEmail_InvalidEmails(t *testing.T) {
	err := IsValidEmail("")
	assert.NotNil(t, err)
	assert.Error(t, err, "Empty email")

	err = IsValidEmail("user@")
	assert.NotNil(t, err)
	assert.Error(t, err, "Invalid email")

	err = IsValidEmail("userexample.com")
	assert.NotNil(t, err)
	assert.Error(t, err, "Invalid email")

	err = IsValidEmail("user@example")
	assert.NotNil(t, err)
	assert.Error(t, err, "Invalid email")

	err = IsValidEmail("user@example.c")
	assert.NotNil(t, err)
	assert.Error(t, err, "Invalid email")

	err = IsValidEmail("user@exa_mple.com")
	assert.NotNil(t, err)
	assert.Error(t, err, "Invalid email")
}

func TestIsValidPassword_ValidPassword(t *testing.T) {
	err := IsValidPassword("Password1")
	assert.Nil(t, err)
	assert.NoError(t, err)

	err = IsValidPassword("Abcdefg1")
	assert.Nil(t, err)
	assert.NoError(t, err)
}

func TestIsValidPassword_TooShort(t *testing.T) {
	err := IsValidPassword("Pwd1")
	assert.NotNil(t, err)
	assert.Error(t, err, "Password too short")

	err = IsValidPassword("1234567")
	assert.NotNil(t, err)
	assert.Error(t, err, "Password too short")
}

func TestIsValidPassword_NoLetter(t *testing.T) {
	err := IsValidPassword("12345678")
	assert.NotNil(t, err)
	assert.Error(t, err, "Password must contain at least one letter")
}

func TestIsValidPassword_NoNumber(t *testing.T) {
	err := IsValidPassword("Password")
	assert.NotNil(t, err)
	assert.Error(t, err, "Password must contain at least one number")
}
