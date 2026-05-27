package materials_test

import (
	"errors"
	"os"
	"testing"

	"github.com/FeedTheRealm-org/core-service/config"
	assets_errors "github.com/FeedTheRealm-org/core-service/internal/assets-service/errors"
	"github.com/FeedTheRealm-org/core-service/internal/assets-service/models"
	materialsrepo "github.com/FeedTheRealm-org/core-service/internal/assets-service/repositories/materials"
	"github.com/FeedTheRealm-org/core-service/internal/utils/logger"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

var materialsConf *config.Config
var materialsDB *config.DB
var materialsRepo materialsrepo.MaterialsRepository

func TestMain(m *testing.M) {
	logger.InitLogger(false)

	materialsConf = config.CreateConfig()
	var err error
	materialsDB, err = config.NewDB(materialsConf)
	if err != nil {
		panic(err)
	}
	materialsRepo = materialsrepo.NewMaterialsRepository(materialsConf, materialsDB)

	clearMaterialsTable()
	code := m.Run()
	clearMaterialsTable()
	os.Exit(code)
}

func clearMaterialsTable() {
	_ = materialsDB.Conn.Exec("TRUNCATE TABLE materials RESTART IDENTITY CASCADE;")
}

func TestMaterialsRepository_UpsertAndGet(t *testing.T) {
	clearMaterialsTable()

	material := &models.Material{
		ID:        uuid.New(),
		WorldID:   uuid.New(),
		Name:      "stone",
		URL:       "/worlds/a/materials/b.png",
		CreatedBy: uuid.New(),
	}

	err := materialsRepo.UpsertMaterial(material)
	assert.NoError(t, err)

	stored, err := materialsRepo.GetMaterialByID(material.ID)
	assert.NoError(t, err)
	assert.Equal(t, material.Name, stored.Name)
}

func TestMaterialsRepository_GetMaterialByID_NotFound(t *testing.T) {
	clearMaterialsTable()

	material, err := materialsRepo.GetMaterialByID(uuid.New())
	assert.Error(t, err)
	assert.Nil(t, material)

	var notFound *assets_errors.MaterialNotFound
	assert.True(t, errors.As(err, &notFound))
}

func TestMaterialsRepository_GetMaterialsListByWorldAndDelete(t *testing.T) {
	clearMaterialsTable()

	worldID := uuid.New()
	otherWorld := uuid.New()
	assert.NoError(t, materialsRepo.UpsertMaterial(&models.Material{ID: uuid.New(), WorldID: worldID, Name: "stone", URL: "/a.png", CreatedBy: uuid.New()}))
	assert.NoError(t, materialsRepo.UpsertMaterial(&models.Material{ID: uuid.New(), WorldID: otherWorld, Name: "wood", URL: "/b.png", CreatedBy: uuid.New()}))

	materials, err := materialsRepo.GetMaterialsListByWorld(worldID, 0, 10)
	assert.NoError(t, err)
	assert.Len(t, materials, 1)

	assert.NoError(t, materialsRepo.DeleteMaterial(materials[0]))
}
