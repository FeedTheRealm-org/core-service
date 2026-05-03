package models

import (
	"fmt"
	"mime/multipart"
	"path/filepath"

	"github.com/FeedTheRealm-org/core-service/config"
	"github.com/FeedTheRealm-org/core-service/internal/assets-service/dtos"
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

func (ms *modelsService) UploadModel(modelRequest dtos.ModelRequest) (*models.Model, error) {
	if modelRequest.Id == uuid.Nil {
		return nil, fmt.Errorf("model_id is required")
	}

	if filepath.Ext(modelRequest.ModelFile.Filename) != ".glb" {
		return nil, fmt.Errorf("model file must be a .glb file")
	}

	url, err := ms.uploadToBucket(modelRequest.WorldID, modelRequest.ModelFile, modelRequest.Id)
	if err != nil {
		logger.Logger.Errorf("SERVICE: Failed to upload model to bucket: %v", err)
		return nil, err
	}

	model := models.Model{
		Id:        modelRequest.Id,
		Url:       url,
		WorldID:   modelRequest.WorldID,
		CreatedBy: modelRequest.CreatedBy,
	}

	publishedModel, err := ms.modelsRepository.UploadModel(model)
	if err != nil {
		filePath := fmt.Sprintf("worlds/%s/models/%s/%s", modelRequest.WorldID, model.Id, modelRequest.ModelFile.Filename)
		_ = ms.bucketRepo.DeleteFile(filePath)
		return nil, fmt.Errorf("failed publishing model: %w", err)
	}

	logger.Logger.Infof("Published model %s to the db", publishedModel.Id)
	return publishedModel, nil
}

func (ms *modelsService) GetModelsByWorld(worldId uuid.UUID) ([]models.Model, error) {
	modelsList, err := ms.modelsRepository.GetModelsByWorld(worldId)
	if err != nil {
		return nil, fmt.Errorf("failed to get models by world: %w", err)
	}
	return modelsList, nil
}

// ---- Private methods ----

func (ms *modelsService) uploadToBucket(worldId uuid.UUID, modelFile *multipart.FileHeader, modelId uuid.UUID) (string, error) {
	file, err := modelFile.Open()
	if err != nil {
		return "", err
	}
	defer file.Close()

	filePath := fmt.Sprintf("worlds/%s/models/%s/%s", worldId, modelId, modelFile.Filename)
	if err := ms.bucketRepo.UploadFile(filePath, modelFile.Header.Get("Content-Type"), file); err != nil {
		return "", fmt.Errorf("failed uploading model file to bucket: %w", err)
	}

	return fmt.Sprintf("/%s", filePath), nil
}
