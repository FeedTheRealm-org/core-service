package acceptance_tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/cucumber/godog"
)

// sessionContext holds auth info used across steps
type sessionContext struct {
	id    string
	token string
}

/* LOGIN helpers and types */
type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type loginResponse struct {
	Data struct {
		AccessToken string `json:"access_token"`
		Id          string `json:"id"`
	} `json:"data"`
}

type world struct {
	ID        string `json:"id"`
	UserID    string `json:"user_id"`
	Name      string `json:"name"`
	Data      string `json:"data"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}
type worldResponse struct {
	Data world `json:"data"`
}

type worldsListResponse struct {
	Data struct {
		Worlds []world `json:"worlds"`
		Amount int     `json:"amount"`
		Limit  int     `json:"limit"`
		Offset int     `json:"offset"`
	} `json:"data"`
}

var baseURL = "http://0.0.0.0:8000"
var ctx sessionContext
var response worldResponse
var errorResponse error

func iHaveLoggedInWithEmailAndPassword(email, password string) error {
	id, token, err := login(email, password)
	if err != nil {
		return err
	}
	ctx.token = token
	ctx.id = id
	return nil
}

/* WORLD steps */
func iPublishAWorld(name string) error {
	if ctx.token == "" {
		return fmt.Errorf("no logged in user")
	}

	worldReq := map[string]any{
		"file_name": name,
		"data": map[string]any{
			"worldName": name,
		},
	}

	status, body, err := httpWithBody("POST", baseURL+"/world", worldReq, ctx.token)
	if err != nil {
		errorResponse = err
		return nil
	}
	if status != http.StatusCreated {
		return nil
	}
	if len(body) > 0 {
		if err := json.Unmarshal(body, &response); err != nil {
			return nil
		}
	}
	return nil
}

func theWorldShouldBePublished() error {
	if response.Data.ID == "" {
		return fmt.Errorf("world ID is empty — world was not published correctly")
	}

	world, err := findWorldById(response.Data.ID)
	if err != nil {
		return fmt.Errorf("failed to retrieve world by ID: %w", err)
	}
	if world == nil {
		return fmt.Errorf("world not found by ID %s", response.Data.ID)
	}

	if world.Data.ID != response.Data.ID {
		return fmt.Errorf("ID mismatch: expected %s, got %s",
			response.Data.ID, world.Data.ID)
	}
	if world.Data.Name != response.Data.Name {
		return fmt.Errorf("name mismatch: expected %s, got %s",
			response.Data.Name, world.Data.Name)
	}
	if world.Data.UserID != response.Data.UserID {
		return fmt.Errorf("user_id mismatch: expected %s, got %s",
			response.Data.UserID, world.Data.UserID)
	}
	if world.Data.Data != response.Data.Data {
		return fmt.Errorf("data mismatch: expected %s, got %s",
			response.Data.Data, world.Data.Data)
	}

	return nil
}

func otherPlayersShouldSeeTheWorldInListings() error {
	listResp, err := getAllWorlds(0, 10)
	if err != nil {
		return fmt.Errorf("failed to fetch worlds list: %w", err)
	}

	var found *world
	for _, w := range listResp.Data.Worlds {
		if w.ID == response.Data.ID {
			found = &w
			break
		}
	}

	if found == nil {
		return fmt.Errorf("expected world ID %s to appear in listing, but it was not found", response.Data.ID)
	}

	if found.UserID != response.Data.UserID {
		return fmt.Errorf("user_id mismatch: expected %s, got %s", response.Data.UserID, found.UserID)
	}

	if found.Name != response.Data.Name {
		return fmt.Errorf("name mismatch: expected %s, got %s", response.Data.Name, found.Name)
	}

	if found.Data != response.Data.Data {
		return fmt.Errorf("data mismatch:\nexpected: %s\n     got: %s", response.Data.Data, found.Data)
	}

	return nil
}

func iShouldSeeAnErrorMessage(errorMessage string) error {

	if errorResponse == nil {
		return fmt.Errorf("expected an error but none occurred")
	}

	if errorResponse.Error() != errorMessage {
		return fmt.Errorf("expected error message '%s', but got '%s'",
			errorMessage, errorResponse.Error())
	}
	return nil
}

/* ---------- Endpoint Helpers ---------- */
func findWorldById(id string) (*worldResponse, error) {
	status, body, err := httpGet(baseURL+"/world/"+id, ctx.token)
	if err != nil {
		return nil, err
	}

	if status != http.StatusOK {
		return nil, fmt.Errorf("get world failed, status=%d, body=%s", status, body)
	}

	var resp worldResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

func getAllWorlds(offset, limit int) (*worldsListResponse, error) {
	url := fmt.Sprintf("%s/world?offset=%d&limit=%d", baseURL, offset, limit)

	status, body, err := httpGet(url, ctx.token)
	if err != nil {
		return nil, err
	}

	if status != http.StatusOK {
		return nil, fmt.Errorf("get worlds failed, status=%d, body=%s", status, body)
	}

	var resp worldsListResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

func login(email, password string) (string, string, error) {
	loginReq := loginRequest{Email: email, Password: password}
	jsonValue, _ := json.Marshal(loginReq)

	req, _ := http.NewRequest("POST", baseURL+"/auth/login", bytes.NewBuffer(jsonValue))
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", "", err
	}
	body, _ := io.ReadAll(resp.Body)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", "", fmt.Errorf("login failed, status=%d, body=%s", resp.StatusCode, string(body))
	}
	var loginResp loginResponse
	if err := json.Unmarshal(body, &loginResp); err != nil {
		return "", "", err
	}
	return loginResp.Data.Id, loginResp.Data.AccessToken, nil
}

// --------- HTTP helpers reused by steps ----------

func httpWithBody(method, url string, body any, auth string) (int, []byte, error) {
	jsonValue, _ := json.Marshal(body)
	req, _ := http.NewRequest(method, url, bytes.NewBuffer(jsonValue))
	req.Header.Set("Content-Type", "application/json")
	if auth != "" {
		req.Header.Set("Authorization", "Bearer "+auth)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, nil, err
	}
	defer resp.Body.Close()
	respBytes, _ := io.ReadAll(resp.Body)
	return resp.StatusCode, respBytes, nil
}

func httpGet(url string, auth string) (int, []byte, error) {
	req, _ := http.NewRequest("GET", url, nil)
	if auth != "" {
		req.Header.Set("Authorization", "Bearer "+auth)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, nil, err
	}
	defer resp.Body.Close()
	respBytes, _ := io.ReadAll(resp.Body)
	return resp.StatusCode, respBytes, nil
}

// -------- Scenario Initialization --------

func InitializeScenarioForWorld(sc *godog.ScenarioContext) {
	sc.Step(`^I have logged in with email "([^\\"]*)" and password "([^\\"]*)"$`, iHaveLoggedInWithEmailAndPassword)

	sc.Step(`^I publish a world with name "([^\\"]*)"$`, iPublishAWorld)
	sc.Step(`^the world should be published$`, theWorldShouldBePublished)
	sc.Step(`^other players should see the world in the world listings$`, otherPlayersShouldSeeTheWorldInListings)
	sc.Step(`^I should see an error message$`, iShouldSeeAnErrorMessage)

}
