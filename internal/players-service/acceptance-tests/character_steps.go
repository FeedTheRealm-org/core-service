package acceptance_tests

import (
	"time"

	"github.com/cucumber/godog"
)

type loginResponse struct {
	Data struct {
		Token string `json:"token"`
	} `json:"data"`
}

type sessionContext struct {
	token     string
	loginTime time.Time
}

func anAccountAlreadyExistsWithTheEmailAndPassword(arg1, arg2 string) error {
	return godog.ErrPending
}

func iAmOnTheCharacterProfilePage() error {
	return godog.ErrPending
}

func iChangeMyCharacterNameTo(arg1 string) error {
	return godog.ErrPending
}

func iChangeMyCharacterNameToLessThanOrMoreThanChars(arg1 string, arg2, arg3 int) error {
	return godog.ErrPending
}

func iHaveLoggedInWithEmailAndPassword(arg1, arg2 string) error {
	return godog.ErrPending
}

func iSaveTheCharacterProfile() error {
	return godog.ErrPending
}

func iShouldSeeAnErrorMessage(arg1 string) error {
	return godog.ErrPending
}

func iUpdateMyCharacterBioTo(arg1 string) error {
	return godog.ErrPending
}

func iUpdateMyCharacterBioToATextLongerThanCharacters(arg1 int) error {
	return godog.ErrPending
}

func myCharacterBioShouldBeUpdated() error {
	return godog.ErrPending
}

func myCharacterNameShouldBeUpdated() error {
	return godog.ErrPending
}

func otherPlayersShouldSeeTheUpdatedName() error {
	return godog.ErrPending
}

func theUpdatedBioShouldBeVisibleToOtherPlayersLater() error {
	return godog.ErrPending
}

func InitializeScenarioForCharacter(ctx *godog.ScenarioContext) {
	ctx.Step(`^an account already exists with the email "([^"]*)" and password "([^"]*)"$`, anAccountAlreadyExistsWithTheEmailAndPassword)
	ctx.Step(`^I am on the character profile page$`, iAmOnTheCharacterProfilePage)
	ctx.Step(`^I change my character name to "([^"]*)"$`, iChangeMyCharacterNameTo)
	ctx.Step(`^I change my character name to "([^"]*)" # less than (\d+) or more than (\d+) chars$`, iChangeMyCharacterNameToLessThanOrMoreThanChars)
	ctx.Step(`^I have logged in with email "([^"]*)" and password "([^"]*)"$`, iHaveLoggedInWithEmailAndPassword)
	ctx.Step(`^I save the character profile$`, iSaveTheCharacterProfile)
	ctx.Step(`^I should see an error message "([^"]*)"$`, iShouldSeeAnErrorMessage)
	ctx.Step(`^I update my character bio to "([^"]*)"$`, iUpdateMyCharacterBioTo)
	ctx.Step(`^I update my character bio to a text longer than (\d+) characters$`, iUpdateMyCharacterBioToATextLongerThanCharacters)
	ctx.Step(`^my character bio should be updated$`, myCharacterBioShouldBeUpdated)
	ctx.Step(`^my character name should be updated$`, myCharacterNameShouldBeUpdated)
	ctx.Step(`^other players should see the updated name$`, otherPlayersShouldSeeTheUpdatedName)
	ctx.Step(`^the updated bio should be visible to other players later$`, theUpdatedBioShouldBeVisibleToOtherPlayersLater)
}
