package gem_packs

import (
	"github.com/FeedTheRealm-org/core-service/config"
	"github.com/FeedTheRealm-org/core-service/internal/payment-service/models"
	gem_packs "github.com/FeedTheRealm-org/core-service/internal/payment-service/repositories/gem-packs"
	"github.com/google/uuid"
)

type gemGemPacksService struct {
	conf *config.Config
	repo gem_packs.GemPacksRepository
}

func (s *gemGemPacksService) seedPacksData() error {
	packs, err := s.GetAllGemPacks()
	if err != nil {
		return err
	}

	if len(packs) > 0 {
		return nil
	}

	newPacks := []struct {
		Name  string
		Gems  int
		Price float32
	}{
		{"Small Pack", 1, 1.99},
		{"Medium Pack", 10, 14.99},
		{"Large Pack", 50, 24.99},
	}

	for _, data := range newPacks {
		_, err := s.CreateGemPack(data.Name, data.Gems, data.Price)
		if err != nil {
			return err
		}
	}

	return nil
}

func NewGemPacksService(conf *config.Config, repo gem_packs.GemPacksRepository) GemPacksService {
	newService := &gemGemPacksService{
		conf: conf,
		repo: repo,
	}

	if err := newService.seedPacksData(); err != nil {
		return nil
	}

	return newService
}

func (s *gemGemPacksService) GetAllGemPacks() ([]*models.GemPack, error) {
	packs, err := s.repo.GetAllGemPacks()
	if err != nil {
		return nil, err
	}
	return packs, nil
}

func (s *gemGemPacksService) GetGemPackById(packId uuid.UUID) (*models.GemPack, error) {
	pack, err := s.repo.GetGemPackById(packId)
	if err != nil {
		return nil, err
	}
	return pack, nil
}

func (s *gemGemPacksService) CreateGemPack(name string, gems int, price float32) (*models.GemPack, error) {
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

func (s *gemGemPacksService) UpdateGemPack(packId uuid.UUID, name string, gems int, price float32) (*models.GemPack, error) {
	pack, err := s.repo.GetGemPackById(packId)
	if err != nil {
		return nil, err
	}

	pack.Name = name
	pack.Gems = gems
	pack.Price = price

	updatedPackage, err := s.repo.UpdateGemPack(packId, pack)
	if err != nil {
		return nil, err
	}
	return updatedPackage, nil
}

func (s *gemGemPacksService) DeleteGemPack(packId uuid.UUID) error {
	err := s.repo.DeleteGemPack(packId)
	if err != nil {
		return err
	}
	return nil
}
