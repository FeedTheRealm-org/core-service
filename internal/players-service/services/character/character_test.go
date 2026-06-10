package character

import (
	"errors"
	"os"
	"testing"

	"github.com/FeedTheRealm-org/core-service/config"
	"github.com/FeedTheRealm-org/core-service/internal/players-service/models"
	"github.com/FeedTheRealm-org/core-service/internal/utils/logger"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	logger.InitLogger(false)
	os.Exit(m.Run())
}

type fakeCharacterRepo struct {
	updateInfoCalled     bool
	deleteSpritesCalled  bool
	updateSpritesCalled  bool
	getInfoCalled        bool
	getSpritesCalled     bool
	updatedInfo          *models.CharacterInfo
	updateInfoErr        error
	deleteSpritesErr     error
	updateSpritesErr     error
	getInfoErr           error
	getSpritesErr        error
	returnInfo           *models.CharacterInfo
	returnSprites        []models.CategorySprite
	lastDeleteUserID     uuid.UUID
	lastDeleteCategoryID []uuid.UUID
}

func (f *fakeCharacterRepo) UpdateCharacterInfo(newCharacterInfo *models.CharacterInfo) error {
	f.updateInfoCalled = true
	f.updatedInfo = newCharacterInfo
	return f.updateInfoErr
}

func (f *fakeCharacterRepo) GetCharacterInfo(userId uuid.UUID) (*models.CharacterInfo, error) {
	f.getInfoCalled = true
	if f.getInfoErr != nil {
		return nil, f.getInfoErr
	}
	return f.returnInfo, nil
}

func (f *fakeCharacterRepo) UpdateCategorySprites(newCategorySprites []models.CategorySprite) error {
	f.updateSpritesCalled = true
	return f.updateSpritesErr
}

func (f *fakeCharacterRepo) DeleteCategorySprites(userId uuid.UUID, categoryIds []uuid.UUID) error {
	f.deleteSpritesCalled = true
	f.lastDeleteUserID = userId
	f.lastDeleteCategoryID = categoryIds
	return f.deleteSpritesErr
}

func (f *fakeCharacterRepo) GetCategorySprites(userId uuid.UUID) ([]models.CategorySprite, error) {
	f.getSpritesCalled = true
	if f.getSpritesErr != nil {
		return nil, f.getSpritesErr
	}
	return f.returnSprites, nil
}

func TestCharacterService_UpdateCharacterInfo_CallsRepositories(t *testing.T) {
	repo := &fakeCharacterRepo{}
	svc := NewCharacterService(config.CreateConfig(), repo)

	userID := uuid.New()
	newInfo := &models.CharacterInfo{CharacterName: "Hero", CharacterBio: "bio"}
	newSprites := []models.CategorySprite{{UserId: userID, CategoryId: uuid.New(), SpriteId: uuid.New()}}
	deleteIds := []uuid.UUID{uuid.New()}

	err := svc.UpdateCharacterInfo(userID, newInfo, newSprites, deleteIds)
	assert.NoError(t, err)
	assert.True(t, repo.updateInfoCalled)
	assert.True(t, repo.deleteSpritesCalled)
	assert.True(t, repo.updateSpritesCalled)
	assert.Equal(t, userID, repo.updatedInfo.UserId)
}

func TestCharacterService_UpdateCharacterInfo_SkipsEmptySlices(t *testing.T) {
	repo := &fakeCharacterRepo{}
	svc := NewCharacterService(config.CreateConfig(), repo)

	userID := uuid.New()
	newInfo := &models.CharacterInfo{CharacterName: "Hero", CharacterBio: "bio"}

	err := svc.UpdateCharacterInfo(userID, newInfo, nil, nil)
	assert.NoError(t, err)
	assert.True(t, repo.updateInfoCalled)
	assert.False(t, repo.deleteSpritesCalled)
	assert.False(t, repo.updateSpritesCalled)
}

func TestCharacterService_GetCharacterInfo_ReturnsData(t *testing.T) {
	userID := uuid.New()
	info := &models.CharacterInfo{UserId: userID, CharacterName: "Hero"}
	sprites := []models.CategorySprite{{UserId: userID, CategoryId: uuid.New(), SpriteId: uuid.New()}}

	repo := &fakeCharacterRepo{returnInfo: info, returnSprites: sprites}
	svc := NewCharacterService(config.CreateConfig(), repo)

	gotInfo, gotSprites, err := svc.GetCharacterInfo(userID)
	assert.NoError(t, err)
	assert.True(t, repo.getInfoCalled)
	assert.True(t, repo.getSpritesCalled)
	assert.Equal(t, info, gotInfo)
	assert.Equal(t, sprites, gotSprites)
}

func TestCharacterService_GetCharacterInfo_PropagatesError(t *testing.T) {
	repo := &fakeCharacterRepo{getInfoErr: errors.New("boom")}
	svc := NewCharacterService(config.CreateConfig(), repo)

	_, _, err := svc.GetCharacterInfo(uuid.New())
	assert.Error(t, err)
}
