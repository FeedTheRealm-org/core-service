package code_generator

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateCode_Produces8AlphanumericChars(t *testing.T) {
	randFn := func() int {
		return 12345678
	}

	code := GenerateCode(randFn)
	assert.Equal(t, 8, len(code))

	matched, err := regexp.MatchString("^[0-9A-Za-z]{8}$", code)
	assert.Nil(t, err)
	assert.True(t, matched, "code should be 8 alphanumeric characters")
}

func TestGenerateCode_DeterministicForStaticSeed(t *testing.T) {
	randFn := func() int {
		return 12345678
	}

	code1 := GenerateCode(randFn)
	code2 := GenerateCode(randFn)
	assert.Equal(t, code1, code2)
}
