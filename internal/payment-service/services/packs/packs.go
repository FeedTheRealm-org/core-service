package packs

import (
	"github.com/FeedTheRealm-org/core-service/config"
	"github.com/FeedTheRealm-org/core-service/internal/payment-service/models"
	"github.com/FeedTheRealm-org/core-service/internal/payment-service/repositories/packs"
	"github.com/google/uuid"
)

type packsService struct {
	conf *config.Config
	repo packs.PacksRepository
}

func NewPacksService(conf *config.Config, repo packs.PacksRepository) PacksService {
	return &packsService{
		conf: conf,
		repo: repo,
	}
}

func (s *packsService) GetAllPacks() ([]*models.Pack, error) {
	packs, err := s.repo.GetAllPacks()
	if err != nil {
		return nil, err
	}
	return packs, nil
}

func (s *packsService) GetPackById(packId uuid.UUID) (*models.Pack, error) {
	pack, err := s.repo.GetPackById(packId)
	if err != nil {
		return nil, err
	}
	return pack, nil
}

func (s *packsService) CreatePack(name string, gems int, price float32) (*models.Pack, error) {
	newPackage := &models.Pack{
		Name:  name,
		Gems:  gems,
		Price: price,
	}

	createdPackage, err := s.repo.CreatePack(newPackage)
	if err != nil {
		return nil, err
	}
	return createdPackage, nil
}

func (s *packsService) UpdatePack(packId uuid.UUID, name string, gems int, price float32) (*models.Pack, error) {
	pack, err := s.repo.GetPackById(packId)
	if err != nil {
		return nil, err
	}

	pack.Name = name
	pack.Gems = gems
	pack.Price = price

	updatedPackage, err := s.repo.UpdatePack(packId, pack)
	if err != nil {
		return nil, err
	}
	return updatedPackage, nil
}

func (s *packsService) DeletePack(packId uuid.UUID) error {
	err := s.repo.DeletePack(packId)
	if err != nil {
		return err
	}
	return nil
}
