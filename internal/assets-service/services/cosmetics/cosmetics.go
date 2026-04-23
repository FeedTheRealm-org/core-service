package cosmetics

import (
	"fmt"
	"mime/multipart"

	"github.com/FeedTheRealm-org/core-service/config"
	assets_errors "github.com/FeedTheRealm-org/core-service/internal/assets-service/errors"
	"github.com/FeedTheRealm-org/core-service/internal/assets-service/models"
	"github.com/FeedTheRealm-org/core-service/internal/assets-service/repositories/bucket"
	"github.com/FeedTheRealm-org/core-service/internal/assets-service/repositories/cosmetics"
	"github.com/FeedTheRealm-org/core-service/internal/utils/logger"

	"github.com/google/uuid"
)

type cosmeticsService struct {
	conf *config.Config

	cosmeticsRepository cosmetics.CosmeticsRepository
	bucketRepo          bucket.BucketRepository
}

// NewCosmeticsService creates a new instance of CosmeticsService.
func NewCosmeticsService(conf *config.Config, cosmeticsRepository cosmetics.CosmeticsRepository, bucketRepo bucket.BucketRepository) CosmeticsService {
	return &cosmeticsService{
		conf:                conf,
		cosmeticsRepository: cosmeticsRepository,
		bucketRepo:          bucketRepo,
	}
}

func (ss *cosmeticsService) GetCategoriesList() ([]*models.CosmeticCategory, error) {
	return ss.cosmeticsRepository.GetCategoriesList()
}

func (ss *cosmeticsService) GetCosmeticsListByCategory(category uuid.UUID, worldId uuid.UUID, playerId uuid.UUID, offset int, limit int) ([]*models.Cosmetic, int64, error) {
	_, err := ss.cosmeticsRepository.GetCategoryById(category)
	if err != nil {
		return nil, 0, assets_errors.NewCategoryNotFound("category not found")
	}
	return ss.cosmeticsRepository.GetCosmeticsListByCategory(category, worldId, playerId, offset, limit)
}

func (ss *cosmeticsService) GetCosmeticById(cosmeticId uuid.UUID) (*models.Cosmetic, error) {
	cosmetic, err := ss.cosmeticsRepository.GetCosmeticById(cosmeticId)
	if err != nil {
		return nil, err
	}
	return cosmetic, nil
}

func (ss *cosmeticsService) GetCosmeticsListByWorld(worldId uuid.UUID, offset int, limit int) ([]*models.Cosmetic, int64, error) {
	return ss.cosmeticsRepository.GetCosmeticsListByWorld(worldId, offset, limit)
}

func (ss *cosmeticsService) AddCategory(category string) (*models.CosmeticCategory, error) {
	return ss.cosmeticsRepository.AddCategory(category)
}

func (ss *cosmeticsService) UploadCosmeticData(categoryId uuid.UUID, worldId uuid.UUID, price float64, cosmeticData multipart.File, ext string, userId uuid.UUID) (*models.Cosmetic, error) {
	cosmeticUniqueUrl := uuid.New().String()

	category, err := ss.cosmeticsRepository.GetCategoryById(categoryId)
	if err != nil {
		logger.Logger.Errorf("Error getting category by id: %v", err)
		return nil, err
	}

	filePath := fmt.Sprintf("%s/%s%s", category.Name, cosmeticUniqueUrl, ext)
	if err := ss.bucketRepo.UploadFile(filePath, "image/png", cosmeticData); err != nil {
		logger.Logger.Errorf("Error uploading file to bucket: %v", err)
		return nil, err
	}

	cosmetic := &models.Cosmetic{
		Url: fmt.Sprintf("/%s", filePath),
	}
	if err := ss.cosmeticsRepository.CreateCosmetic(categoryId, worldId, price, cosmetic, userId); err != nil {
		logger.Logger.Errorf("Error creating cosmetic: %v", err)
		return nil, err
	}

	return cosmetic, nil
}

func (ss *cosmeticsService) UploadCosmeticByID(categoryId uuid.UUID, worldId uuid.UUID, price float64, spriteId uuid.UUID, userId uuid.UUID) (*models.Cosmetic, error) {
	if _, err := ss.cosmeticsRepository.GetCategoryById(categoryId); err != nil {
		logger.Logger.Errorf("Error getting category by id: %v", err)
		return nil, err
	}

	sourceCosmetic, err := ss.cosmeticsRepository.GetCosmeticById(spriteId)
	if err != nil {
		logger.Logger.Errorf("Error getting source cosmetic by id: %v", err)
		return nil, err
	}

	cosmetic := &models.Cosmetic{
		Url: sourceCosmetic.Url,
	}
	if err := ss.cosmeticsRepository.CreateCosmetic(categoryId, worldId, price, cosmetic, userId); err != nil {
		logger.Logger.Errorf("Error creating linked cosmetic: %v", err)
		return nil, err
	}

	return cosmetic, nil
}

func (ss *cosmeticsService) DeleteCosmetic(cosmeticId uuid.UUID) error {
	cosmetic, err := ss.cosmeticsRepository.GetCosmeticById(cosmeticId)
	if err != nil {
		logger.Logger.Errorf("Error getting cosmetic by id: %v", err)
		return err
	}

	if err := ss.bucketRepo.DeleteFile(cosmetic.Url); err != nil {
		logger.Logger.Errorf("Error deleting file from bucket: %v", err)
		return err
	}

	if err := ss.cosmeticsRepository.DeleteCosmetic(cosmeticId); err != nil {
		logger.Logger.Errorf("Error deleting cosmetic: %v", err)
		return err
	}

	return nil
}

func (ss *cosmeticsService) PurchaseCosmeticForUserInternal(userId uuid.UUID, cosmeticId uuid.UUID) error {
	if err := ss.cosmeticsRepository.AddPurchaseForUserId(cosmeticId, userId); err != nil {
		logger.Logger.Errorf("Error adding purchase for user: %v", err)
		return err
	}

	return nil
}
