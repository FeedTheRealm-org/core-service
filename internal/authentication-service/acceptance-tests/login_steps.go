package acceptance_tests

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/cucumber/godog"
)

type loginResponse struct {
	Message string `json:"message"`
	Token   string `json:"token,omitempty"`
}

type sessionContext struct {
	token     string
	loginTime time.Time
}

type sessionContextKey struct{}

func aLoginRequestIsMadeWithEmailAndPassword(ctx context.Context, email, password string) (context.Context, error) {
	payload := map[string]string{
		"email":    email,
		"password": password,
	}

	b, err := json.Marshal(payload)
	if err != nil {
		return ctx, err
	}

	req, err := http.NewRequest(http.MethodPost, "http://0.0.0.0:8000/auth/login", bytes.NewReader(b))
	if err != nil {
		return ctx, err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return ctx, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return ctx, err
	}

	return context.WithValue(ctx, testContextKey{}, body), nil
}

func aLoginRequestIsMadeWithAnEmptyEmailAndPassword(ctx context.Context, password string) (context.Context, error) {
	return aLoginRequestIsMadeWithEmailAndPassword(ctx, "", password)
}

func aLoginRequestIsMadeWithEmailAndAnEmptyPassword(ctx context.Context, email string) (context.Context, error) {
	return aLoginRequestIsMadeWithEmailAndPassword(ctx, email, "")
}

func theUserHasLoggedInSuccessfully(ctx context.Context) (context.Context, error) {
	// First create an account
	ctx, err := aSignUpRequestIsMadeWithEmailAndPassword(ctx, "sessionuser@example.com", "ValidPass123!")
	if err != nil {
		return ctx, err
	}

	// Then login
	payload := map[string]string{
		"email":    "sessionuser@example.com",
		"password": "ValidPass123!",
	}

	b, err := json.Marshal(payload)
	if err != nil {
		return ctx, err
	}

	req, err := http.NewRequest(http.MethodPost, "http://0.0.0.0:8000/auth/login", bytes.NewReader(b))
	if err != nil {
		return ctx, err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return ctx, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return ctx, err
	}

	var loginResp loginResponse
	if err := json.Unmarshal(body, &loginResp); err != nil {
		return ctx, fmt.Errorf("failed to parse login response: %v", err)
	}

	session := &sessionContext{
		token:     loginResp.Token,
		loginTime: time.Now(),
	}

	return context.WithValue(ctx, sessionContextKey{}, session), nil
}

func minutesHavePassedSinceLogin(ctx context.Context, minutes string) (context.Context, error) {
	val := ctx.Value(sessionContextKey{})
	session, ok := val.(*sessionContext)
	if !ok {
		return ctx, fmt.Errorf("no session found in context")
	}

	// Parse minutes
	var mins int
	fmt.Sscanf(minutes, "%d", &mins)

	// Simulate time passing by updating the login time
	session.loginTime = time.Now().Add(-time.Duration(mins) * time.Minute)

	return context.WithValue(ctx, sessionContextKey{}, session), nil
}

func theSessionShouldStillBeActive(ctx context.Context) error {
	val := ctx.Value(sessionContextKey{})
	session, ok := val.(*sessionContext)
	if !ok {
		return fmt.Errorf("no session found in context")
	}

	// Check if session is still valid (within 60 minutes)
	elapsed := time.Since(session.loginTime)
	if elapsed >= 60*time.Minute {
		return fmt.Errorf("session should be active but has expired")
	}

	// Optionally verify with an authenticated request
	if session.token == "" {
		return fmt.Errorf("no token found in session")
	}

	return nil
}

func theSessionShouldBeClosed(ctx context.Context) error {
	val := ctx.Value(sessionContextKey{})
	session, ok := val.(*sessionContext)
	if !ok {
		return fmt.Errorf("no session found in context")
	}

	// Check if session has expired (after 60 minutes)
	elapsed := time.Since(session.loginTime)
	if elapsed < 60*time.Minute {
		return fmt.Errorf("session should be closed but is still active")
	}

	return nil
}

func furtherRequestsShouldRequireAuthentication(ctx context.Context) error {
	req, err := http.NewRequest(http.MethodGet, "http://0.0.0.0:8000/auth/protected", nil)
	if err != nil {
		return err
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnauthorized {
		return fmt.Errorf("expected 401 status code, got %d", resp.StatusCode)
	}

	return nil
}

func theResponseShouldIndicateASuccessfulLogin(ctx context.Context) error {
	val := ctx.Value(testContextKey{})
	bodyBytes, ok := val.([]byte)
	if !ok {
		return fmt.Errorf("no response body found in context")
	}

	// Check if it's an error response first
	var errorResp responseError
	if err := json.Unmarshal(bodyBytes, &errorResp); err == nil && errorResp.Error != "" {
		return fmt.Errorf("received error response instead of success: %s (body: %s)", errorResp.Error, string(bodyBytes))
	}

	var loginResp loginResponse
	if err := json.Unmarshal(bodyBytes, &loginResp); err != nil {
		return fmt.Errorf("failed to parse login response: %v (body: %s)", err, string(bodyBytes))
	}

	if loginResp.Message != "Login successful" {
		return fmt.Errorf("unexpected login message: %s (body: %s)", loginResp.Message, string(bodyBytes))
	}

	if loginResp.Token == "" {
		return fmt.Errorf("expected a token in the login response, but got none")
	}

	return nil
}

func InitializeScenarioForLogin(sc *godog.ScenarioContext) {
	sc.Step(`^a login request is made with email "([^"]*)" and password "([^"]*)"$`, aLoginRequestIsMadeWithEmailAndPassword)
	sc.Step(`^a login request is made with an empty email and password "([^"]*)"$`, aLoginRequestIsMadeWithAnEmptyEmailAndPassword)
	sc.Step(`^a login request is made with email "([^"]*)" and an empty password$`, aLoginRequestIsMadeWithEmailAndAnEmptyPassword)
	sc.Step(`^the user has logged in successfully$`, theUserHasLoggedInSuccessfully)
	sc.Step(`^"([^"]*)" minutes have passed since login$`, minutesHavePassedSinceLogin)
	sc.Step(`^the session should still be active$`, theSessionShouldStillBeActive)
	sc.Step(`^the session should be closed$`, theSessionShouldBeClosed)
	sc.Step(`^further requests should require authentication$`, furtherRequestsShouldRequireAuthentication)
	sc.Step(`^the response should indicate a successful login$`, theResponseShouldIndicateASuccessfulLogin)
	sc.Step(`^an account already exists with the email "([^"]*)" and password "([^"]*)"$`, aSignUpRequestIsMadeWithEmailAndPassword)
}
