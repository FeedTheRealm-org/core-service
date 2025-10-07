package acceptance_tests

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/cucumber/godog"
)

type testContextKey struct{}

func iGoToQueryPage(ctx context.Context) (context.Context, error) {
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

func iGetExampleMessage(ctx context.Context) error {
	val := ctx.Value(testContextKey{})
	body, ok := val.([]byte)
	if !ok {
		return fmt.Errorf("no result found in context")
	}

	type response struct {
		Message string `json:"message"`
	}

	var res response
	if err := json.Unmarshal(body, &res); err != nil {
		return fmt.Errorf("failed to parse response: %v", err)
	}

	if res.Message != "IM AUTH" {
		return fmt.Errorf("unexpected result: %s", res.Message)
	}

	return nil
}

func iGoToExamplePage(ctx context.Context) (context.Context, error) {
	resp, err := http.Get("http://0.0.0.0:8000/auth/example-msg")
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

func iGetSumResult(ctx context.Context) error {
	val := ctx.Value(testContextKey{})
	body, ok := val.([]byte)
	if !ok {
		return fmt.Errorf("no result found in context")
	}

	type response struct {
		Message string `json:"message"`
	}

	var res response
	if err := json.Unmarshal(body, &res); err != nil {
		return fmt.Errorf("failed to parse response: %v", err)
	}

	if res.Message != "The sum is 2" {
		return fmt.Errorf("unexpected result: %s", res.Message)
	}

	return nil
}

func InitializeScenarioForExample(sc *godog.ScenarioContext) {
	sc.Step(`^I get example message$`, iGetExampleMessage)
	sc.Step(`^I get the sum$`, iGetSumResult)
	sc.Step(`^I go to example page$`, iGoToExamplePage)
	sc.Step(`^I go to query page$`, iGoToQueryPage)
}
