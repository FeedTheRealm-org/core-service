package acceptance_tests

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/cucumber/godog"
)

type response struct {
	Message string `json:"message"`
	Email   string `json:"email,omitempty"`
}

type responseError struct {
	Error string `json:"error"`
}

func aSignUpRequestIsMadeWithEmailAndPassword(ctx context.Context, email, password string) (context.Context, error) {
	payload := map[string]string{
		"email":    email,
		"password": password,
	}

	b, err := json.Marshal(payload)
	if err != nil {
		return ctx, err
	}

	req, err := http.NewRequest(http.MethodPost, "http://0.0.0.0:8000/auth/signup", bytes.NewReader(b))
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

func aSignUpRequestIsMadeWithAnEmptyEmailAndPassword(ctx context.Context, password string) (context.Context, error) {
	return aSignUpRequestIsMadeWithEmailAndPassword(ctx, "", password)
}

func aSignUpRequestIsMadeWithEmailAndAnEmptyPassword(ctx context.Context, email string) (context.Context, error) {
	return aSignUpRequestIsMadeWithEmailAndPassword(ctx, email, "")
}

func theResponseShouldIndicateSuccess(ctx context.Context) error {
	val := ctx.Value(testContextKey{})
	body, ok := val.([]byte)
	if !ok {
		return fmt.Errorf("no result found in context")
	}

	var res response
	if err := json.Unmarshal(body, &res); err != nil {
		return fmt.Errorf("failed to parse response: %v", err)
	}

	if res.Message != "Account created successfully" {
		return fmt.Errorf("unexpected result: %s", res.Message)
	}

	return nil
}

func theResponseShouldIncludeAnErrorMessage(ctx context.Context, expectedMsg string) error {
	val := ctx.Value(testContextKey{})
	body, ok := val.([]byte)
	if !ok {
		return fmt.Errorf("no result found in context")
	}

	var res responseError
	if err := json.Unmarshal(body, &res); err != nil {
		return fmt.Errorf("failed to parse response: %v", err)
	}

	if res.Error != expectedMsg {
		return fmt.Errorf("unexpected result: %s", res.Error)
	}

	return nil
}

func anAccountAlreadyExistsWithTheEmail(ctx context.Context, email string) (context.Context, error) {
	return aSignUpRequestIsMadeWithEmailAndPassword(ctx, email, "somepassword")
}

func InitializeScenarioForAccount(sc *godog.ScenarioContext) {
	sc.Step(`^a sign-up request is made with email "([^"]*)" and password "([^"]*)"$`, aSignUpRequestIsMadeWithEmailAndPassword)
	sc.Step(`^a sign-up request is made with an empty email and password "([^"]*)"$`, aSignUpRequestIsMadeWithAnEmptyEmailAndPassword)
	sc.Step(`^a sign-up request is made with email "([^"]*)" and an empty password$`, aSignUpRequestIsMadeWithEmailAndAnEmptyPassword)
	sc.Step(`^the response should indicate success$`, theResponseShouldIndicateSuccess)
	sc.Step(`^the response should include an error message "([^"]*)"$`, theResponseShouldIncludeAnErrorMessage)
	sc.Step(`^an account already exists with the email "([^"]*)"$`, anAccountAlreadyExistsWithTheEmail)
}
