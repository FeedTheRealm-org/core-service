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

func (is *itemService) UploadSprite(worldID uuid.UUID, categoryId uuid.UUID, id uuid.UUID, fileHeader *multipart.FileHeader) (*models.Item, error) {
	file, err := fileHeader.Open()
	if err != nil {
		return nil, err
	}

	defer func() {
		_ = file.Close()
	}()

	if is.conf != nil {
		if fileHeader.Size > is.conf.Assets.MaxUploadSizeBytes {
			return nil, fmt.Errorf("file size exceeds the limit")
		}
	}

	contentType := fileHeader.Header.Get("Content-Type")
	if contentType != "image/png" && contentType != "image/jpeg" && contentType != "application/octet-stream" {
		return nil, fmt.Errorf("file must be PNG, JPEG, or octet-stream format")
	}

	ext := filepath.Ext(fileHeader.Filename)
	filePath := fmt.Sprintf("/items/worlds/%s/categories/%s/%s%s", worldID.String(), categoryId.String(), id.String(), ext)
	if err := is.bucketRepo.UploadFile(filePath, contentType, file); err != nil {
		return nil, err
	}

	item := &models.Item{
		Id:         id,
		Url:        filePath,
		WorldID:    worldID,
		CategoryID: categoryId,
	}
	if err := is.repository.UpsertItem(item); err != nil {
		_ = os.Remove(filePath)
		return nil, err
	}

	logger.Logger.Infof("Item sprite uploaded: %s (ID: %s)", filePath, item.Id)

	return item, nil
}

func (is *itemService) GetItemById(id uuid.UUID) (*models.Item, error) {
	return is.repository.GetItemById(id)
}

func (is *itemService) GetItemsListByCategory(worldId uuid.UUID, categoryId uuid.UUID) ([]*models.Item, error) {
	return is.repository.GetItemsListByCategory(worldId, categoryId)
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
