package common_handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"

	"github.com/FeedTheRealm-org/core-service/internal/errors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestContext(method string, path string) (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = httptest.NewRequest(method, path, nil)
	return ctx, w
}

func TestIsSessionValid_Errors(t *testing.T) {
	ctx, _ := setupTestContext(http.MethodGet, "/")
	_, err := GetUserIDFromSession(ctx)
	assert.Error(t, err)
	_, notIn := err.(*errors.NotInSessionError)
	assert.True(t, notIn)

	ctx, _ = setupTestContext(http.MethodGet, "/")
	ctx.Set("includedJWT", true)
	ctx.Set("invalidJWT", true)
	_, err = GetUserIDFromSession(ctx)
	assert.Error(t, err)
	_, invalid := err.(*errors.InvalidSessionError)
	assert.True(t, invalid)

	ctx, _ = setupTestContext(http.MethodGet, "/")
	ctx.Set("includedJWT", true)
	ctx.Set("expiredJWT", true)
	_, err = GetUserIDFromSession(ctx)
	assert.Error(t, err)
	_, expired := err.(*errors.ExpiredSessionError)
	assert.True(t, expired)
}

func TestIsAdminSession(t *testing.T) {
	ctx, _ := setupTestContext(http.MethodGet, "/")
	ctx.Set("includedJWT", true)
	ctx.Set("isAdmin", false)
	err := IsAdminSession(ctx)
	assert.Error(t, err)

	ctx, _ = setupTestContext(http.MethodGet, "/")
	ctx.Set("includedJWT", true)
	ctx.Set("isAdmin", true)
	err = IsAdminSession(ctx)
	assert.NoError(t, err)
}

func TestIsServerSession_AllowsServerFlag(t *testing.T) {
	ctx, _ := setupTestContext(http.MethodGet, "/")
	ctx.Set("includedJWT", true)
	ctx.Set("invalidJWT", true)
	ctx.Set("isServer", true)
	err := IsServerSession(ctx)
	assert.NoError(t, err)
}

func TestGetEmailFromSession(t *testing.T) {
	ctx, _ := setupTestContext(http.MethodGet, "/")
	ctx.Set("includedJWT", true)
	ctx.Set("email", "user@example.com")

	email, err := GetEmailFromSession(ctx)
	assert.NoError(t, err)
	assert.Equal(t, "user@example.com", email)
}

func TestGetUserIDFromSession_InvalidUUID(t *testing.T) {
	ctx, _ := setupTestContext(http.MethodGet, "/")
	ctx.Set("includedJWT", true)
	ctx.Set("userID", "not-a-uuid")

	_, err := GetUserIDFromSession(ctx)
	assert.Error(t, err)
	_, invalid := err.(*errors.InvalidSessionError)
	assert.True(t, invalid)
}

func TestIsGithubOIDCTokenValid(t *testing.T) {
	ctx, _ := setupTestContext(http.MethodGet, "/")
	ctx.Set("invalidGithubOIDC", true)
	err := IsGithubOIDCTokenValid(ctx)
	assert.Error(t, err)
	_, invalid := err.(*errors.InvalidGithubOIDCTokenError)
	assert.True(t, invalid)

	ctx, _ = setupTestContext(http.MethodGet, "/")
	ctx.Set("invalidGithubOIDC", false)
	err = IsGithubOIDCTokenValid(ctx)
	assert.NoError(t, err)
}

func TestHandleSuccessResponse(t *testing.T) {
	ctx, w := setupTestContext(http.MethodGet, "/")

	HandleSuccessResponse(ctx, http.StatusOK, map[string]string{"status": "ok"})
	assert.Equal(t, http.StatusOK, w.Code)

	var payload map[string]map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &payload)
	assert.NoError(t, err)
	assert.Equal(t, "ok", payload["data"]["status"])
}

func TestHandleBodilessResponse(t *testing.T) {
	ctx, w := setupTestContext(http.MethodDelete, "/")

	HandleBodilessResponse(ctx, http.StatusNoContent)
	assert.Equal(t, http.StatusNoContent, w.Code)
	assert.Equal(t, "", w.Body.String())
}

func TestHealthController(t *testing.T) {
	ctx, w := setupTestContext(http.MethodGet, "/health")
	HealthController(ctx)

	assert.Equal(t, http.StatusOK, w.Code)
	var payload map[string]map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &payload)
	assert.NoError(t, err)
	assert.Equal(t, "ok", payload["data"]["status"])
}

func TestCheckWorldOwnership(t *testing.T) {
	userID := uuid.New()
	worldID := uuid.New()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/world/"+worldID.String() {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"user_id":"` + userID.String() + `"}}`))
	}))
	defer server.Close()

	parsed, err := url.Parse(server.URL)
	require.NoError(t, err)
	port, err := strconv.Atoi(parsed.Port())
	require.NoError(t, err)

	ctx, _ := setupTestContext(http.MethodGet, "/")
	ctx.Request.Header.Set("Authorization", "Bearer token")

	err = CheckWorldOwnership(ctx, port, worldID, userID)
	assert.NoError(t, err)
}

func TestCheckWorldOwnership_StatusErrors(t *testing.T) {
	userID := uuid.New()
	worldID := uuid.New()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case bytes.Contains([]byte(r.URL.Path), []byte("unauthorized")):
			w.WriteHeader(http.StatusUnauthorized)
		case bytes.Contains([]byte(r.URL.Path), []byte("bad")):
			w.WriteHeader(http.StatusBadRequest)
		case bytes.Contains([]byte(r.URL.Path), []byte("boom")):
			w.WriteHeader(http.StatusInternalServerError)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	parsed, err := url.Parse(server.URL)
	require.NoError(t, err)
	port, err := strconv.Atoi(parsed.Port())
	require.NoError(t, err)

	ctx, _ := setupTestContext(http.MethodGet, "/")
	ctx.Request.Header.Set("Authorization", "Bearer token")

	cases := []struct {
		path string
		want int
	}{
		{path: "/world/unauthorized", want: http.StatusUnauthorized},
		{path: "/world/bad", want: http.StatusBadRequest},
		{path: "/world/boom", want: http.StatusInternalServerError},
		{path: "/world/notfound", want: http.StatusBadRequest},
	}

	for _, tc := range cases {
		ctx.Request.URL.Path = tc.path
		err = CheckWorldOwnership(ctx, port, worldID, userID)
		assert.Error(t, err)
		httpErr, ok := err.(*errors.HttpError)
		assert.True(t, ok)
		assert.Equal(t, tc.want, httpErr.Status)
	}
}

func TestCheckWorldOwnership_UserMismatch(t *testing.T) {
	userID := uuid.New()
	worldID := uuid.New()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"user_id":"` + uuid.NewString() + `"}}`))
	}))
	defer server.Close()

	parsed, err := url.Parse(server.URL)
	require.NoError(t, err)
	port, err := strconv.Atoi(parsed.Port())
	require.NoError(t, err)

	ctx, _ := setupTestContext(http.MethodGet, "/")
	ctx.Request.Header.Set("Authorization", "Bearer token")

	err = CheckWorldOwnership(ctx, port, worldID, userID)
	assert.Error(t, err)
	httpErr, ok := err.(*errors.HttpError)
	assert.True(t, ok)
	assert.Equal(t, http.StatusUnauthorized, httpErr.Status)
}

func TestCheckWorldOwnership_BadJSON(t *testing.T) {
	userID := uuid.New()
	worldID := uuid.New()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("not-json"))
	}))
	defer server.Close()

	parsed, err := url.Parse(server.URL)
	require.NoError(t, err)
	port, err := strconv.Atoi(parsed.Port())
	require.NoError(t, err)

	ctx, _ := setupTestContext(http.MethodGet, "/")
	ctx.Request.Header.Set("Authorization", "Bearer token")

	err = CheckWorldOwnership(ctx, port, worldID, userID)
	assert.Error(t, err)
	httpErr, ok := err.(*errors.HttpError)
	assert.True(t, ok)
	assert.Equal(t, http.StatusInternalServerError, httpErr.Status)
}
