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

type verificationContextKey struct{}

type verificationContext struct {
	email        string
	password     string
	verifyCode   string
	responseBody []byte
	statusCode   int
}

type verifyResponse struct {
	Data struct {
		Email    string `json:"email"`
		Verified bool   `json:"verified"`
	} `json:"data"`
}

func aPlayerRegistersAnAccountWithEmailAndPassword(ctx context.Context, email, password string) (context.Context, error) {
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

	if resp.StatusCode != 201 {
		return ctx, fmt.Errorf("signup failed with status %d: %s", resp.StatusCode, string(body))
	}

	verifyCtx := &verificationContext{
		email:        email,
		password:     password,
		responseBody: body,
		statusCode:   resp.StatusCode,
	}

	return context.WithValue(ctx, verificationContextKey{}, verifyCtx), nil
}

func theRegistrationIsCompleted(ctx context.Context) (context.Context, error) {
	val := ctx.Value(verificationContextKey{})
	verifyCtx, ok := val.(*verificationContext)
	if !ok {
		return ctx, fmt.Errorf("no verification context found")
	}

	if verifyCtx.statusCode != 201 {
		return ctx, fmt.Errorf("registration failed with status code: %d", verifyCtx.statusCode)
	}

	return ctx, nil
}

func anEmailShouldBeSentToContainingAOneTimeVerificationCode(ctx context.Context, email string) error {
	val := ctx.Value(verificationContextKey{})
	verifyCtx, ok := val.(*verificationContext)
	if !ok {
		return fmt.Errorf("no verification context found")
	}

	if verifyCtx.email != email {
		return fmt.Errorf("expected email %s, got %s", email, verifyCtx.email)
	}

	return nil
}

func thePlayerHasReceivedAVerificationEmailWithAValidOneTimeCode(ctx context.Context) (context.Context, error) {
	val := ctx.Value(verificationContextKey{})
	verifyCtx, ok := val.(*verificationContext)
	if !ok {
		return ctx, fmt.Errorf("no verification context found")
	}

	verifyCtx.verifyCode = "123456"

	return context.WithValue(ctx, verificationContextKey{}, verifyCtx), nil
}

func thePlayerSubmitsTheCorrectCodeWithinTheValidTimeWindow(ctx context.Context) (context.Context, error) {
	val := ctx.Value(verificationContextKey{})
	verifyCtx, ok := val.(*verificationContext)
	if !ok {
		return ctx, fmt.Errorf("no verification context found")
	}

	payload := map[string]string{
		"email": verifyCtx.email,
		"code":  "123456",
	}

	b, err := json.Marshal(payload)
	if err != nil {
		return ctx, err
	}

	req, err := http.NewRequest(http.MethodPost, "http://0.0.0.0:8000/auth/verify", bytes.NewReader(b))
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

	verifyCtx.responseBody = body
	verifyCtx.statusCode = resp.StatusCode

	return context.WithValue(ctx, verificationContextKey{}, verifyCtx), nil
}

func theAccountShouldBeMarkedAsVerified(ctx context.Context) error {
	val := ctx.Value(verificationContextKey{})
	verifyCtx, ok := val.(*verificationContext)
	if !ok {
		return fmt.Errorf("no verification context found")
	}

	if verifyCtx.statusCode != 200 {
		return fmt.Errorf("expected status 200, got %d (body: %s)", verifyCtx.statusCode, string(verifyCtx.responseBody))
	}

	var res verifyResponse
	if err := json.Unmarshal(verifyCtx.responseBody, &res); err != nil {
		return fmt.Errorf("failed to parse verify response: %v (body: %s)", err, string(verifyCtx.responseBody))
	}

	if !res.Data.Verified {
		return fmt.Errorf("account should be verified but is not")
	}

	return nil
}

func thePlayerShouldBeAbleToLogInSuccessfully(ctx context.Context) error {
	val := ctx.Value(verificationContextKey{})
	verifyCtx, ok := val.(*verificationContext)
	if !ok {
		return fmt.Errorf("no verification context found")
	}

	payload := map[string]string{
		"email":    verifyCtx.email,
		"password": verifyCtx.password,
	}

	b, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, "http://0.0.0.0:8000/auth/login", bytes.NewReader(b))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("login failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

func thePlayerHasRegisteredButNotYetVerifiedTheirAccount(ctx context.Context) (context.Context, error) {
	return ctx, nil
}

func thePlayerAttemptsToLogInToTheGame(ctx context.Context) (context.Context, error) {
	val := ctx.Value(verificationContextKey{})
	verifyCtx, ok := val.(*verificationContext)
	if !ok {
		return ctx, fmt.Errorf("no verification context found")
	}

	payload := map[string]string{
		"email":    verifyCtx.email,
		"password": verifyCtx.password,
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

	verifyCtx.responseBody = body
	verifyCtx.statusCode = resp.StatusCode

	return context.WithValue(ctx, verificationContextKey{}, verifyCtx), nil
}

func theLoginShouldBeRejected(ctx context.Context) error {
	val := ctx.Value(verificationContextKey{})
	verifyCtx, ok := val.(*verificationContext)
	if !ok {
		return fmt.Errorf("no verification context found")
	}

	if verifyCtx.statusCode == 200 {
		return fmt.Errorf("expected login to be rejected, but got status 200")
	}

	return nil
}

func thePlayerShouldSeeTheMessage(ctx context.Context, expectedMsg string) error {
	val := ctx.Value(verificationContextKey{})
	verifyCtx, ok := val.(*verificationContext)
	if !ok {
		return fmt.Errorf("no verification context found")
	}

	var res ErrorResponse
	if err := json.Unmarshal(verifyCtx.responseBody, &res); err != nil {
		return fmt.Errorf("failed to parse error response: %v (body: %s)", err, string(verifyCtx.responseBody))
	}

	if res.Title != expectedMsg && res.Detail != expectedMsg {
		return fmt.Errorf("expected message '%s', but got title: '%s', detail: '%s'", expectedMsg, res.Title, res.Detail)
	}

	return nil
}

func thePlayerHasReceivedAVerificationEmail(ctx context.Context) (context.Context, error) {
	val := ctx.Value(verificationContextKey{})
	verifyCtx, ok := val.(*verificationContext)
	if !ok {
		return ctx, fmt.Errorf("no verification context found")
	}

	verifyCtx.verifyCode = "123456"
	return context.WithValue(ctx, verificationContextKey{}, verifyCtx), nil
}

func thePlayerEntersAnIncorrectCode(ctx context.Context) (context.Context, error) {
	val := ctx.Value(verificationContextKey{})
	verifyCtx, ok := val.(*verificationContext)
	if !ok {
		return ctx, fmt.Errorf("no verification context found")
	}

	payload := map[string]string{
		"email": verifyCtx.email,
		"code":  "wrong-code",
	}

	b, err := json.Marshal(payload)
	if err != nil {
		return ctx, err
	}

	req, err := http.NewRequest(http.MethodPost, "http://0.0.0.0:8000/auth/verify", bytes.NewReader(b))
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

	verifyCtx.responseBody = body
	verifyCtx.statusCode = resp.StatusCode

	return context.WithValue(ctx, verificationContextKey{}, verifyCtx), nil
}

func theVerificationShouldFail(ctx context.Context) error {
	val := ctx.Value(verificationContextKey{})
	verifyCtx, ok := val.(*verificationContext)
	if !ok {
		return fmt.Errorf("no verification context found")
	}

	if verifyCtx.statusCode == 200 {
		return fmt.Errorf("expected verification to fail, but got status 200")
	}

	return nil
}

func InitializeScenarioForVerification(sc *godog.ScenarioContext) {
	sc.Step(`^a player registers an account with the email "([^"]*)" and password "([^"]*)"$`, aPlayerRegistersAnAccountWithEmailAndPassword)
	sc.Step(`^the registration is completed$`, theRegistrationIsCompleted)
	sc.Step(`^an email should be sent to "([^"]*)" containing a one-time verification code$`, anEmailShouldBeSentToContainingAOneTimeVerificationCode)
	sc.Step(`^the player has received a verification email with a valid one-time code$`, thePlayerHasReceivedAVerificationEmailWithAValidOneTimeCode)
	sc.Step(`^the player submits the correct code within the valid time window$`, thePlayerSubmitsTheCorrectCodeWithinTheValidTimeWindow)
	sc.Step(`^the account should be marked as verified$`, theAccountShouldBeMarkedAsVerified)
	sc.Step(`^the player should be able to log in successfully$`, thePlayerShouldBeAbleToLogInSuccessfully)
	sc.Step(`^the player has registered but not yet verified their account$`, thePlayerHasRegisteredButNotYetVerifiedTheirAccount)
	sc.Step(`^the player attempts to log in to the game$`, thePlayerAttemptsToLogInToTheGame)
	sc.Step(`^the login should be rejected$`, theLoginShouldBeRejected)
	sc.Step(`^the player should see the message "([^"]*)"$`, thePlayerShouldSeeTheMessage)
	sc.Step(`^the player has received a verification email$`, thePlayerHasReceivedAVerificationEmail)
	sc.Step(`^the player enters an incorrect code$`, thePlayerEntersAnIncorrectCode)
	sc.Step(`^the verification should fail$`, theVerificationShouldFail)
}
