package models

import (
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"

	"github.com/FeedTheRealm-org/core-service/config"
	"github.com/FeedTheRealm-org/core-service/internal/assets-service/models"
	repo "github.com/FeedTheRealm-org/core-service/internal/assets-service/repositories/models"
	"github.com/google/uuid"
)

type modelsService struct {
	conf             *config.Config
	modelsRepository repo.ModelsRepository
}

// NewModelsService creates a new instance of ModelsService.
func NewModelsService(conf *config.Config, modelsRepository repo.ModelsRepository) ModelsService {
	return &modelsService{
		conf:             conf,
		modelsRepository: modelsRepository,
	}
}

func (ms *modelsService) PublishModels(worldId uuid.UUID, models []models.Model) ([]models.Model, error) {
	if len(models) == 0 {
		return nil, fmt.Errorf("model list is empty")
	}

	for i := range models {
		// Create the bucket directory and paths
		baseDir := fmt.Sprintf("bucket/worlds/%s/models/%s", worldId, models[i].ModelID)
		if err := os.MkdirAll(baseDir, os.ModePerm); err != nil {
			return nil, fmt.Errorf("failed creating directory: %w", err)
		}
		modelPath := filepath.Join(baseDir, "model"+filepath.Ext(models[i].ModelFile.Filename))
		materialPath := filepath.Join(baseDir, "material"+filepath.Ext(models[i].MaterialFile.Filename))
		// Save to bucket
		if err := saveUploadedFile(models[i].ModelFile, modelPath); err != nil {
			return nil, fmt.Errorf("failed saving model file: %w", err)
		}
		if err := saveUploadedFile(models[i].MaterialFile, materialPath); err != nil {
			return nil, fmt.Errorf("failed saving material file: %w", err)
		}
		// Update model URLs and world ID
		models[i].ModelURL = modelPath
		models[i].MaterialURL = materialPath
		models[i].WorldID = worldId
	}
	// Save to DB
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

func saveUploadedFile(fileHeader *multipart.FileHeader, path string) error {
	in, err := fileHeader.Open()
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(path)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}
