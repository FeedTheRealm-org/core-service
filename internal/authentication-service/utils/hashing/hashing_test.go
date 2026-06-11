package hashing

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHashPassword_ReturnsHash(t *testing.T) {
	password := "myPassword123"

	hash, err := HashPassword(password)

	assert.NoError(t, err)
	assert.NotEmpty(t, hash)
	assert.NotEqual(t, password, hash)
}

func TestHashPassword_EmptyPassword(t *testing.T) {
	hash, err := HashPassword("")

	assert.NoError(t, err)
	assert.NotEmpty(t, hash)
}

func TestHashPassword_LongPassword(t *testing.T) {
	password := "thisIsAVeryLongPasswordThatExceedsTypicalLength1234567890!@#$%"

	hash, err := HashPassword(password)

	assert.NoError(t, err)
	assert.NotEmpty(t, hash)
}

func TestVerifyPassword_CorrectPassword(t *testing.T) {
	password := "correctPassword"
	hash, err := HashPassword(password)
	require.NoError(t, err)

	result := VerifyPassword(hash, password)

	assert.True(t, result)
}

func TestVerifyPassword_IncorrectPassword(t *testing.T) {
	password := "correctPassword"
	hash, err := HashPassword(password)
	require.NoError(t, err)

	result := VerifyPassword(hash, "wrongPassword")

	assert.False(t, result)
}

func TestVerifyPassword_EmptyHash(t *testing.T) {
	result := VerifyPassword("", "anything")

	assert.False(t, result)
}

func TestVerifyPassword_InvalidHash(t *testing.T) {
	result := VerifyPassword("notAValidHash", "password")

	assert.False(t, result)
}

func TestVerifyPassword_EmptyPassword(t *testing.T) {
	hash, err := HashPassword("")
	require.NoError(t, err)

	result := VerifyPassword(hash, "")

	assert.True(t, result)
}

func TestHashPassword_SamePasswordDifferentHashes(t *testing.T) {
	password := "samePassword"

	hash1, err1 := HashPassword(password)
	hash2, err2 := HashPassword(password)

	require.NoError(t, err1)
	require.NoError(t, err2)

	assert.NotEqual(t, hash1, hash2)

	assert.True(t, VerifyPassword(hash1, password))
	assert.True(t, VerifyPassword(hash2, password))
}

func TestHashPassword_ErrorHandling(t *testing.T) {
	hash, err := HashPassword("test")

	assert.NoError(t, err)
	assert.NotEmpty(t, hash)
}
