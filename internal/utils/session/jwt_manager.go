package session

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JWTManager struct {
	secretKey     string
	tokenDuration time.Duration
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

func NewJWTManager(secretKey string, tokenDuration time.Duration) *JWTManager {
	return &JWTManager{
		secretKey:     secretKey,
		tokenDuration: tokenDuration,
	}
}

func (m *JWTManager) GenerateToken(id string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, &jwt.MapClaims{
		"userID": id,
		"exp":    time.Now().Add(m.tokenDuration).Unix(),
		"iss":    time.Now().Unix(),
	})

	tokenString, err := token.SignedString([]byte(m.secretKey))
	if err != nil {
		return "", &JWTFailedToGenerateError{}
	}

	return tokenString, nil
}

func (m *JWTManager) IsValidateToken(tokenString string, now time.Time) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, &JWTInvalidTokenError{}
		}
		return []byte(m.secretKey), nil
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
