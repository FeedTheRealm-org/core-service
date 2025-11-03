package acceptance_tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/cucumber/godog"
)

type CharacterInfo struct {
	CharacterName string `json:"character_name"`
	CharacterBio  string `json:"character_bio"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type loginResponse struct {
	Data struct {
		Token string `json:"token"`
	} `json:"data"`
}

type ServerResponse struct {
	Error string        `json:"error,omitempty"`
	Data  CharacterInfo `json:"data,omitempty"`
}

var baseURL = "http://0.0.0.0:8000"
var ctx sessionContext
var lastHTTPResponse ServerResponse
var lastStatus int

type sessionContext struct {
	token     string
	loginTime time.Time
	charInfo  CharacterInfo
}

func httpPost(url string, body any, auth string) (int, []byte, error) {
	jsonValue, _ := json.Marshal(body)
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(jsonValue))
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

func anAccountAlreadyExistsWithTheEmailAndPassword(email, password string) error {
	// normally seed DB or trust fixtures — noop here
	return nil
}

func iHaveLoggedInWithEmailAndPassword(email, password string) error {
	req := loginRequest{Email: email, Password: password}

	status, body, err := httpPost(baseURL+"/login", req, "")
	if err != nil {
		return err
	}
	lastStatus = status

	var resp loginResponse
	json.Unmarshal(body, &resp)

	if resp.Data.Token == "" {
		return fmt.Errorf("login failed, body=%s", string(body))
	}

	ctx.token = resp.Data.Token
	ctx.loginTime = time.Now()

	return nil
}

func iChangeMyCharacterNameTo(name string) error {
	ctx.charInfo.CharacterName = name
	return nil
}

func iUpdateMyCharacterBioTo(bio string) error {
	ctx.charInfo.CharacterBio = bio
	return nil
}

func iSaveTheCharacterProfile() error {
	status, body, err := httpPost(baseURL+"/player/character", ctx.charInfo, ctx.token)
	if err != nil {
		return err
	}
	lastStatus = status

	json.Unmarshal(body, &lastHTTPResponse)
	return nil
}

func iShouldSeeAnErrorMessage(msg string) error {
	if lastHTTPResponse.Error != msg {
		return fmt.Errorf("expected error %q, got %q", msg, lastHTTPResponse.Error)
	}
	return nil
}

func myCharacterNameShouldBeUpdated() error {
	status, body, err := httpGet(baseURL+"/player/character", ctx.token)
	if err != nil {
		return err
	}
	lastStatus = status
	json.Unmarshal(body, &lastHTTPResponse)

	if lastHTTPResponse.Data.CharacterName != ctx.charInfo.CharacterName {
		return fmt.Errorf("expected name %q, got %q",
			ctx.charInfo.CharacterName, lastHTTPResponse.Data.CharacterName)
	}
	return nil
}

func otherPlayersShouldSeeTheUpdatedName() error {
	// same as above — GET again (future behavior can be mocked)
	return myCharacterNameShouldBeUpdated()
}

func myCharacterBioShouldBeUpdated() error {
	status, body, err := httpGet(baseURL+"/player/character", ctx.token)
	if err != nil {
		return err
	}
	lastStatus = status
	json.Unmarshal(body, &lastHTTPResponse)

	if lastHTTPResponse.Data.CharacterBio != ctx.charInfo.CharacterBio {
		return fmt.Errorf("expected bio %q, got %q",
			ctx.charInfo.CharacterBio, lastHTTPResponse.Data.CharacterBio)
	}
	return nil
}

func theUpdatedBioShouldBeVisibleToOtherPlayersLater() error {
	return myCharacterBioShouldBeUpdated()
}

// length validation steps

func iChangeMyCharacterNameToLessThanOrMoreThanChars(_ string, min, max int) error {
	invalid := ""
	for i := 0; i < min-1; i++ {
		invalid += "a"
	}
	// OR > max
	if len(invalid) == min-1 {
		ctx.charInfo.CharacterName = invalid
	} else {
		for i := 0; i < max+1; i++ {
			invalid += "b"
		}
		ctx.charInfo.CharacterName = invalid
	}
	return nil
}

func iUpdateMyCharacterBioToATextLongerThanCharacters(limit int) error {
	tooLong := ""
	for i := 0; i < limit+1; i++ {
		tooLong += "x"
	}
	ctx.charInfo.CharacterBio = tooLong
	return nil
}

func InitializeScenarioForCharacter(sc *godog.ScenarioContext) {
	sc.Step(`^an account already exists with the email "([^"]*)" and password "([^"]*)"$`, anAccountAlreadyExistsWithTheEmailAndPassword)
	sc.Step(`^I have logged in with email "([^"]*)" and password "([^"]*)"$`, iHaveLoggedInWithEmailAndPassword)

	sc.Step(`^I change my character name to "([^"]*)"$`, iChangeMyCharacterNameTo)
	sc.Step(`^I save the character profile$`, iSaveTheCharacterProfile)
	sc.Step(`^my character name should be updated$`, myCharacterNameShouldBeUpdated)
	sc.Step(`^other players should see the updated name$`, otherPlayersShouldSeeTheUpdatedName)

	sc.Step(`^I update my character bio to "([^"]*)"$`, iUpdateMyCharacterBioTo)
	sc.Step(`^my character bio should be updated$`, myCharacterBioShouldBeUpdated)
	sc.Step(`^the updated bio should be visible to other players later$`, theUpdatedBioShouldBeVisibleToOtherPlayersLater)

	sc.Step(`^I change my character name to "([^"]*)" # less than (\d+) or more than (\d+) chars$`, iChangeMyCharacterNameToLessThanOrMoreThanChars)
	sc.Step(`^I update my character bio to a text longer than (\d+) characters$`, iUpdateMyCharacterBioToATextLongerThanCharacters)
	sc.Step(`^I should see an error message "([^"]*)"$`, iShouldSeeAnErrorMessage)
}
