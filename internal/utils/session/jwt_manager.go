package session

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JWTManager struct {
	secretAccessToken    string
	secretRefreshToken   string
	accessTokenDuration  time.Duration
	refreshTokenDuration time.Duration
}

type JWTFailedToGenerateError struct{}

func (e *JWTFailedToGenerateError) Error() string {
	return "failed to generate JWT token"
}

type JWTInvalidTokenError struct {
	message string
}

func (e *JWTInvalidTokenError) Error() string {
	return "invalid token: " + e.message
}

type JWTExpiredTokenError struct {
}

func (e *JWTExpiredTokenError) Error() string {
	return "token expired"
}

func NewJWTManager(secretAccessToken string, secretRefreshToken string, accessTokenDuration time.Duration, refreshTokenDuration time.Duration) *JWTManager {
	return &JWTManager{
		secretAccessToken:    secretAccessToken,
		secretRefreshToken:   secretRefreshToken,
		accessTokenDuration:  accessTokenDuration,
		refreshTokenDuration: refreshTokenDuration,
	}
}

func (m *JWTManager) GenerateAccessToken(id string, isAdmin bool) (string, error) {
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, &jwt.MapClaims{
		"userID":  id,
		"exp":     time.Now().Add(m.accessTokenDuration).Unix(),
		"iss":     time.Now().Unix(),
		"isAdmin": isAdmin,
	})

	tokenString, err := accessToken.SignedString([]byte(m.secretAccessToken))
	if err != nil {
		return "", &JWTFailedToGenerateError{}
	}

	return tokenString, nil
}

func (m *JWTManager) GenerateRefreshToken(id string, isAdmin bool) (string, error) {
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, &jwt.MapClaims{
		"userID":  id,
		"exp":     time.Now().Add(m.refreshTokenDuration).Unix(),
		"iss":     time.Now().Unix(),
		"isAdmin": isAdmin,
	})

	refreshTokenString, err := refreshToken.SignedString([]byte(m.secretRefreshToken))
	if err != nil {
		return "", &JWTFailedToGenerateError{}
	}

	return refreshTokenString, nil
}

func isValidToken(tokenString string, now time.Time, signature string, lastUpdate time.Time) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, &JWTInvalidTokenError{}
		}
		return []byte(signature), nil
	})
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, &JWTExpiredTokenError{}
		}
		return nil, &JWTInvalidTokenError{message: err.Error()}
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, &JWTInvalidTokenError{message: "invalid claims"}
	}

	expiresAtFloat, ok := claims["exp"].(float64)
	if !ok {
		return nil, &JWTInvalidTokenError{message: "expires_at claim missing or invalid"}
	}

	if int64(expiresAtFloat) < now.Unix() {
		return nil, &JWTExpiredTokenError{}
	}

	return claims, nil
}

func (m *JWTManager) IsValidateAccessToken(tokenString string, now time.Time) (jwt.MapClaims, error) {
	return isValidToken(tokenString, now, m.secretAccessToken, now)
}

func (m *JWTManager) IsValidateRefreshToken(tokenString string, now time.Time, lastUpdate time.Time) (jwt.MapClaims, error) {
	claims, err := isValidToken(tokenString, now, m.secretRefreshToken, lastUpdate)
	if err != nil {
		return nil, err
	}

	if lastUpdate.Unix() > int64(claims["iss"].(float64)) {
		return nil, &JWTExpiredTokenError{}
	}

	return claims, nil
}
