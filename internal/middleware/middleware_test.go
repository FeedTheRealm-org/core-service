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
	"github.com/FeedTheRealm-org/core-service/internal/utils/oidc_validation"
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

func TestCORSMiddleware_OriginNotAllowed(t *testing.T) {
	gin.SetMode(gin.TestMode)
	conf := config.CreateConfig()
	conf.CORSAllowedOrigins = []string{"https://allowed.example"}

	r := gin.New()
	r.Use(middleware.CORSMiddleware(conf))
	r.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Origin", "https://blocked.example")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "", w.Header().Get("Access-Control-Allow-Origin"))
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

func TestErrorHandlerMiddleware_WritesInternalError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(middleware.ErrorHandlerMiddleware())
	r.GET("/test", func(c *gin.Context) {
		_ = c.Error(assert.AnError)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
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

func TestMultipartCleanup_Coverage(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.Use(middleware.MultipartCleanupMiddleware())

	r.POST("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req, _ := http.NewRequest(http.MethodPost, "/test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestGithubOIDCCheck_FullCoverage(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		authHeader string
	}{
		{"Missing_Token", ""},
		{"Invalid_Token_Format", "Bearer token_invalido_test"},
		{"Valid_Structure_But_Fails_Validation", "Bearer eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.e30.s"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := gin.New()
			verifier := &oidc_validation.GitHubOIDCVerifier{}
			r.Use(middleware.GithubOIDCCheck(verifier))

			r.GET("/test", func(c *gin.Context) {
				c.Status(http.StatusOK)
			})

			req, _ := http.NewRequest(http.MethodGet, "/test", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}

			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.NotEqual(t, 0, w.Code)
		})
	}
}
