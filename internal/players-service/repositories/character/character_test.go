package character_test

import (
	"os"
	"testing"

	"github.com/FeedTheRealm-org/core-service/config"
	player_errors "github.com/FeedTheRealm-org/core-service/internal/players-service/errors"
	"github.com/FeedTheRealm-org/core-service/internal/players-service/models"
	"github.com/FeedTheRealm-org/core-service/internal/players-service/repositories/character"
	"github.com/FeedTheRealm-org/core-service/internal/utils/logger"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

var repo character.CharacterRepository

func TestMain(m *testing.M) {
	conf := config.CreateConfig()
	logger.InitLogger(false)
	db, err := config.NewDB(conf)
	if err != nil {
		panic(err)
	}
	repo = character.NewCharacterRepository(conf, db)

	db.Conn.Exec("TRUNCATE TABLE category_sprites, character_infos CASCADE;")
	code := m.Run()
	db.Conn.Exec("TRUNCATE TABLE category_sprites, character_infos CASCADE;")
	os.Exit(code)
}

// --- GetCharacterInfo ---

func TestCharacterRepository_GetCharacterInfo_NotFound(t *testing.T) {
	info, err := repo.GetCharacterInfo(uuid.New())

	assert.Error(t, err)
	assert.Nil(t, info)
	var notFound *player_errors.CharacterInfoNotFound
	assert.ErrorAs(t, err, &notFound)
}

func TestCharacterRepository_GetCharacterInfo_Success(t *testing.T) {
	userId := uuid.New()
	charInfo := &models.CharacterInfo{
		UserId:        userId,
		CharacterName: "Ranger_" + userId.String()[:8],
		CharacterBio:  "Lives in the woods",
		SkinColor:     models.DefaultCharacterColorHsv(),
		HairColor:     models.DefaultCharacterColorHsv(),
		EyeColor:      models.DefaultCharacterColorHsv(),
	}
	err := repo.UpdateCharacterInfo(charInfo)
	assert.NoError(t, err)

	got, err := repo.GetCharacterInfo(userId)
	assert.NoError(t, err)
	assert.NotNil(t, got)
	assert.Equal(t, userId, got.UserId)
	assert.Equal(t, charInfo.CharacterName, got.CharacterName)
	assert.Equal(t, charInfo.CharacterBio, got.CharacterBio)
}

// --- UpdateCharacterInfo ---

func TestCharacterRepository_UpdateCharacterInfo_Create(t *testing.T) {
	userId := uuid.New()
	charInfo := &models.CharacterInfo{
		UserId:        userId,
		CharacterName: "Bard_" + userId.String()[:8],
		CharacterBio:  "Plays music",
		SkinColor:     models.DefaultCharacterColorHsv(),
		HairColor:     models.DefaultCharacterColorHsv(),
		EyeColor:      models.DefaultCharacterColorHsv(),
	}

	err := repo.UpdateCharacterInfo(charInfo)
	assert.NoError(t, err)

	got, err := repo.GetCharacterInfo(userId)
	assert.NoError(t, err)
	assert.Equal(t, charInfo.CharacterName, got.CharacterName)
}

func TestCharacterRepository_UpdateCharacterInfo_Update(t *testing.T) {
	userId := uuid.New()
	charInfo := &models.CharacterInfo{
		UserId:        userId,
		CharacterName: "Sorcerer_" + userId.String()[:8],
		CharacterBio:  "Original",
		SkinColor:     models.DefaultCharacterColorHsv(),
		HairColor:     models.DefaultCharacterColorHsv(),
		EyeColor:      models.DefaultCharacterColorHsv(),
	}
	err := repo.UpdateCharacterInfo(charInfo)
	assert.NoError(t, err)

	charInfo.CharacterBio = "Updated"
	err = repo.UpdateCharacterInfo(charInfo)
	assert.NoError(t, err)

	got, err := repo.GetCharacterInfo(userId)
	assert.NoError(t, err)
	assert.Equal(t, "Updated", got.CharacterBio)
}

func TestCharacterRepository_UpdateCharacterInfo_DuplicateName(t *testing.T) {
	userId1 := uuid.New()
	userId2 := uuid.New()
	sharedName := "Monk_Shared_" + userId1.String()[:8]

	err := repo.UpdateCharacterInfo(&models.CharacterInfo{
		UserId:        userId1,
		CharacterName: sharedName,
		CharacterBio:  "First",
		SkinColor:     models.DefaultCharacterColorHsv(),
		HairColor:     models.DefaultCharacterColorHsv(),
		EyeColor:      models.DefaultCharacterColorHsv(),
	})
	assert.NoError(t, err)

	err = repo.UpdateCharacterInfo(&models.CharacterInfo{
		UserId:        userId2,
		CharacterName: sharedName,
		CharacterBio:  "Second",
		SkinColor:     models.DefaultCharacterColorHsv(),
		HairColor:     models.DefaultCharacterColorHsv(),
		EyeColor:      models.DefaultCharacterColorHsv(),
	})
	assert.Error(t, err)
	var nameTaken *player_errors.CharacterNameTaken
	assert.ErrorAs(t, err, &nameTaken)
}

// --- CategorySprites ---

func TestCharacterRepository_GetCategorySprites_NotFound(t *testing.T) {
	sprites, err := repo.GetCategorySprites(uuid.New())

	assert.Error(t, err)
	assert.Nil(t, sprites)
	var notFound *player_errors.CategorySpritesNotFound
	assert.ErrorAs(t, err, &notFound)
}

func TestCharacterRepository_UpdateAndGetCategorySprites(t *testing.T) {
	userId := uuid.New()
	categoryId := uuid.New()
	spriteId := uuid.New()

	sprites := []models.CategorySprite{
		{UserId: userId, CategoryId: categoryId, SpriteId: spriteId},
	}
	err := repo.UpdateCategorySprites(sprites)
	assert.NoError(t, err)

	got, err := repo.GetCategorySprites(userId)
	assert.NoError(t, err)
	assert.Len(t, got, 1)
	assert.Equal(t, categoryId, got[0].CategoryId)
	assert.Equal(t, spriteId, got[0].SpriteId)
}

func TestCharacterRepository_UpdateCategorySprites_Upsert(t *testing.T) {
	userId := uuid.New()
	categoryId := uuid.New()
	spriteId1 := uuid.New()
	spriteId2 := uuid.New()

	err := repo.UpdateCategorySprites([]models.CategorySprite{
		{UserId: userId, CategoryId: categoryId, SpriteId: spriteId1},
	})
	assert.NoError(t, err)

	// Upsert same category/user with new sprite
	err = repo.UpdateCategorySprites([]models.CategorySprite{
		{UserId: userId, CategoryId: categoryId, SpriteId: spriteId2},
	})
	assert.NoError(t, err)

	got, err := repo.GetCategorySprites(userId)
	assert.NoError(t, err)
	assert.Len(t, got, 1)
	assert.Equal(t, spriteId2, got[0].SpriteId, "sprite should be updated to spriteId2")
}

func TestCharacterRepository_UpdateCategorySprites_MultipleCategories(t *testing.T) {
	userId := uuid.New()
	catId1, catId2 := uuid.New(), uuid.New()
	spriteId1, spriteId2 := uuid.New(), uuid.New()

	err := repo.UpdateCategorySprites([]models.CategorySprite{
		{UserId: userId, CategoryId: catId1, SpriteId: spriteId1},
		{UserId: userId, CategoryId: catId2, SpriteId: spriteId2},
	})
	assert.NoError(t, err)

	got, err := repo.GetCategorySprites(userId)
	assert.NoError(t, err)
	assert.Len(t, got, 2)
}

func TestCharacterRepository_DeleteCategorySprites(t *testing.T) {
	userId := uuid.New()
	categoryId := uuid.New()

	err := repo.UpdateCategorySprites([]models.CategorySprite{
		{UserId: userId, CategoryId: categoryId, SpriteId: uuid.New()},
	})
	assert.NoError(t, err)

	err = repo.DeleteCategorySprites(userId, []uuid.UUID{categoryId})
	assert.NoError(t, err)

	got, err := repo.GetCategorySprites(userId)
	assert.Error(t, err)
	var notFound *player_errors.CategorySpritesNotFound
	assert.ErrorAs(t, err, &notFound)
	assert.Nil(t, got)
}

func TestCharacterRepository_DeleteCategorySprites_EmptyList(t *testing.T) {
	// Deleting with empty list is a no-op and should not error
	err := repo.DeleteCategorySprites(uuid.New(), []uuid.UUID{})
	assert.NoError(t, err)
}

func TestCharacterRepository_DeleteCategorySprites_NonExistent(t *testing.T) {
	// Deleting a category that doesn't exist should not error
	err := repo.DeleteCategorySprites(uuid.New(), []uuid.UUID{uuid.New()})
	assert.NoError(t, err)
}

func TestCharacterRepository_DeleteCategorySprites_OnlyDeletesTargetCategory(t *testing.T) {
	userId := uuid.New()
	catId1 := uuid.New()
	catId2 := uuid.New()

	err := repo.UpdateCategorySprites([]models.CategorySprite{
		{UserId: userId, CategoryId: catId1, SpriteId: uuid.New()},
		{UserId: userId, CategoryId: catId2, SpriteId: uuid.New()},
	})
	assert.NoError(t, err)

	// Delete only catId1
	err = repo.DeleteCategorySprites(userId, []uuid.UUID{catId1})
	assert.NoError(t, err)

	got, err := repo.GetCategorySprites(userId)
	assert.NoError(t, err)
	assert.Len(t, got, 1)
	assert.Equal(t, catId2, got[0].CategoryId)
}
