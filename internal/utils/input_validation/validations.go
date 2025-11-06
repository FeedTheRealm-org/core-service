package input_validation

import "strings"

const DangerousCharacters = "'\",.<>/`~\\"

func ValidateInvalidCharacters(input string) bool {
	for _, char := range input {
		if strings.ContainsRune(DangerousCharacters, char) {
			return true
		}
	}
	return false
}
