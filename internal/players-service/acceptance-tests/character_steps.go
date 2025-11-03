package acceptance_tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/cucumber/godog"
)

/* LOGIN */

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type loginResponse struct {
	Data struct {
		Token string `json:"token"`
	} `json:"data"`
}

/* CHARACTER_INFO */

type CharacterInfoRequest struct {
	CharacterName string `json:"character_name"`
	CharacterBio  string `json:"character_bio"`
}

type CharacterInfoResponse struct {
	Data struct {
		CharacterName string `json:"character_name"`
		CharacterBio  string `json:"character_bio"`
	} `json:"data"`
}

/* SESSION */

type sessionContext struct {
	id       string
	token    string
	charInfo CharacterInfoRequest
}

var baseURL = "http://0.0.0.0:8000"
var ctx sessionContext

func iHaveLoggedInWithEmailAndPassword(email, password string) error {
	resp, err := login(email, password)
	if err != nil {
		return err
	}

	if resp.Data.Token == "" {
		return fmt.Errorf("login failed")
	}

	ctx.token = resp.Data.Token

	return nil
}

func iChangeMyCharacterNameTo(name string) error {
	characterReq := CharacterInfoRequest{
		CharacterName: name,
		CharacterBio:  ctx.charInfo.CharacterBio, // keep existing bio
	}
	_, body, err := httpWithBody("PUT", baseURL+"/player/character", characterReq, ctx.token)
	if err != nil {
		return err
	}

	var resp CharacterInfoResponse
	json.Unmarshal(body, &resp)

	if resp.Data.CharacterName != name {
		return fmt.Errorf("failed to update name, body=%s", string(body))
	}

	ctx.charInfo.CharacterName = name

	return nil
}

func myCharacterNameShouldBeUpdated() error {
	_, body, err := httpGet(baseURL+"/player/character", ctx.token)
	if err != nil {
		return err
	}
	res := &CharacterInfoResponse{}
	json.Unmarshal(body, res)

	if res.Data.CharacterName != ctx.charInfo.CharacterName {
		return fmt.Errorf("expected name %q, got %q",
			ctx.charInfo.CharacterName, res.Data.CharacterName)
	}
	return nil
}

func otherPlayersShouldSeeTheUpdatedName() error {
	loginRes, err := login("test2@email.com", "Password123")
	if err != nil {
		return err
	}

	_, body, err := httpGet(baseURL+"/player/character/"+ctx.id, loginRes.Data.Token)
	if err != nil {
		return err
	}
	res := &CharacterInfoResponse{}
	json.Unmarshal(body, res)

	if res.Data.CharacterName != ctx.charInfo.CharacterName {
		return fmt.Errorf("expected name %q, got %q",
			ctx.charInfo.CharacterName, res.Data.CharacterName)
	}
	return nil
}

func iUpdateMyCharacterBioTo(bio string) error {
	return nil
}

func myCharacterBioShouldBeUpdated() error {
	return nil
}

func theUpdatedBioShouldBeVisibleToOtherPlayersLater() error {
	return myCharacterBioShouldBeUpdated()
}

// length validation steps

func iChangeMyCharacterNameToLessThanOrMoreThanChars(_ string, min, max int) error {
	return nil
}

func iUpdateMyCharacterBioToATextLongerThanCharacters(limit int) error {
	return nil
}

func iShouldSeeAnErrorMessage(msg string) error {
	return nil
}

func InitializeScenarioForCharacter(sc *godog.ScenarioContext) {
	sc.Step(`^I have logged in with email "([^"]*)" and password "([^"]*)"$`, iHaveLoggedInWithEmailAndPassword)

	sc.Step(`^I change my character name to "([^"]*)"$`, iChangeMyCharacterNameTo)
	sc.Step(`^my character name should be updated$`, myCharacterNameShouldBeUpdated)
	sc.Step(`^other players should see the updated name$`, otherPlayersShouldSeeTheUpdatedName)

	sc.Step(`^I update my character bio to "([^"]*)"$`, iUpdateMyCharacterBioTo)
	sc.Step(`^my character bio should be updated$`, myCharacterBioShouldBeUpdated)
	sc.Step(`^the updated bio should be visible to other players later$`, theUpdatedBioShouldBeVisibleToOtherPlayersLater)

	sc.Step(`^I change my character name to "([^"]*)" # less than (\d+) or more than (\d+) chars$`, iChangeMyCharacterNameToLessThanOrMoreThanChars)
	sc.Step(`^I should see an error message "([^"]*)"$`, iShouldSeeAnErrorMessage)

	sc.Step(`^I update my character bio to a text longer than (\d+) characters$`, iUpdateMyCharacterBioToATextLongerThanCharacters)
	sc.Step(`^I should see an error message "([^"]*)"$`, iShouldSeeAnErrorMessage)
}

/* HTTP UTILS */

func login(email, password string) (*loginResponse, error) {
	req := loginRequest{Email: email, Password: password}

	_, body, err := httpWithBody("POST", baseURL+"/auth/login", req, "")
	if err != nil {
		return nil, err
	}

	resp := &loginResponse{}
	json.Unmarshal(body, resp)
	return resp, nil
}

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
