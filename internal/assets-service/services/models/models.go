package models

import (
	"fmt"

	"github.com/FeedTheRealm-org/core-service/config"
	"github.com/FeedTheRealm-org/core-service/internal/assets-service/models"
	"github.com/FeedTheRealm-org/core-service/internal/assets-service/repositories/bucket"
	repo "github.com/FeedTheRealm-org/core-service/internal/assets-service/repositories/models"
	"github.com/FeedTheRealm-org/core-service/internal/utils/logger"
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

func (ms *modelsService) UploadModels(worldId uuid.UUID, models []models.Model) ([]models.Model, error) {
	if len(models) == 0 {
		return nil, fmt.Errorf("model list is empty")
	}

	if err := ms.uploadToBucket(worldId, models); err != nil {
		return nil, err
	}

	publishedModels, err := ms.modelsRepository.UploadModels(models)

	logger.Logger.Infof("SERVICE: Published %d models to the db", len(publishedModels))

	if err != nil {
		// If publishing to the database fails, roll back the file uploads
		for _, model := range models {
			filePath := fmt.Sprintf("worlds/%s/models/%s/model.glb", worldId, model.Id)
			_ = ms.bucketRepo.DeleteFile(filePath)
		}
		return nil, fmt.Errorf("failed publishing models: %w", err)
	}
	return publishedModels, nil
}

func (ms *modelsService) GetModelsByWorld(worldId uuid.UUID) ([]models.Model, error) {
	modelsList, err := ms.modelsRepository.GetModelsByWorld(worldId)
	if err != nil {
		return nil, fmt.Errorf("failed to get models by world: %w", err)
	}
	return modelsList, nil
}

// ---- Private methods ----

func (ms *modelsService) uploadToBucket(worldId uuid.UUID, models []models.Model) error {
	uploadedFilePaths := []string{}

	for i := range models {
		models[i].WorldID = worldId

		file, err := models[i].ModelFile.Open()
		if err != nil {
			return err
		}
		filePath := fmt.Sprintf("worlds/%s/models/%s/model.glb", worldId, models[i].Id)
		if err := ms.bucketRepo.UploadFile(filePath, models[i].ModelFile.Header.Get("Content-Type"), file); err != nil {
			_ = file.Close()
			// Rollback previously uploaded files from this batch
			for _, uploadedPath := range uploadedFilePaths {
				_ = ms.bucketRepo.DeleteFile(uploadedPath)
			}
			return fmt.Errorf("failed uploading model file to bucket: %w", err)
		}
		_ = file.Close()
		uploadedFilePaths = append(uploadedFilePaths, filePath)
		models[i].Url = fmt.Sprintf("/%s", filePath)
	}

	return nil
}
