package hashing

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHashPassword(t *testing.T) {
	password := "SecurePass123!"

	hashedPassword, err := HashPassword(password)

	require.NoError(t, err)
	assert.NotEmpty(t, hashedPassword)
	assert.NotEqual(t, password, hashedPassword)
}

func TestVerifyPassword(t *testing.T) {
	password := "SecurePass123!"
	wrongPassword := "WrongPass456!"

	hashedPassword, err := HashPassword(password)
	require.NoError(t, err)

	assert.True(t, VerifyPassword(hashedPassword, password))
	assert.False(t, VerifyPassword(hashedPassword, wrongPassword))
}
