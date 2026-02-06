package models

import (
	"fmt"

	"github.com/FeedTheRealm-org/core-service/config"
	"github.com/FeedTheRealm-org/core-service/internal/assets-service/models"
	"github.com/FeedTheRealm-org/core-service/internal/assets-service/repositories/bucket"
	repo "github.com/FeedTheRealm-org/core-service/internal/assets-service/repositories/models"
	"github.com/google/uuid"
)

type modelsService struct {
	conf *config.Config

	modelsRepository repo.ModelsRepository
	bucketRepo       bucket.BucketRepository
}

// NewModelsService creates a new instance of ModelsService.
func NewModelsService(conf *config.Config, modelsRepository repo.ModelsRepository, bucketRepo bucket.BucketRepository) ModelsService {
	return &modelsService{
		conf:             conf,
		modelsRepository: modelsRepository,
		bucketRepo:       bucketRepo,
	}
}

func (ms *modelsService) PublishModels(worldId uuid.UUID, models []models.Model) ([]models.Model, error) {
	if len(models) == 0 {
		return nil, fmt.Errorf("model list is empty")
	}

	for i := range models {
		models[i].WorldID = worldId

		file, err := models[i].ModelFile.Open()
		if err != nil {
			return nil, err
		}

		defer func() {
			_ = file.Close()
		}()

		filePath := fmt.Sprintf("/worlds/%s/models/%s/model.glb", worldId, models[i].ModelID)
		if err := ms.bucketRepo.UploadFile(filePath, models[i].ModelFile.Header.Get("Content-Type"), file); err != nil {
			return nil, fmt.Errorf("failed uploading model file to bucket: %w", err)
		}

		models[i].ModelURL = filePath
	}

	PublishedModels, err := ms.modelsRepository.PublishModels(models)
	if err != nil {
		return nil, fmt.Errorf("failed publishing models: %w", err)
	}

	return PublishedModels, nil
}

func (ms *modelsService) GetModelsByWorld(worldId uuid.UUID) ([]models.Model, error) {
	modelsList, err := ms.modelsRepository.GetModelsByWorld(worldId)
	if err != nil {
		return nil, fmt.Errorf("failed to get models by world: %w", err)
	}
	return modelsList, nil
}
