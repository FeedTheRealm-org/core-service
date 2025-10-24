package credential_validation

import (
	"regexp"
)

type EmptyEmailError struct{}

func (EmptyEmailError) Error() string { return "Empty email" }

type InvalidEmailError struct{}

func (InvalidEmailError) Error() string { return "Invalid email" }

type EmptyPasswordError struct{}

func (EmptyPasswordError) Error() string { return "Empty password" }

type PasswordTooShortError struct{}

func (PasswordTooShortError) Error() string { return "Password too short" }

type PasswordNoLetterError struct{}

func (PasswordNoLetterError) Error() string { return "Password must contain at least one letter" }

type PasswordNoNumberError struct{}

func (PasswordNoNumberError) Error() string { return "Password must contain at least one number" }

func IsValidEmail(email string) error {
	if email == "" {
		return &EmptyEmailError{}
	}

	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

	if !emailRegex.MatchString(email) {
		return &InvalidEmailError{}
	}

	return nil
}

func containsLetter(s string) bool {
	letterRe := regexp.MustCompile(`[A-Za-z]`)
	return letterRe.MatchString(s)
}

func containsNumber(s string) bool {
	numberRe := regexp.MustCompile(`\d`)
	return numberRe.MatchString(s)
}

func IsValidPassword(password string) error {
	if password == "" {
		return &EmptyPasswordError{}
	}

	if len(password) < 8 {
		return &PasswordTooShortError{}
	}

	if !containsLetter(password) {
		return &PasswordNoLetterError{}
	}

	if !containsNumber(password) {
		return &PasswordNoNumberError{}
	}

	return nil
}
