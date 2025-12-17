package input_validation

import "strings"

const DangerousCharacters = "'\"/`\\"

func ValidateInvalidCharacters(input string) bool {
	for _, char := range input {
		if strings.ContainsRune(DangerousCharacters, char) {
			return true
		}
	}
	return false
}

func HasSpaces(input string) bool {
	for _, char := range input {
		if char == ' ' {
			return true
		}
	}
	return false
}
