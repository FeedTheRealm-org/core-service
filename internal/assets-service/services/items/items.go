package items

import (
	"fmt"
	"mime/multipart"
	"os"
	"path/filepath"

	"github.com/FeedTheRealm-org/core-service/config"
	"github.com/FeedTheRealm-org/core-service/internal/assets-service/models"
	"github.com/FeedTheRealm-org/core-service/internal/assets-service/repositories/bucket"
	"github.com/FeedTheRealm-org/core-service/internal/assets-service/repositories/items"
	"github.com/FeedTheRealm-org/core-service/internal/utils/logger"
	"github.com/google/uuid"
)

type itemSpritesService struct {
	conf *config.Config

	repository items.ItemSpritesRepository
	bucketRepo bucket.BucketRepository
}

// NewItemSpritesService creates a new instance of ItemSpritesService.
func NewItemSpritesService(conf *config.Config, repository items.ItemSpritesRepository, bucketRepo bucket.BucketRepository) ItemSpritesService {
	return &itemSpritesService{
		conf:       conf,
		repository: repository,
		bucketRepo: bucketRepo,
	}
}

func (iss *itemSpritesService) UploadSprites(worldID uuid.UUID, ids []uuid.UUID, files []*multipart.FileHeader) ([]*models.ItemSprite, error) {
	if len(ids) != len(files) {
		return nil, fmt.Errorf("number of ids and files must match")
	}

	var result []*models.ItemSprite
	for i, fileHeader := range files {
		id := ids[i]
		file, err := fileHeader.Open()
		if err != nil {
			return nil, err
		}
		defer file.Close()

		ext := filepath.Ext(fileHeader.Filename)
		filePath := fmt.Sprintf("/items/worlds/%s/%s%s", worldID.String(), id.String(), ext)
		if err := iss.bucketRepo.UploadFile(filePath, fileHeader.Header.Get("Content-Type"), file); err != nil {
			return nil, err
		}

		sprite := &models.ItemSprite{
			Id:  id,
			Url: filePath,
		}
		if err := iss.repository.UpsertSprite(sprite); err != nil {
			_ = os.Remove(filePath)
			return nil, err
		}

		logger.Logger.Infof("Item sprite uploaded: %s (ID: %s)", filePath, sprite.Id)

		result = append(result, sprite)
	}
	return result, nil
}

func (iss *itemSpritesService) GetSpriteById(id uuid.UUID) (*models.ItemSprite, error) {
	return iss.repository.GetSpriteById(id)
}

func (iss *itemSpritesService) GetAllSprites() ([]models.ItemSprite, error) {
	return iss.repository.GetAllSprites()
}

func (iss *itemSpritesService) GetSpriteFile(id uuid.UUID) (string, error) {
	sprite, err := iss.repository.GetSpriteById(id)
	if err != nil {
		return "", err
	}
	return sprite.Url, nil
}

func (iss *itemSpritesService) DeleteSprite(id uuid.UUID) error {
	sprite, err := iss.repository.GetSpriteById(id)
	if err != nil {
		return err
	}

	if err := iss.repository.DeleteSprite(id); err != nil {
		return err
	}

	if err := iss.bucketRepo.DeleteFile(sprite.Url); err != nil {
		logger.Logger.Warnf("Failed to delete sprite file from bucket %s: %v", sprite.Url, err)
	}

	logger.Logger.Infof("Item sprite deleted: %s (ID: %s)", sprite.Url, id)
	return nil
}
