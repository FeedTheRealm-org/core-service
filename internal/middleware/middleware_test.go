package middleware_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/FeedTheRealm-org/core-service/config"
	"github.com/FeedTheRealm-org/core-service/internal/dtos"
	"github.com/FeedTheRealm-org/core-service/internal/errors"
	"github.com/FeedTheRealm-org/core-service/internal/middleware"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestCORSMiddleware_AllowsSpecificOrigin(t *testing.T) {
	gin.SetMode(gin.TestMode)
	conf := config.CreateConfig()
	conf.CORSAllowedOrigins = []string{"https://example.com"}

	r := gin.New()
	r.Use(middleware.CORSMiddleware(conf))
	r.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Origin", "https://example.com")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "https://example.com", w.Header().Get("Access-Control-Allow-Origin"))
}

func TestCORSMiddleware_OptionsShortCircuit(t *testing.T) {
	gin.SetMode(gin.TestMode)
	conf := config.CreateConfig()
	conf.CORSAllowedOrigins = []string{"*"}

	r := gin.New()
	r.Use(middleware.CORSMiddleware(conf))
	r.Any("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	req := httptest.NewRequest(http.MethodOptions, "/test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
	assert.Equal(t, "", w.Body.String())
	assert.Equal(t, "*", w.Header().Get("Access-Control-Allow-Origin"))
}

func TestErrorHandlerMiddleware_WritesHttpError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(middleware.ErrorHandlerMiddleware())
	r.GET("/test", func(c *gin.Context) {
		_ = c.Error(errors.NewBadRequestError("bad"))
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var payload dtos.ErrorResponse
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &payload))
	assert.Equal(t, http.StatusBadRequest, payload.Status)
	assert.Equal(t, "bad", payload.Detail)
}

func TestAdminCheckMiddleware_AllowsAdmin(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("includedJWT", true)
		c.Set("isAdmin", true)
	})
	r.Use(middleware.AdminCheckMiddleware())
	r.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "ok", w.Body.String())
}

func TestAdminCheckMiddleware_BlocksNonAdmin(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(middleware.ErrorHandlerMiddleware())
	r.Use(func(c *gin.Context) {
		c.Set("includedJWT", true)
		c.Set("isAdmin", false)
	})
	r.Use(middleware.AdminCheckMiddleware())
	r.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestServerCheckMiddleware_BlocksInvalidSession(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(middleware.ErrorHandlerMiddleware())
	r.Use(func(c *gin.Context) {
		c.Set("includedJWT", true)
		c.Set("invalidJWT", true)
		c.Set("isServer", false)
	})
	r.Use(middleware.ServerCheckMiddleware())
	r.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestServerCheckMiddleware_AllowsServer(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("includedJWT", true)
		c.Set("isServer", true)
	})
	r.Use(middleware.ServerCheckMiddleware())
	r.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "ok", w.Body.String())
}
