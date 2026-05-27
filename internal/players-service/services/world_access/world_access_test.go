package world_access

import (
	"errors"
	"testing"
	"time"

	"github.com/FeedTheRealm-org/core-service/config"
	player_errors "github.com/FeedTheRealm-org/core-service/internal/players-service/errors"
	"github.com/FeedTheRealm-org/core-service/internal/players-service/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

type fakeWorldAccessRepo struct {
	createToken *models.WorldJoinToken
	createErr   error
	consumeID   uuid.UUID
	consumeTime time.Time
	consumeErr  error
	consumeResp *models.WorldJoinToken
}

func (f *fakeWorldAccessRepo) CreateWorldJoinToken(token *models.WorldJoinToken) error {
	f.createToken = token
	return f.createErr
}

func (f *fakeWorldAccessRepo) ConsumeWorldJoinToken(tokenId uuid.UUID, now time.Time) (*models.WorldJoinToken, error) {
	f.consumeID = tokenId
	f.consumeTime = now
	if f.consumeErr != nil {
		return nil, f.consumeErr
	}
	return f.consumeResp, nil
}

type fakeCharacterRepo struct {
	getInfoErr error
}

func (f *fakeCharacterRepo) UpdateCharacterInfo(newCharacterInfo *models.CharacterInfo) error {
	return nil
}

func (f *fakeCharacterRepo) GetCharacterInfo(userId uuid.UUID) (*models.CharacterInfo, error) {
	if f.getInfoErr != nil {
		return nil, f.getInfoErr
	}
	return &models.CharacterInfo{UserId: userId}, nil
}

func (f *fakeCharacterRepo) UpdateCategorySprites(newCategorySprites []models.CategorySprite) error {
	return nil
}

func (f *fakeCharacterRepo) DeleteCategorySprites(userId uuid.UUID, categoryIds []uuid.UUID) error {
	return nil
}

func (f *fakeCharacterRepo) GetCategorySprites(userId uuid.UUID) ([]models.CategorySprite, error) {
	return nil, nil
}

func TestWorldAccessService_IssueWorldJoinToken_InvalidWorldId(t *testing.T) {
	repo := &fakeWorldAccessRepo{}
	characterRepo := &fakeCharacterRepo{}
	svc := NewWorldAccessService(config.CreateConfig(), repo, characterRepo)

	_, err := svc.IssueWorldJoinToken(uuid.New(), "   ")
	assert.Error(t, err)
	var invalid *player_errors.WorldJoinTokenInvalid
	assert.True(t, errors.As(err, &invalid))
}

func TestWorldAccessService_IssueWorldJoinToken_CharacterError(t *testing.T) {
	repo := &fakeWorldAccessRepo{}
	characterRepo := &fakeCharacterRepo{getInfoErr: errors.New("missing")}
	svc := NewWorldAccessService(config.CreateConfig(), repo, characterRepo)

	_, err := svc.IssueWorldJoinToken(uuid.New(), "world")
	assert.Error(t, err)
}

func TestWorldAccessService_IssueWorldJoinToken_CreatesToken(t *testing.T) {
	repo := &fakeWorldAccessRepo{}
	characterRepo := &fakeCharacterRepo{}
	svc := NewWorldAccessService(config.CreateConfig(), repo, characterRepo)

	userID := uuid.New()
	token, err := svc.IssueWorldJoinToken(userID, "  world_1  ")
	assert.NoError(t, err)
	assert.NotNil(t, token)
	assert.Equal(t, userID, token.UserId)
	assert.Equal(t, "world_1", token.WorldId)
	assert.True(t, token.ExpiresAt.After(time.Now().UTC()))
	assert.Equal(t, token, repo.createToken)
}

func TestWorldAccessService_ConsumeWorldJoinToken_InvalidTokenId(t *testing.T) {
	repo := &fakeWorldAccessRepo{}
	characterRepo := &fakeCharacterRepo{}
	svc := NewWorldAccessService(config.CreateConfig(), repo, characterRepo)

	_, err := svc.ConsumeWorldJoinToken("not-a-uuid")
	assert.Error(t, err)
	var invalid *player_errors.WorldJoinTokenInvalid
	assert.True(t, errors.As(err, &invalid))
}

func TestWorldAccessService_ConsumeWorldJoinToken_CallsRepo(t *testing.T) {
	repo := &fakeWorldAccessRepo{consumeResp: &models.WorldJoinToken{TokenId: uuid.New()}}
	characterRepo := &fakeCharacterRepo{}
	svc := NewWorldAccessService(config.CreateConfig(), repo, characterRepo)

	id := uuid.New()
	token, err := svc.ConsumeWorldJoinToken(id.String())
	assert.NoError(t, err)
	assert.Equal(t, repo.consumeResp, token)
	assert.Equal(t, id, repo.consumeID)
}
