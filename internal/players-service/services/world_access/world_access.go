package world_access

import (
	"strings"
	"time"

	"github.com/FeedTheRealm-org/core-service/config"
	player_errors "github.com/FeedTheRealm-org/core-service/internal/players-service/errors"
	"github.com/FeedTheRealm-org/core-service/internal/players-service/models"
	"github.com/FeedTheRealm-org/core-service/internal/players-service/repositories/character"
	"github.com/FeedTheRealm-org/core-service/internal/players-service/repositories/world_access"
	"github.com/google/uuid"
)

const worldJoinTokenTTL = 5 * time.Minute

type worldAccessService struct {
	conf                  *config.Config
	worldAccessRepository world_access.WorldAccessRepository
	characterRepository   character.CharacterRepository
}

func NewWorldAccessService(
	conf *config.Config,
	worldAccessRepository world_access.WorldAccessRepository,
	characterRepository character.CharacterRepository,
) WorldAccessService {
	return &worldAccessService{
		conf:                  conf,
		worldAccessRepository: worldAccessRepository,
		characterRepository:   characterRepository,
	}
}

func (ws *worldAccessService) IssueWorldJoinToken(userId uuid.UUID, worldId string) (*models.WorldJoinToken, error) {
	_, err := ws.characterRepository.GetCharacterInfo(userId)
	if err != nil {
		return nil, err
	}

	trimmedWorldId := strings.TrimSpace(worldId)
	if trimmedWorldId == "" {
		return nil, player_errors.NewWorldJoinTokenInvalid("world_id is required")
	}

	token := &models.WorldJoinToken{
		TokenId:   uuid.New(),
		UserId:    userId,
		WorldId:   trimmedWorldId,
		ExpiresAt: time.Now().UTC().Add(worldJoinTokenTTL),
	}

	if err := ws.worldAccessRepository.CreateWorldJoinToken(token); err != nil {
		return nil, err
	}
	return token, nil
}

func (ws *worldAccessService) ConsumeWorldJoinToken(tokenId string) (*models.WorldJoinToken, error) {
	tokenUUID, err := uuid.Parse(strings.TrimSpace(tokenId))
	if err != nil {
		return nil, player_errors.NewWorldJoinTokenInvalid("invalid token_id")
	}

	return ws.worldAccessRepository.ConsumeWorldJoinToken(tokenUUID, time.Now().UTC())
}
