package materials

import (
	"fmt"
	"mime/multipart"
	"os"
	"path/filepath"

	"github.com/FeedTheRealm-org/core-service/config"
	"github.com/FeedTheRealm-org/core-service/internal/assets-service/models"
	"github.com/FeedTheRealm-org/core-service/internal/assets-service/repositories/bucket"
	"github.com/FeedTheRealm-org/core-service/internal/assets-service/repositories/materials"
	"github.com/FeedTheRealm-org/core-service/internal/utils/logger"
	"github.com/google/uuid"
)

type materialsService struct {
	conf       *config.Config
	repository materials.MaterialsRepository
	bucketRepo bucket.BucketRepository
}

// NewMaterialsService creates a new instance of MaterialsService.
func NewMaterialsService(conf *config.Config, repository materials.MaterialsRepository, bucketRepo bucket.BucketRepository) MaterialsService {
	return &materialsService{
		conf:       conf,
		repository: repository,
		bucketRepo: bucketRepo,
	}
}

func (ms *materialsService) UploadMaterial(worldID uuid.UUID, id uuid.UUID, fileHeader *multipart.FileHeader, userId uuid.UUID) (*models.Material, error) {
	file, err := fileHeader.Open()
	if err != nil {
		return nil, err
	}

	defer func() {
		_ = file.Close()
	}()

	if ms.conf != nil {
		if fileHeader.Size > ms.conf.Assets.MaxUploadSizeBytes {
			return nil, fmt.Errorf("file size exceeds the limit")
		}
	}

	contentType := fileHeader.Header.Get("Content-Type")
	if contentType != "image/png" && contentType != "image/jpeg" && contentType != "application/octet-stream" {
		return nil, fmt.Errorf("file must be PNG, JPEG, or octet-stream format")
	}

	ext := filepath.Ext(fileHeader.Filename)
	filePath := fmt.Sprintf("worlds/%s/materials/%s%s", worldID.String(), id.String(), ext)
	if err := ms.bucketRepo.UploadFile(filePath, contentType, file); err != nil {
		return nil, err
	}

	material := &models.Material{
		ID:        id,
		URL:       fmt.Sprintf("/%s", filePath),
		WorldID:   worldID,
		CreatedBy: userId,
	}
	if err := ms.repository.UpsertMaterial(material); err != nil {
		_ = os.Remove(filePath)
		return nil, err
	}

	logger.Logger.Infof("Material uploaded: %s (ID: %s)", filePath, material.ID)

	return material, nil
}

func (ms *materialsService) GetMaterialByID(id uuid.UUID) (*models.Material, error) {
	return ms.repository.GetMaterialByID(id)
}

func (ms *materialsService) GetMaterialsListByWorld(worldID uuid.UUID, offset int, limit int) ([]*models.Material, error) {
	return ms.repository.GetMaterialsListByWorld(worldID, offset, limit)
}

func (ms *materialsService) DeleteMaterial(id uuid.UUID) error {
	material, err := ms.repository.GetMaterialByID(id)
	if err != nil {
		return err
	}

	if err := ms.bucketRepo.DeleteFile(material.URL); err != nil {
		logger.Logger.Warnf("Failed to delete material file from bucket %s: %v", material.URL, err)
	}

	if err := ms.repository.DeleteMaterial(material); err != nil {
		return err
	}

	logger.Logger.Infof("Material deleted: %s (ID: %s)", material.URL, id)
	return nil
}
