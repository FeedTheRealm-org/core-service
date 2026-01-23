package middleware_test

import (
	"net/http/httptest"
	"testing"
	"time"

	"github.com/FeedTheRealm-org/core-service/internal/middleware"
	"github.com/FeedTheRealm-org/core-service/internal/utils/logger"
	"github.com/FeedTheRealm-org/core-service/internal/utils/session"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
)

// TestJWTAuth_NoAuthHeader tests the JWT authentication middleware
// when no Authorization header is provided.
func TestJWTAuth_NoAuthHeader(t *testing.T) {
	r := setupRouterJWT("testsecret", "")

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	assert.Equal(t, "", w.Body.String())
}

// TestJWTAuth_ValidToken tests the JWT authentication middleware
// when a valid Authorization header is provided.
func TestJWTAuth_ValidToken(t *testing.T) {
	secret := "testsecret"
	token := createTestToken(secret, "12345", time.Now().Add(time.Hour))
	r := setupRouterJWT(secret, "")

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	assert.Equal(t, "12345", w.Body.String())
}

// TestJWTAuth_ExpiredToken tests the JWT authentication middleware
// when an expired token is provided in the Authorization header.
func TestJWTAuth_ExpiredToken(t *testing.T) {
	secret := "testsecret"
	token := createTestToken(secret, "expired", time.Now().Add(-time.Hour))
	r := setupRouterJWT(secret, "")

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, 401, w.Code)
	assert.Equal(t, "expired", w.Body.String())
}

/* UTILS */

// createTestToken creates a JWT token with the given secret, userID, and expiration time.
func createTestToken(secret, userID string, expiration time.Time) string {
	claims := jwt.MapClaims{
		"userID": userID,
		"exp":    expiration.Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	ss, _ := token.SignedString([]byte(secret))
	return ss
}

// setupRouterJWT initializes a Gin router with the JWT authentication middleware.
func setupRouterJWT(secret string, fixedToken string) *gin.Engine {
	gin.SetMode(gin.TestMode)
	jwtManager := session.NewJWTManager(secret, time.Hour)
	logger.InitLogger(false)

	r := gin.New()
	r.Use(middleware.JWTAuthMiddleware(jwtManager, fixedToken)) // Include Auth middleware
	r.GET("/test", func(c *gin.Context) {
		userID := c.GetString("userID")
		invalidJWT := c.GetBool("invalidJWT")
		expiredJWT := c.GetBool("expiredJWT")
		if expiredJWT {
			c.String(401, "expired")
		} else if invalidJWT {
			c.String(401, "invalid")
		} else {
			c.String(200, userID)
		}
	})

	return r
}

// TestJWTAuth_FixedToken verifies that when a fixed server token is configured
// the middleware treats it as a valid session without attempting JWT parsing.
func TestJWTAuth_FixedToken(t *testing.T) {
	secret := "testsecret"
	fixedToken := "fixed-server-token"
	r := setupRouterJWT(secret, fixedToken)

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+fixedToken)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	// No userID is set for the fixed token, but the session is
	// considered valid, so the handler returns empty body with 200.
	assert.Equal(t, "", w.Body.String())
}
