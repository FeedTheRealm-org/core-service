package gem_packs

import (
	"errors"
	"testing"

	"github.com/FeedTheRealm-org/core-service/config"
	"github.com/FeedTheRealm-org/core-service/internal/payment-service/models"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

type fakeGemPacksRepo struct {
	getAllErr   error
	getErr      error
	createErr   error
	updateErr   error
	deleteErr   error
	storedPack  *models.GemPack
	returnedAll []*models.GemPack
}

func (f *fakeGemPacksRepo) CreateGemPack(pkg *models.GemPack) (*models.GemPack, error) {
	if f.createErr != nil {
		return nil, f.createErr
	}
	pkg.Id = uuid.New()
	f.storedPack = pkg
	return pkg, nil
}

func (f *fakeGemPacksRepo) GetAllGemPacks() ([]*models.GemPack, error) {
	if f.getAllErr != nil {
		return nil, f.getAllErr
	}
	return f.returnedAll, nil
}

func (f *fakeGemPacksRepo) GetGemPackById(id uuid.UUID) (*models.GemPack, error) {
	if f.getErr != nil {
		return nil, f.getErr
	}
	if f.storedPack != nil {
		return f.storedPack, nil
	}
	return &models.GemPack{Id: id, Name: "Starter", Gems: 100, Price: decimal.NewFromFloat(1.5)}, nil
}

func (f *fakeGemPacksRepo) UpdateGemPack(id uuid.UUID, updatedPkg *models.GemPack) error {
	if f.updateErr != nil {
		return f.updateErr
	}
	f.storedPack = updatedPkg
	return nil
}

func (f *fakeGemPacksRepo) DeleteGemPack(id uuid.UUID) error {
	if f.deleteErr != nil {
		return f.deleteErr
	}
	return nil
}

func TestGemPacksService_GetAll_Error(t *testing.T) {
	conf := config.CreateConfig()
	repo := &fakeGemPacksRepo{getAllErr: errors.New("boom")}
	service := NewGemPacksService(conf, repo).(*gemPacksService)

	packs, err := service.GetAllGemPacks()
	assert.Error(t, err)
	assert.Nil(t, packs)
}

func TestGemPacksService_GetById_Error(t *testing.T) {
	conf := config.CreateConfig()
	repo := &fakeGemPacksRepo{getErr: errors.New("missing")}
	service := NewGemPacksService(conf, repo).(*gemPacksService)

	pack, err := service.GetGemPackById(uuid.New())
	assert.Error(t, err)
	assert.Nil(t, pack)
}

func TestGemPacksService_Create_Error(t *testing.T) {
	conf := config.CreateConfig()
	repo := &fakeGemPacksRepo{createErr: errors.New("boom")}
	service := NewGemPacksService(conf, repo).(*gemPacksService)

	pack, err := service.CreateGemPack("Starter", 100, decimal.NewFromFloat(1.5))
	assert.Error(t, err)
	assert.Nil(t, pack)
}

func TestGemPacksService_Update_GetError(t *testing.T) {
	conf := config.CreateConfig()
	repo := &fakeGemPacksRepo{getErr: errors.New("missing")}
	service := NewGemPacksService(conf, repo).(*gemPacksService)

	pack, err := service.UpdateGemPack(uuid.New(), "", 0, decimal.Zero)
	assert.Error(t, err)
	assert.Nil(t, pack)
}

func TestGemPacksService_Update_UpdateError(t *testing.T) {
	conf := config.CreateConfig()
	repo := &fakeGemPacksRepo{updateErr: errors.New("boom")}
	service := NewGemPacksService(conf, repo).(*gemPacksService)

	pack, err := service.UpdateGemPack(uuid.New(), "Pro", 200, decimal.NewFromFloat(3.0))
	assert.Error(t, err)
	assert.Nil(t, pack)
}

func TestGemPacksService_Delete_Error(t *testing.T) {
	conf := config.CreateConfig()
	repo := &fakeGemPacksRepo{deleteErr: errors.New("boom")}
	service := NewGemPacksService(conf, repo).(*gemPacksService)

	err := service.DeleteGemPack(uuid.New())
	assert.Error(t, err)
}
