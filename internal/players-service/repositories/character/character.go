package character

import (
	"github.com/FeedTheRealm-org/core-service/config"
	"github.com/FeedTheRealm-org/core-service/internal/errors"
	player_errors "github.com/FeedTheRealm-org/core-service/internal/players-service/errors"
	"github.com/FeedTheRealm-org/core-service/internal/players-service/models"
	"github.com/google/uuid"
)

type characterRepository struct {
	conf *config.Config
	db   *config.DB
}

// NewCharacterRepository creates a new instance of CharacteRepository.
func NewCharacterRepository(conf *config.Config, db *config.DB) CharacterRepository {
	return &characterRepository{
		conf: conf,
		db:   db,
	}
}

func (cr *characterRepository) UpdateCharacterInfo(newCharacterInfo *models.CharacterInfo) error {
	if err := cr.db.Conn.Save(newCharacterInfo).Error; err != nil {
		if errors.IsDuplicateEntryError(err) {
			return player_errors.NewCharacterNameTaken(err.Error())
		}
		return err
	}
	return nil
}

func (cr *characterRepository) GetCharacterInfo(userId uuid.UUID) (*models.CharacterInfo, error) {
	var characterInfo models.CharacterInfo
	if err := cr.db.Conn.Where("user_id = ?", userId).First(&characterInfo).Error; err != nil {
		if errors.IsRecordNotFound(err) {
			return nil, player_errors.NewCharacterInfoNotFound(err.Error())
		}
		return nil, err
	}
	return &characterInfo, nil
}
