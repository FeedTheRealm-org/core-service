package session

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIsValidateToken_Valid(t *testing.T) {
	manager := NewJWTManager("my-secret-key", time.Minute)
	email := "user@example.com"
	token, err := manager.GenerateToken(email)
	require.NoError(t, err, "Token generation failed")

	claims, err := manager.IsValidateToken(token, time.Now().Add(time.Minute/2))

	require.NoError(t, err, "Expected no error for valid token")
	assert.NotNil(t, claims, "Expected non-nil claims")
}

func TestIsValidateToken_Expired(t *testing.T) {
	manager := NewJWTManager("my-secret-key", time.Minute)
	email := "user@example.com"
	token, err := manager.GenerateToken(email)
	require.NoError(t, err, "Token generation failed")

	_, err = manager.IsValidateToken(token, time.Now().Add(time.Minute*2))

	require.Error(t, err, "Expected error for expired token")
	_, isExpiredErr := err.(*JWTExpiredTokenError)
	assert.True(t, isExpiredErr, "Expected JWTExpiredTokenError")
}

func TestIsValidateToken_InvalidSigningMethod(t *testing.T) {
	manager := NewJWTManager("my-secret-key", time.Minute)
	email := "user@example.com"
	token, err := manager.GenerateToken(email)
	require.NoError(t, err, "Token generation failed")

	invalidToken := token[:len(token)-1] + "X"
	_, err = manager.IsValidateToken(invalidToken, time.Now())

	require.Error(t, err, "Expected error for invalid token")
	_, isInvalidErr := err.(*JWTInvalidTokenError)
	assert.True(t, isInvalidErr, "Expected JWTInvalidTokenError")
}

func TestIsValidateToken_MalformedToken(t *testing.T) {
	manager := NewJWTManager("my-secret-key", time.Minute)
	invalidToken := "malformed.token.string"

	_, err := manager.IsValidateToken(invalidToken, time.Now())

	require.Error(t, err, "Expected error for malformed token")
	_, isInvalidErr := err.(*JWTInvalidTokenError)
	assert.True(t, isInvalidErr, "Expected JWTInvalidTokenError")
}
