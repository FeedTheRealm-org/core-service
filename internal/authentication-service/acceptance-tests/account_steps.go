package acceptance_tests

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/cucumber/godog"
)

type response struct {
	Message string `json:"message"`
}

func aSignUpRequestIsMadeWithEmailAndPassword(ctx context.Context, email, password string) (context.Context, error) {
	resp, err := http.Get("http://0.0.0.0:8000/auth/example-query")
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

	if res.Message != "User created successfully" {
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

	var res response
	if err := json.Unmarshal(body, &res); err != nil {
		return fmt.Errorf("failed to parse response: %v", err)
	}

	if res.Message != expectedMsg {
		return fmt.Errorf("unexpected result: %s", res.Message)
	}

	return nil
}

func anAccountAlreadyExistsWithTheEmail(ctx context.Context, email string) error {
	val := ctx.Value(testContextKey{})
	body, ok := val.([]byte)
	if !ok {
		return fmt.Errorf("no result found in context")
	}

	var res response
	if err := json.Unmarshal(body, &res); err != nil {
		return fmt.Errorf("failed to parse response: %v", err)
	}

	expectedMsg := fmt.Sprintf("Account with email %s already exists", email)

	if res.Message != expectedMsg {
		return fmt.Errorf("unexpected result: %s", res.Message)
	}

	return nil
}

func InitializeScenarioForAccount(sc *godog.ScenarioContext) {
	sc.Step(`^a sign-up request is made with email "([^"]*)" and password "([^"]*)"$`, aSignUpRequestIsMadeWithEmailAndPassword)
	sc.Step(`^a sign-up request is made with an empty email and password "([^"]*)"$`, aSignUpRequestIsMadeWithAnEmptyEmailAndPassword)
	sc.Step(`^a sign-up request is made with email "([^"]*)" and an empty password$`, aSignUpRequestIsMadeWithEmailAndAnEmptyPassword)
	sc.Step(`^the response should indicate success$`, theResponseShouldIndicateSuccess)
	sc.Step(`^the response should include an error message "([^"]*)"$`, theResponseShouldIncludeAnErrorMessage)
	sc.Step(`^an account already exists with the email "([^"]*)"$`, anAccountAlreadyExistsWithTheEmail)
}
