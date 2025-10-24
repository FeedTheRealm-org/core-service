package credential_validation

import (
	"regexp"
)

func IsValidEmail(email string) bool {
	re := regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	return re.MatchString(email)
}

func containsLetter(s string) bool {
	letterRe := regexp.MustCompile(`[A-Za-z]`)
	return letterRe.MatchString(s)
}

func containsNumber(s string) bool {
	numberRe := regexp.MustCompile(`\d`)
	return numberRe.MatchString(s)
}

func IsValidPassword(password string) bool {
	if len(password) < 8 {
		return false
	}

	if !containsLetter(password) || !containsNumber(password) {
		return false
	}

	return true
}
