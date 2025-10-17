package services_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExampleService_GetExampleData(t *testing.T) {
	data := exampleService.GetExampleData()
	assert.Equal(t, "IM AUTH", data)
}

func TestExampleService_GetSumQuery(t *testing.T) {
	data := exampleService.GetSumQuery()
	assert.Equal(t, "The sum is 2", data)
}
