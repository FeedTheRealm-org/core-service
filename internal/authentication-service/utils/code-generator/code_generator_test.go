package code_generator

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateCode_ReturnsStringOfLength8(t *testing.T) {
	randFn := func() int { return 0 }
	code := GenerateCode(randFn)

	assert.Equal(t, 8, len(code))
}

func TestGenerateCode_OnlyUppercaseAlphanumeric(t *testing.T) {
	randFn := func() int { return 0 }
	code := GenerateCode(randFn)

	matched, err := regexp.MatchString("^[0-9A-Z]{8}$", code)
	assert.NoError(t, err)
	assert.True(t, matched, "code contains invalid characters: %s", code)
}

func TestGenerateCode_DeterministicWithSameRandFn(t *testing.T) {
	randFn := func() int { return 42 }

	code1 := GenerateCode(randFn)
	code2 := GenerateCode(randFn)

	assert.Equal(t, code1, code2)
}

func TestGenerateCode_DifferentRandFnGivesDifferentCode(t *testing.T) {
	randFn1 := func() int { return 1 }
	randFn2 := func() int { return 2 }

	code1 := GenerateCode(randFn1)
	code2 := GenerateCode(randFn2)

	assert.NotEqual(t, code1, code2)
}

func TestGenerateCode_WhenRandFnReturns0_ReturnsAllZeros(t *testing.T) {
	randFn := func() int { return 0 }

	code := GenerateCode(randFn)

	assert.Equal(t, "00000000", code)
}

func TestGenerateCode_WhenRandFnReturnsNegative_StillWorks(t *testing.T) {
	randFn := func() int { return -5 }

	code := GenerateCode(randFn)

	assert.Equal(t, 8, len(code))
	matched, _ := regexp.MatchString("^[0-9A-Z]{8}$", code)
	assert.True(t, matched)
}
