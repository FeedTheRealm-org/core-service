package models_test

import (
	"os"
	"testing"

	"github.com/FeedTheRealm-org/core-service/config"
	assetmodels "github.com/FeedTheRealm-org/core-service/internal/assets-service/models"
	modelsrepo "github.com/FeedTheRealm-org/core-service/internal/assets-service/repositories/models"
	"github.com/FeedTheRealm-org/core-service/internal/utils/logger"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

var modelsConf *config.Config
var modelsDB *config.DB
var modelsRepo modelsrepo.ModelsRepository

func TestMain(m *testing.M) {
	logger.InitLogger(false)

	modelsConf = config.CreateConfig()
	var err error
	modelsDB, err = config.NewDB(modelsConf)
	if err != nil {
		panic(err)
	}
	modelsRepo = modelsrepo.NewModelsRepository(modelsConf, modelsDB)

	clearModelsTable()
	code := m.Run()
	clearModelsTable()
	os.Exit(code)
}

func clearModelsTable() {
	_ = modelsDB.Conn.Exec("TRUNCATE TABLE models RESTART IDENTITY CASCADE;")
}

func TestModelsRepository_UploadAndGetByWorld(t *testing.T) {
	clearModelsTable()

	worldID := uuid.New()
	model := assetmodels.Model{
		Id:        uuid.New(),
		WorldID:   worldID,
		Url:       "/worlds/a/models/b.glb",
		CreatedBy: uuid.New(),
	}

	stored, err := modelsRepo.UploadModel(model)
	assert.NoError(t, err)
	assert.NotNil(t, stored)

	list, err := modelsRepo.GetModelsByWorld(worldID)
	assert.NoError(t, err)
	assert.Len(t, list, 1)
}
