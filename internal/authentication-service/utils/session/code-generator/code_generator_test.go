package session

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateCode_RandomCode(t *testing.T) {
	randFn := func() int {
		return 123456
	}

	code := GenerateCode(randFn)
	assert.Equal(t, "123456", code)
}

func TestGenerateCode_ZeroCode(t *testing.T) {
	randFn := func() int {
		return 0
	}

	code := GenerateCode(randFn)
	assert.Equal(t, "000000", code)
}

func TestGenerateCode_Padding(t *testing.T) {
	randFn := func() int {
		return 42
	}

	code := GenerateCode(randFn)
	assert.Equal(t, "000042", code)
}

func TestGeneratorCode_LargeNumber(t *testing.T) {
	randFn := func() int {
		return 1234567
	}

	code := GenerateCode(randFn)
	assert.Equal(t, "234567", code)
}

func TestGenerateCode_NegativeNumber(t *testing.T) {
	randFn := func() int {
		return -1
	}

	code := GenerateCode(randFn)
	assert.Equal(t, "999999", code)
}
