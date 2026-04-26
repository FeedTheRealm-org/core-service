package cosmetics

import (
	"github.com/FeedTheRealm-org/core-service/config"
	assets_errors "github.com/FeedTheRealm-org/core-service/internal/assets-service/errors"
	"github.com/FeedTheRealm-org/core-service/internal/assets-service/models"
	"github.com/FeedTheRealm-org/core-service/internal/errors"
	"github.com/google/uuid"
)

type cosmeticsRepository struct {
	conf *config.Config
	db   *config.DB
}

// NewCosmeticsRepository creates a new instance of CosmeticsRepository.
func NewCosmeticsRepository(conf *config.Config, db *config.DB) CosmeticsRepository {
	return &cosmeticsRepository{
		conf: conf,
		db:   db,
	}
}

func (cr *cosmeticsRepository) GetCategoriesList() ([]*models.CosmeticCategory, error) {
	var categories []*models.CosmeticCategory
	if err := cr.db.Conn.Find(&categories).Error; err != nil {
		return nil, err
	}
	return categories, nil
}

func (cr *cosmeticsRepository) GetCosmeticsListByCategory(category uuid.UUID, worldId *uuid.UUID, playerId *uuid.UUID, offset int, limit int) ([]*models.Cosmetic, int64, error) {
	query := cr.db.Conn.Model(&models.Cosmetic{}).
		Where("cosmetics.category_id = ?", category)

	switch {
	case worldId == nil && playerId == nil:
		query = query.Where("cosmetics.world_id = ?", uuid.Nil)
	case worldId != nil && playerId == nil:
		query = query.Where("cosmetics.world_id = ? OR cosmetics.world_id = ?", *worldId, uuid.Nil)
	case worldId == nil && playerId != nil:
		query = query.
			Joins("LEFT JOIN purchases ON purchases.cosmetic_id = cosmetics.id AND purchases.player_id = ?", *playerId).
			Where("cosmetics.world_id = ? OR purchases.cosmetic_id IS NOT NULL", uuid.Nil)
	case worldId != nil && playerId != nil:
		query = query.
			Joins("LEFT JOIN purchases ON purchases.cosmetic_id = cosmetics.id AND purchases.player_id = ?", *playerId).
			Where(
				"cosmetics.world_id = ? OR (cosmetics.world_id = ? AND purchases.cosmetic_id IS NOT NULL)",
				uuid.Nil, *worldId,
			)
	}

	var totalCount int64
	if err := query.Count(&totalCount).Error; err != nil {
		return nil, 0, err
	}

	if totalCount == 0 {
		return []*models.Cosmetic{}, 0, nil
	}

	var cosmetics []*models.Cosmetic
	if err := query.
		Order("cosmetics.world_id ASC, cosmetics.id ASC").
		Offset(offset).
		Limit(limit).
		Find(&cosmetics).Error; err != nil {
		return nil, 0, err
	}

	return cosmetics, totalCount, nil
}

func (cr *cosmeticsRepository) GetCosmeticById(cosmeticId uuid.UUID) (*models.Cosmetic, error) {
	var cosmetic models.Cosmetic
	if err := cr.db.Conn.First(&cosmetic, "id = ?", cosmeticId).Error; err != nil {
		if errors.IsRecordNotFound(err) {
			return nil, assets_errors.NewCosmeticNotFound("cosmetic not found")
		}
		return nil, err
	}
	return &cosmetic, nil
}

func (cr *cosmeticsRepository) AddCategory(categoryName string) (*models.CosmeticCategory, error) {
	category := &models.CosmeticCategory{
		Name: categoryName,
	}

	if err := cr.db.Conn.Create(category).Error; err != nil {
		if errors.IsDuplicateEntryError(err) {
			return nil, assets_errors.NewCategoryConflict(err.Error())
		}
		return nil, err
	}

	return category, nil
}

func (cr *cosmeticsRepository) AddPurchaseForUserId(cosmeticId uuid.UUID, userId uuid.UUID) error {
	purchase := &models.Purchase{
		CosmeticID: cosmeticId,
		PlayerID:   userId,
	}

	if err := cr.db.Conn.Where("cosmetic_id = ? AND player_id = ?", cosmeticId, userId).First(&models.Purchase{}).Error; err == nil {
		return assets_errors.NewCosmeticsWasPurchasedBefore("cosmetic was already purchased by the user")
	} else if !errors.IsRecordNotFound(err) {
		return err
	}

	if err := cr.db.Conn.Create(purchase).Error; err != nil {
		return err
	}

	return nil
}

func (cr *cosmeticsRepository) GetCategoryById(categoryId uuid.UUID) (*models.CosmeticCategory, error) {
	var category models.CosmeticCategory
	if err := cr.db.Conn.First(&category, "id = ?", categoryId).Error; err != nil {
		if errors.IsRecordNotFound(err) {
			return nil, assets_errors.NewCategoryNotFound("category not found")
		}
		return nil, err
	}
	return &category, nil
}

func (cr *cosmeticsRepository) GetCosmeticsListByWorld(worldId uuid.UUID, offset int, limit int) ([]*models.Cosmetic, int64, error) {
	var totalCount int64
	if err := cr.db.Conn.Model(&models.Cosmetic{}).
		Where("world_id = ?", worldId).
		Count(&totalCount).Error; err != nil {
		return nil, 0, err
	}

	if totalCount == 0 {
		return []*models.Cosmetic{}, 0, nil
	}

	var cosmetics []*models.Cosmetic
	if err := cr.db.Conn.Where("world_id = ?", worldId).
		Order("id ASC").
		Offset(offset).
		Limit(limit).
		Find(&cosmetics).Error; err != nil {
		return nil, 0, err
	}

	return cosmetics, totalCount, nil
}

func (cr *cosmeticsRepository) CreateCosmetic(categoryId uuid.UUID, worldId uuid.UUID, price float64, cosmetic *models.Cosmetic, userId uuid.UUID) error {
	var category models.CosmeticCategory
	if err := cr.db.Conn.First(&category, "id = ?", categoryId).Error; err != nil {
		if errors.IsRecordNotFound(err) {
			return assets_errors.NewCategoryNotFound("category not found")
		}
		return err
	}

	cosmetic.CategoryID = category.Id
	cosmetic.WorldID = worldId
	cosmetic.CreatedBy = userId
	cosmetic.Price = price

	if err := cr.db.Conn.Create(cosmetic).Error; err != nil {
		return err
	}

	return nil
}

func (cr *cosmeticsRepository) DeleteCosmetic(cosmeticId uuid.UUID) error {
	if err := cr.db.Conn.Model(&models.Cosmetic{}).Where("id = ?", cosmeticId).Update("world_id", nil).Error; err != nil {
		return err
	}
	return nil
}
