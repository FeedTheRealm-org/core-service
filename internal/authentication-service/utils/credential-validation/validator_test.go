package credential_validation

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsValidEmail_ValidEmail_ReturnsNil(t *testing.T) {
	validEmails := []string{
		"user@example.com",
		"user.name@example.co.uk",
		"user+label@example.com",
		"user123@example.com",
		"user@subdomain.example.com",
	}

	for _, email := range validEmails {
		t.Run(email, func(t *testing.T) {
			err := IsValidEmail(email)
			assert.NoError(t, err)
		})
	}
}

func TestIsValidEmail_EmptyEmail_ReturnsEmptyEmailError(t *testing.T) {
	err := IsValidEmail("")

	assert.Error(t, err)
	assert.IsType(t, &EmptyEmailError{}, err)
	assert.Equal(t, "Empty email", err.Error())
}

func TestIsValidEmail_InvalidEmail_ReturnsInvalidEmailError(t *testing.T) {
	invalidEmails := []string{
		"user@",
		"userexample.com",
		"user@example",
		"user@example.c",
		"user@exa_mple.com",
		"user@.com",
		"@example.com",
	}

	for _, email := range invalidEmails {
		t.Run(email, func(t *testing.T) {
			err := IsValidEmail(email)
			assert.Error(t, err)
			assert.IsType(t, &InvalidEmailError{}, err)
			assert.Equal(t, "Invalid email", err.Error())
		})
	}
}

func TestIsValidPassword_ValidPassword_ReturnsNil(t *testing.T) {
	validPasswords := []string{
		"Password1",
		"Abcdefg1",
		"mypassword123",
		"12345678a",
		"a1b2c3d4",
		"Password12345678",
	}

	for _, password := range validPasswords {
		t.Run(password, func(t *testing.T) {
			err := IsValidPassword(password)
			assert.NoError(t, err)
		})
	}
}

func TestIsValidPassword_EmptyPassword_ReturnsEmptyPasswordError(t *testing.T) {
	err := IsValidPassword("")

	assert.Error(t, err)
	assert.IsType(t, &EmptyPasswordError{}, err)
	assert.Equal(t, "Empty password", err.Error())
}

func TestIsValidPassword_TooShort_ReturnsPasswordTooShortError(t *testing.T) {
	shortPasswords := []string{
		"Pwd1",
		"1234567",
		"a1b2c3d",
		"Pass1",
		"",
	}

	for _, password := range shortPasswords {
		t.Run(password, func(t *testing.T) {
			err := IsValidPassword(password)
			if password == "" {
				assert.IsType(t, &EmptyPasswordError{}, err)
			} else {
				assert.IsType(t, &PasswordTooShortError{}, err)
				assert.Equal(t, "Password too short", err.Error())
			}
		})
	}
}

func TestIsValidPassword_NoLetter_ReturnsPasswordNoLetterError(t *testing.T) {
	err := IsValidPassword("12345678")

	assert.Error(t, err)
	assert.IsType(t, &PasswordNoLetterError{}, err)
	assert.Equal(t, "Password must contain at least one letter", err.Error())
}

func TestIsValidPassword_NoNumber_ReturnsPasswordNoNumberError(t *testing.T) {
	err := IsValidPassword("Password")

	assert.Error(t, err)
	assert.IsType(t, &PasswordNoNumberError{}, err)
	assert.Equal(t, "Password must contain at least one number", err.Error())
}

func TestIsValidPassword_TooShortAndNoLetter(t *testing.T) {
	err := IsValidPassword("123")

	assert.Error(t, err)
	assert.IsType(t, &PasswordTooShortError{}, err)
}

func TestIsValidPassword_TooShortAndNoNumber(t *testing.T) {
	err := IsValidPassword("abc")

	assert.Error(t, err)
	assert.IsType(t, &PasswordTooShortError{}, err)
}
