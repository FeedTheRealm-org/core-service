package session

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateToken_NotExpired(t *testing.T) {
	manager := NewJWTManager("my-secret-key", time.Hour)

	email := "user@example.com"

	token, err := manager.GenerateToken(email)
	require.NoError(t, err)

	err = manager.IsValidateToken(token, time.Now().Add(time.Minute))
	require.NoError(t, err)
}

func TestValidateToken_Expired(t *testing.T) {
	manager := NewJWTManager("my-secret-key", time.Minute)

	email := "user@example.com"

	token, err := manager.GenerateToken(email)
	require.NoError(t, err)

	err = manager.IsValidateToken(token, time.Now().Add(time.Hour))
	require.Error(t, err)
	_, isExpiredErr := err.(*JWTExpiredTokenError)
	assert.True(t, isExpiredErr, "expected JWTExpiredTokenError")
}
