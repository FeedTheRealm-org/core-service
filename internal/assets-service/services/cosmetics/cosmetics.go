package cosmetics

import (
	"fmt"
	"mime/multipart"

	"github.com/FeedTheRealm-org/core-service/config"
	"github.com/FeedTheRealm-org/core-service/internal/assets-service/models"
	"github.com/FeedTheRealm-org/core-service/internal/assets-service/repositories/bucket"
	"github.com/FeedTheRealm-org/core-service/internal/assets-service/repositories/cosmetics"
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

func (ss *cosmeticsService) GetCosmeticsListByCategory(category uuid.UUID) ([]*models.Cosmetic, error) {
	return ss.cosmeticsRepository.GetCosmeticsListByCategory(category)
}

func (ss *cosmeticsService) GetCosmeticById(cosmeticId uuid.UUID) (*models.Cosmetic, error) {
	cosmetic, err := ss.cosmeticsRepository.GetCosmeticById(cosmeticId)
	if err != nil {
		return nil, err
	}
	return cosmetic, nil
}

func (ss *cosmeticsService) AddCategory(category string) (*models.CosmeticCategory, error) {
	return ss.cosmeticsRepository.AddCategory(category)
}

func (ss *cosmeticsService) UploadCosmeticData(categoryId uuid.UUID, cosmeticData multipart.File, ext string) (*models.Cosmetic, error) {
	cosmeticUniqueUrl := uuid.New().String()

	category, err := ss.cosmeticsRepository.GetCategoryById(categoryId)
	if err != nil {
		return nil, err
	}

	filePath := fmt.Sprintf("/%s/%s%s", category.Name, cosmeticUniqueUrl, ext)
	if err := ss.bucketRepo.UploadFile(filePath, "image/png", cosmeticData); err != nil {
		return nil, err
	}

	cosmetic := &models.Cosmetic{
		Url: filePath,
	}
	if err := ss.cosmeticsRepository.CreateCosmetic(categoryId, cosmetic); err != nil {
		return nil, err
	}

	return cosmetic, nil
}
