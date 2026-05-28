package session

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIsValidateToken_Valid(t *testing.T) {
	manager := NewJWTManager("my-secret-access-key", "my-secret-refresh-key", time.Minute, time.Hour)
	email := "user@example.com"
	token, err := manager.GenerateAccessToken(email, email, false)
	require.NoError(t, err, "Token generation failed")

	claims, err := manager.IsValidateAccessToken(token, time.Now().Add(time.Minute/2))

	require.NoError(t, err, "Expected no error for valid token")
	assert.NotNil(t, claims, "Expected non-nil claims")
}

func TestIsValidateToken_Expired(t *testing.T) {
	manager := NewJWTManager("my-secret-access-key", "my-secret-refresh-key", time.Minute, time.Hour)
	email := "user@example.com"
	token, err := manager.GenerateAccessToken(email, email, false)
	require.NoError(t, err, "Token generation failed")

	_, err = manager.IsValidateAccessToken(token, time.Now().Add(time.Minute*2))

	require.Error(t, err, "Expected error for expired token")
	_, isExpiredErr := err.(*JWTExpiredTokenError)
	assert.True(t, isExpiredErr, "Expected JWTExpiredTokenError")
}

func TestIsValidateToken_InvalidSigningMethod(t *testing.T) {
	manager := NewJWTManager("my-secret-access-key", "my-secret-refresh-key", time.Minute, time.Hour)
	email := "user@example.com"
	token, err := manager.GenerateAccessToken(email, email, false)
	require.NoError(t, err, "Token generation failed")

	invalidToken := token[:len(token)-1] + "X"
	_, err = manager.IsValidateAccessToken(invalidToken, time.Now())

	require.Error(t, err, "Expected error for invalid token")
	_, isInvalidErr := err.(*JWTInvalidTokenError)
	assert.True(t, isInvalidErr, "Expected JWTInvalidTokenError")
}

func TestIsValidateToken_MalformedToken(t *testing.T) {
	manager := NewJWTManager("my-secret-access-key", "my-secret-refresh-key", time.Minute, time.Hour)
	invalidToken := "malformed.token.string"

	_, err := manager.IsValidateAccessToken(invalidToken, time.Now())

	require.Error(t, err, "Expected error for malformed token")
	_, isInvalidErr := err.(*JWTInvalidTokenError)
	assert.True(t, isInvalidErr, "Expected JWTInvalidTokenError")
}

func TestIsValidateRefreshToken_Valid(t *testing.T) {
	manager := NewJWTManager("my-secret-access-key", "my-secret-refresh-key", time.Minute, time.Hour)
	userID := "user-id"
	token, err := manager.GenerateRefreshToken(userID, "user@example.com", true)
	require.NoError(t, err)

	claims, err := manager.IsValidateRefreshToken(token, time.Now().Add(time.Minute), time.Now())
	require.NoError(t, err)
	assert.Equal(t, userID, claims["userID"])
}

func TestIsValidateRefreshToken_IssuedBeforeLastUpdate(t *testing.T) {
	manager := NewJWTManager("my-secret-access-key", "my-secret-refresh-key", time.Minute, time.Hour)
	token, err := manager.GenerateRefreshToken("user-id", "user@example.com", false)
	require.NoError(t, err)

	lastUpdate := time.Now().Add(time.Minute)
	_, err = manager.IsValidateRefreshToken(token, time.Now().Add(time.Minute), lastUpdate)
	require.Error(t, err)
	_, isExpired := err.(*JWTExpiredTokenError)
	assert.True(t, isExpired)
}

func TestIsValidateAccessToken_InvalidSignature(t *testing.T) {
	manager := NewJWTManager("my-secret-access-key", "my-secret-refresh-key", time.Minute, time.Hour)
	refreshToken, err := manager.GenerateRefreshToken("user-id", "user@example.com", false)
	require.NoError(t, err)

	_, err = manager.IsValidateAccessToken(refreshToken, time.Now())
	require.Error(t, err)
	_, isInvalid := err.(*JWTInvalidTokenError)
	assert.True(t, isInvalid)
}

func TestIsValidateAccessToken_InvalidExpClaim(t *testing.T) {
	secret := "my-secret-access-key"
	manager := NewJWTManager(secret, "my-secret-refresh-key", time.Minute, time.Hour)

	badToken := jwt.NewWithClaims(jwt.SigningMethodHS256, &jwt.MapClaims{
		"userID": "user-id",
		"email":  "user@example.com",
		"exp":    "not-a-number",
		"iss":    time.Now().Unix(),
	})

	tokenString, err := badToken.SignedString([]byte(secret))
	require.NoError(t, err)

	_, err = manager.IsValidateAccessToken(tokenString, time.Now())
	require.Error(t, err)
	_, isInvalid := err.(*JWTInvalidTokenError)
	assert.True(t, isInvalid)
}

func TestIsValidateAccessToken_ExpiredFromParser(t *testing.T) {
	secret := "my-secret-access-key"
	manager := NewJWTManager(secret, "my-secret-refresh-key", time.Minute, time.Hour)

	expiredToken := jwt.NewWithClaims(jwt.SigningMethodHS256, &jwt.MapClaims{
		"userID": "user-id",
		"email":  "user@example.com",
		"exp":    time.Now().Add(-time.Minute).Unix(),
		"iss":    time.Now().Unix(),
	})

	tokenString, err := expiredToken.SignedString([]byte(secret))
	require.NoError(t, err)

	_, err = manager.IsValidateAccessToken(tokenString, time.Now())
	require.Error(t, err)
	_, isExpired := err.(*JWTExpiredTokenError)
	assert.True(t, isExpired)
}

func TestIsValidateAccessToken_InvalidSigningMethod(t *testing.T) {
	manager := NewJWTManager("my-secret-access-key", "my-secret-refresh-key", time.Minute, time.Hour)

	noneToken := jwt.NewWithClaims(jwt.SigningMethodNone, &jwt.MapClaims{
		"userID": "user-id",
		"email":  "user@example.com",
		"exp":    time.Now().Add(time.Minute).Unix(),
		"iss":    time.Now().Unix(),
	})

	tokenString, err := noneToken.SignedString(jwt.UnsafeAllowNoneSignatureType)
	require.NoError(t, err)

	_, err = manager.IsValidateAccessToken(tokenString, time.Now())
	require.Error(t, err)
	_, isInvalid := err.(*JWTInvalidTokenError)
	assert.True(t, isInvalid)
}
