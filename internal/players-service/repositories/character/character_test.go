package character

import (
	"errors"
	"os"
	"testing"

	"github.com/FeedTheRealm-org/core-service/config"
	player_errors "github.com/FeedTheRealm-org/core-service/internal/players-service/errors"
	"github.com/FeedTheRealm-org/core-service/internal/players-service/models"
	"github.com/FeedTheRealm-org/core-service/internal/utils/logger"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

var characterConf *config.Config
var characterDB *config.DB
var characterRepo CharacterRepository

func TestMain(m *testing.M) {
	logger.InitLogger(false)
	characterConf = config.CreateConfig()
	var err error
	characterDB, err = config.NewDB(characterConf)
	if err != nil {
		panic(err)
	}
	characterRepo = NewCharacterRepository(characterConf, characterDB)

	clearCharacterTables()
	code := m.Run()
	clearCharacterTables()
	os.Exit(code)
}

func clearCharacterTables() {
}

func createCharacterInfo(t *testing.T, userId uuid.UUID, name string) *models.CharacterInfo {
	info := &models.CharacterInfo{
		UserId:        userId,
		CharacterName: name,
		CharacterBio:  "bio",
		SkinColor:     models.DefaultCharacterColorHsv(),
		HairColor:     models.DefaultCharacterColorHsv(),
		EyeColor:      models.DefaultCharacterColorHsv(),
	}
	assert.NoError(t, characterRepo.UpdateCharacterInfo(info))
	return info
}

func TestCharacterRepository_UpdateAndGetCharacterInfo(t *testing.T) {
	clearCharacterTables()

	userID := uuid.New()
	name := "Hero-" + uuid.NewString()
	createCharacterInfo(t, userID, name)

	stored, err := characterRepo.GetCharacterInfo(userID)
	assert.NoError(t, err)
	assert.Equal(t, userID, stored.UserId)
	assert.Equal(t, name, stored.CharacterName)
}

func TestCharacterRepository_UpdateCharacterInfo_DuplicateName(t *testing.T) {
	clearCharacterTables()

	name := "Hero-" + uuid.NewString()
	createCharacterInfo(t, uuid.New(), name)

	info := &models.CharacterInfo{
		UserId:        uuid.New(),
		CharacterName: name,
		CharacterBio:  "bio",
		SkinColor:     models.DefaultCharacterColorHsv(),
		HairColor:     models.DefaultCharacterColorHsv(),
		EyeColor:      models.DefaultCharacterColorHsv(),
	}
	err := characterRepo.UpdateCharacterInfo(info)
	assert.Error(t, err)
	var nameTaken *player_errors.CharacterNameTaken
	assert.True(t, errors.As(err, &nameTaken))
}

func TestCharacterRepository_GetCharacterInfo_NotFound(t *testing.T) {
	clearCharacterTables()

	_, err := characterRepo.GetCharacterInfo(uuid.New())
	assert.Error(t, err)
	var notFound *player_errors.CharacterInfoNotFound
	assert.True(t, errors.As(err, &notFound))
}

func TestCharacterRepository_CategorySprites(t *testing.T) {
	clearCharacterTables()

	userID := uuid.New()
	createCharacterInfo(t, userID, "SpriteUser-"+uuid.NewString())

	categoryID1 := uuid.New()
	categoryID2 := uuid.New()
	sprites := []models.CategorySprite{
		{UserId: userID, CategoryId: categoryID1, SpriteId: uuid.New()},
		{UserId: userID, CategoryId: categoryID2, SpriteId: uuid.New()},
	}

	err := characterRepo.UpdateCategorySprites(sprites)
	assert.NoError(t, err)

	stored, err := characterRepo.GetCategorySprites(userID)
	assert.NoError(t, err)
	assert.Len(t, stored, 2)

	err = characterRepo.DeleteCategorySprites(userID, []uuid.UUID{categoryID1})
	assert.NoError(t, err)

	stored, err = characterRepo.GetCategorySprites(userID)
	assert.NoError(t, err)
	assert.Len(t, stored, 1)
}

func TestCharacterRepository_GetCategorySprites_NotFound(t *testing.T) {
	clearCharacterTables()

	userID := uuid.New()
	createCharacterInfo(t, userID, "NoSprites-"+uuid.NewString())

	_, err := characterRepo.GetCategorySprites(userID)
	assert.Error(t, err)
	var notFound *player_errors.CategorySpritesNotFound
	assert.True(t, errors.As(err, &notFound))
}

func TestCharacterRepository_DeleteCategorySprites_Empty(t *testing.T) {
	clearCharacterTables()

	userID := uuid.New()
	createCharacterInfo(t, userID, "NoDeletes-"+uuid.NewString())

	err := characterRepo.DeleteCategorySprites(userID, nil)
	assert.NoError(t, err)
}
