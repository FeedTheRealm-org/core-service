package input_validation

import "testing"

func TestValidateInvalidCharacters(t *testing.T) {
	cases := []struct {
		name  string
		input string
		want  bool
	}{
		{name: "clean", input: "hello_world", want: false},
		{name: "has_quote", input: "bad'input", want: true},
		{name: "has_slash", input: "bad/input", want: true},
	}

	for _, tc := range cases {
		if got := ValidateInvalidCharacters(tc.input); got != tc.want {
			t.Fatalf("%s: expected %v, got %v", tc.name, tc.want, got)
		}
	}
}

func TestHasSpaces(t *testing.T) {
	cases := []struct {
		name  string
		input string
		want  bool
	}{
		{name: "no_spaces", input: "hello_world", want: false},
		{name: "has_space", input: "hello world", want: true},
		{name: "leading_space", input: " leading", want: true},
	}

	for _, tc := range cases {
		if got := HasSpaces(tc.input); got != tc.want {
			t.Fatalf("%s: expected %v, got %v", tc.name, tc.want, got)
		}
	}
}
