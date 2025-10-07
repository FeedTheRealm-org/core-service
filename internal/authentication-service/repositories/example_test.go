package repositories_test

import (
	"testing"

	"github.com/FeedTheRealm-org/core-service/config"
	"github.com/FeedTheRealm-org/core-service/internal/authentication-service/repositories"
)

func TestExampleRepository_GetExampleRecord(t *testing.T) {
	conf := config.CreateConfig()
	repo, err := repositories.NewExampleRepository(conf)
	if err != nil {
		t.Errorf("Failed to connect: %v", err)
	}

	result := repo.GetExampleRecord()
	expected := "IM AUTH"

	if result != expected {
		t.Errorf("expected %s, got %s", expected, result)
	}
}

func TestExampleRepository_Addition(t *testing.T) {
	conf := config.CreateConfig()
	repo, err := repositories.NewExampleRepository(conf)
	if err != nil {
		t.Errorf("Failed to connect: %v", err)
	}

	sum := repo.GetSumQuery()
	expected := 2

	if sum != expected {
		t.Errorf("expected %d, got %d", expected, sum)
	}
}
