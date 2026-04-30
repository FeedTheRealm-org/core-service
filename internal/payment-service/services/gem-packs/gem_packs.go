package gem_packs

import (
	"github.com/FeedTheRealm-org/core-service/config"
	"github.com/FeedTheRealm-org/core-service/internal/payment-service/models"
	gem_packs "github.com/FeedTheRealm-org/core-service/internal/payment-service/repositories/gem-packs"
	"github.com/FeedTheRealm-org/core-service/internal/utils/logger"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type gemPacksService struct {
	conf *config.Config
	repo gem_packs.GemPacksRepository
}

func (s *gemPacksService) seedPacksData() error {
	packs, err := s.GetAllGemPacks()
	if err != nil {
		return err
	}

	if len(s.conf.Stripe.GemPacks) == 0 {
		logger.Logger.Warn("No gem packs defined in configuration, skipping seeding")
		return nil
	}

	for _, pack := range packs {
		if err := s.DeleteGemPack(pack.Id); err != nil {
			return err
		}
	}

	for _, data := range s.conf.Stripe.GemPacks {
		_, err := s.CreateGemPack(data.Name, data.Amount, decimal.NewFromFloat(data.Price))
		if err != nil {
			return err
		}
	}

	return nil
}

func NewGemPacksService(conf *config.Config, repo gem_packs.GemPacksRepository) GemPacksService {
	newService := &gemPacksService{
		conf: conf,
		repo: repo,
	}

	if err := newService.seedPacksData(); err != nil {
		logger.Logger.Errorf("Failed to seed gem packs data: %v", err)
		return newService
	}

	return newService
}

func (s *gemPacksService) GetAllGemPacks() ([]*models.GemPack, error) {
	packs, err := s.repo.GetAllGemPacks()
	if err != nil {
		return nil, err
	}
	return packs, nil
}

func (s *gemPacksService) GetGemPackById(packId uuid.UUID) (*models.GemPack, error) {
	pack, err := s.repo.GetGemPackById(packId)
	if err != nil {
		return nil, err
	}
	return pack, nil
}

func (s *gemPacksService) CreateGemPack(name string, gems int, price decimal.Decimal) (*models.GemPack, error) {
	newPackage := &models.GemPack{
		Name:  name,
		Gems:  gems,
		Price: price,
	}

	createdPackage, err := s.repo.CreateGemPack(newPackage)
	if err != nil {
		return nil, err
	}
	return createdPackage, nil
}

func (s *gemPacksService) UpdateGemPack(packId uuid.UUID, name string, gems int, price decimal.Decimal) (*models.GemPack, error) {
	pack, err := s.repo.GetGemPackById(packId)
	if err != nil {
		return nil, err
	}

	if name != "" {
		pack.Name = name
	}

	if gems != 0 {
		pack.Gems = gems
	}

	if price.IsPositive() {
		pack.Price = price
	}

	err = s.repo.UpdateGemPack(packId, pack)
	if err != nil {
		return nil, err
	}
	return pack, nil
}

func (s *gemPacksService) DeleteGemPack(packId uuid.UUID) error {
	err := s.repo.DeleteGemPack(packId)
	if err != nil {
		return err
	}
	return nil
}
