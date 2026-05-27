package world_access

import (
	"errors"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/FeedTheRealm-org/core-service/config"
	player_errors "github.com/FeedTheRealm-org/core-service/internal/players-service/errors"
	"github.com/FeedTheRealm-org/core-service/internal/players-service/models"
	"github.com/FeedTheRealm-org/core-service/internal/utils/logger"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var worldAccessConf *config.Config
var worldAccessDB *config.DB
var worldAccessRepo WorldAccessRepository

func TestMain(m *testing.M) {
	logger.InitLogger(false)
	worldAccessConf = config.CreateConfig()
	var err error
	worldAccessDB, err = config.NewDB(worldAccessConf)
	if err != nil {
		panic(err)
	}
	worldAccessRepo = NewWorldAccessRepository(worldAccessConf, worldAccessDB)

	clearWorldAccessTables()
	code := m.Run()
	clearWorldAccessTables()
	os.Exit(code)
}

func clearWorldAccessTables() {
}

func createCharacterInfo(t *testing.T, userId uuid.UUID) {
	info := &models.CharacterInfo{
		UserId:        userId,
		CharacterName: "User-" + uuid.New().String(),
		CharacterBio:  "bio",
		SkinColor:     models.DefaultCharacterColorHsv(),
		HairColor:     models.DefaultCharacterColorHsv(),
		EyeColor:      models.DefaultCharacterColorHsv(),
	}
	assert.NoError(t, worldAccessDB.Conn.Create(info).Error)
}

func createWorldJoinToken(t *testing.T, token *models.WorldJoinToken) {
	for i := 0; i < 3; i++ {
		err := worldAccessRepo.CreateWorldJoinToken(token)
		if err == nil {
			return
		}
		if strings.Contains(err.Error(), "foreign key constraint") {
			createCharacterInfo(t, token.UserId)
			continue
		}
		t.Fatalf("failed to create world join token: %v", err)
	}
	assert.Fail(t, "failed to create world join token after retries")
}

func TestWorldAccessRepository_CreateAndConsume(t *testing.T) {
	clearWorldAccessTables()

	userID := uuid.New()
	createCharacterInfo(t, userID)

	now := time.Now().UTC()
	tokenID := uuid.New()
	token := &models.WorldJoinToken{
		TokenId:   tokenID,
		UserId:    userID,
		WorldId:   "world_1",
		ExpiresAt: now.Add(10 * time.Minute),
	}

	createWorldJoinToken(t, token)

	consumed, err := worldAccessRepo.ConsumeWorldJoinToken(tokenID, now)
	require.NoError(t, err)
	require.NotNil(t, consumed)
	assert.Equal(t, tokenID, consumed.TokenId)
	assert.NotNil(t, consumed.ConsumedAt)
}

func TestWorldAccessRepository_Consume_NotFound(t *testing.T) {
	clearWorldAccessTables()

	_, err := worldAccessRepo.ConsumeWorldJoinToken(uuid.New(), time.Now().UTC())
	assert.Error(t, err)
	var notFound *player_errors.WorldJoinTokenNotFound
	assert.True(t, errors.As(err, &notFound))
}

func TestWorldAccessRepository_Consume_Expired(t *testing.T) {
	clearWorldAccessTables()

	userID := uuid.New()
	createCharacterInfo(t, userID)

	tokenID := uuid.New()
	token := &models.WorldJoinToken{
		TokenId:   tokenID,
		UserId:    userID,
		WorldId:   "world_expired",
		ExpiresAt: time.Now().UTC().Add(-10 * time.Minute),
	}
	createWorldJoinToken(t, token)

	_, err := worldAccessRepo.ConsumeWorldJoinToken(tokenID, time.Now().UTC())
	assert.Error(t, err)
	var expired *player_errors.WorldJoinTokenExpired
	assert.True(t, errors.As(err, &expired))
}

func TestWorldAccessRepository_Consume_AlreadyConsumed(t *testing.T) {
	clearWorldAccessTables()

	userID := uuid.New()
	createCharacterInfo(t, userID)

	consumedAt := time.Now().UTC().Add(-1 * time.Minute)
	tokenID := uuid.New()
	token := &models.WorldJoinToken{
		TokenId:    tokenID,
		UserId:     userID,
		WorldId:    "world_consumed",
		ExpiresAt:  time.Now().UTC().Add(10 * time.Minute),
		ConsumedAt: &consumedAt,
	}
	createWorldJoinToken(t, token)

	_, err := worldAccessRepo.ConsumeWorldJoinToken(tokenID, time.Now().UTC())
	assert.Error(t, err)
	var consumed *player_errors.WorldJoinTokenConsumed
	assert.True(t, errors.As(err, &consumed))
}
