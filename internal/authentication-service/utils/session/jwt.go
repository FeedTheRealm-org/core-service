package session

import (
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

type JWTInvalidTokenError struct{}

func (e *JWTInvalidTokenError) Error() string {
	return "invalid token"
}

type JWTExpiredTokenError struct{}

func (e *JWTExpiredTokenError) Error() string {
	return "token expired"
}

func NewJWTManager(secretKey string, tokenDuration time.Duration) *JWTManager {
	return &JWTManager{
		secretKey:     secretKey,
		tokenDuration: tokenDuration,
	}
}

func (m *JWTManager) GenerateToken(email string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, &jwt.MapClaims{
		"email":      email,
		"expires_at": time.Now().Add(m.tokenDuration).Unix(),
		"issued_at":  time.Now().Unix(),
	})

	tokenString, err := token.SignedString([]byte(m.secretKey))
	if err != nil {
		return "", &JWTFailedToGenerateError{}
	}

	return tokenString, nil
}

func (m *JWTManager) IsValidateToken(tokenString string, now time.Time) error {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, &JWTInvalidTokenError{}
		}
		return []byte(m.secretKey), nil
	})

	if err != nil {
		return err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return &JWTInvalidTokenError{}
	}

	expiresAtFloat, ok := claims["expires_at"].(float64)
	if !ok {
		return &JWTInvalidTokenError{}
	}

	if int64(expiresAtFloat) < now.Unix() {
		return &JWTExpiredTokenError{}
	}

	return nil
}
