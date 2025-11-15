package character

import (
	"github.com/FeedTheRealm-org/core-service/config"
	"github.com/FeedTheRealm-org/core-service/internal/players-service/models"
	"github.com/FeedTheRealm-org/core-service/internal/players-service/repositories/character"
	"github.com/FeedTheRealm-org/core-service/internal/utils/logger"
	"github.com/google/uuid"
)

type characterService struct {
	conf                *config.Config
	characterRepository character.CharacterRepository
}

// NewCharacterService creates a new instance of CharacteService.
func NewCharacterService(conf *config.Config, characterRepository character.CharacterRepository) CharacterService {
	return &characterService{
		conf:                conf,
		characterRepository: characterRepository,
	}
}

func (cs *characterService) UpdateCharacterInfo(userId uuid.UUID, newCharacterInfo *models.CharacterInfo, newCategorySprites []models.CategorySprite) error {
	newCharacterInfo.UserId = userId
	if err := cs.characterRepository.UpdateCharacterInfo(newCharacterInfo); err != nil {
		return err
	}
	if err := cs.characterRepository.UpdateCategorySprites(newCategorySprites); err != nil {
		return err
	}
	logger.Logger.Infof("Character info updated for user ID: %s", userId)
	return nil
}

func (cs *characterService) GetCharacterInfo(userId uuid.UUID) (*models.CharacterInfo, []models.CategorySprite, error) {
	info, err := cs.characterRepository.GetCharacterInfo(userId)
	if err != nil {
		return nil, nil, err
	}
	categorySprites, err := cs.characterRepository.GetCatergorySprites(userId)
	if err != nil {
		return nil, nil, err
	}
	return info, categorySprites, nil

}
