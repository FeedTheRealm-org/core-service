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

type itemService struct {
	conf *config.Config

	repository items.ItemRepository
	bucketRepo bucket.BucketRepository
}

// NewItemService creates a new instance of ItemService.
func NewItemService(conf *config.Config, repository items.ItemRepository, bucketRepo bucket.BucketRepository) ItemService {
	return &itemService{
		conf:       conf,
		repository: repository,
		bucketRepo: bucketRepo,
	}
}

func (is *itemService) UploadSprites(worldID uuid.UUID, ids []uuid.UUID, files []*multipart.FileHeader) ([]*models.Item, error) {
	if len(ids) != len(files) {
		return nil, fmt.Errorf("number of ids and files must match")
	}

	var result []*models.Item
	for i, fileHeader := range files {
		id := ids[i]
		file, err := fileHeader.Open()
		if err != nil {
			return nil, err
		}

		defer func() {
			_ = file.Close()
		}()

		ext := filepath.Ext(fileHeader.Filename)
		filePath := fmt.Sprintf("/items/worlds/%s/%s%s", worldID.String(), id.String(), ext)
		if err := is.bucketRepo.UploadFile(filePath, fileHeader.Header.Get("Content-Type"), file); err != nil {
			return nil, err
		}

		item := &models.Item{
			Id:  id,
			Url: filePath,
		}
		if err := is.repository.UpsertItem(item); err != nil {
			_ = os.Remove(filePath)
			return nil, err
		}

		logger.Logger.Infof("Item sprite uploaded: %s (ID: %s)", filePath, item.Id)

		result = append(result, item)
	}
	return result, nil
}

func (is *itemService) GetItemById(id uuid.UUID) (*models.Item, error) {
	return is.repository.GetItemById(id)
}

func (is *itemService) GetItemsListByCategory(categoryId uuid.UUID) ([]*models.Item, error) {
	return is.repository.GetItemsListByCategory(categoryId)
}

func (is *itemService) GetAllItems() ([]*models.Item, error) {
	return is.repository.GetAllItems()
}

func (is *itemService) DeleteSprite(id uuid.UUID) error {
	sprite, err := is.repository.GetItemById(id)
	if err != nil {
		return err
	}

	if err := is.repository.DeleteSprite(id); err != nil {
		return err
	}

	if err := is.bucketRepo.DeleteFile(sprite.Url); err != nil {
		logger.Logger.Warnf("Failed to delete sprite file from bucket %s: %v", sprite.Url, err)
	}

	logger.Logger.Infof("Item sprite deleted: %s (ID: %s)", sprite.Url, id)
	return nil
}

func (is *itemService) AddCategory(name string) (*models.ItemCategory, error) {
	return is.repository.AddCategory(name)
}
