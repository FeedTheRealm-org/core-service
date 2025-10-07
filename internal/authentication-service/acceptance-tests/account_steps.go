package acceptance_tests

import (
	"github.com/cucumber/godog"
)

func aSignUpRequestIsMadeWithEmailAndPassword(email, password string) error {
	return nil
}

func aSignUpRequestIsMadeWithAnEmptyEmailAndPassword(password string) error {
	return aSignUpRequestIsMadeWithEmailAndPassword("", password)
}

func aSignUpRequestIsMadeWithEmailAndAnEmptyPassword(email string) error {
	return aSignUpRequestIsMadeWithEmailAndPassword(email, "")
}

func aNewAccountShouldBeCreated() error {
	return nil
}

func theResponseShouldIndicateSuccess() error {
	return nil
}

func theUserShouldBeAbleToLogInWithThoseCredentials() error {
	return nil
}

func theResponseShouldIncludeAnErrorMessage(expectedMsg string) error {
	return nil
}

func theAccountShouldNotBeCreated() error {
	return nil
}

func anAccountAlreadyExistsWithTheEmail(email string) error {
	return nil
}

func InitializeScenarioForAccount(sc *godog.ScenarioContext) {
	sc.Step(`^a sign-up request is made with email "([^"]*)" and password "([^"]*)"$`, aSignUpRequestIsMadeWithEmailAndPassword)
	sc.Step(`^a sign-up request is made with an empty email and password "([^"]*)"$`, aSignUpRequestIsMadeWithAnEmptyEmailAndPassword)
	sc.Step(`^a sign-up request is made with email "([^"]*)" and an empty password$`, aSignUpRequestIsMadeWithEmailAndAnEmptyPassword)
	sc.Step(`^a new account should be created$`, aNewAccountShouldBeCreated)
	sc.Step(`^the response should indicate success$`, theResponseShouldIndicateSuccess)
	sc.Step(`^the user should be able to log in with those credentials$`, theUserShouldBeAbleToLogInWithThoseCredentials)
	sc.Step(`^the response should include an error message "([^"]*)"$`, theResponseShouldIncludeAnErrorMessage)
	sc.Step(`^the account should not be created$`, theAccountShouldNotBeCreated)
	sc.Step(`^an account already exists with the email "([^"]*)"$`, anAccountAlreadyExistsWithTheEmail)
}
