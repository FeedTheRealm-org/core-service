package credential_validation

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsValidEmail_ValidEmail(t *testing.T) {
	assert.True(t, IsValidEmail("user@example.com"))
}

func TestIsValidEmail_NoDomain(t *testing.T) {
	assert.False(t, IsValidEmail("user@"))
	assert.False(t, IsValidEmail("userexample.com"))
	assert.False(t, IsValidEmail("user@example"))
}

func TestIsValidEmail_InvalidDomain(t *testing.T) {
	assert.False(t, IsValidEmail("user@example.c"))
	assert.False(t, IsValidEmail("user@exa_mple.com"))
}

func TestIsValidPassword_ValidPassword(t *testing.T) {
	assert.True(t, IsValidPassword("Password1"))
	assert.True(t, IsValidPassword("abcDEF123"))
}

func TestIsValidPassword_TooShort(t *testing.T) {
	assert.False(t, IsValidPassword("Pwd1"))
	assert.False(t, IsValidPassword("1234567"))
}

func TestIsValidPassword_NoLetter(t *testing.T) {
	assert.False(t, IsValidPassword("12345678"))
}

func TestIsValidPassword_NoNumber(t *testing.T) {
	assert.False(t, IsValidPassword("Password"))
}
