package world_access

import (
	"time"

	"github.com/FeedTheRealm-org/core-service/config"
	core_errors "github.com/FeedTheRealm-org/core-service/internal/errors"
	player_errors "github.com/FeedTheRealm-org/core-service/internal/players-service/errors"
	"github.com/FeedTheRealm-org/core-service/internal/players-service/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type worldAccessRepository struct {
	conf *config.Config
	db   *config.DB
}

func NewWorldAccessRepository(conf *config.Config, db *config.DB) WorldAccessRepository {
	return &worldAccessRepository{conf: conf, db: db}
}

func (wr *worldAccessRepository) CreateWorldJoinToken(token *models.WorldJoinToken) error {
	return wr.db.Conn.Create(token).Error
}

func (wr *worldAccessRepository) ConsumeWorldJoinToken(tokenId uuid.UUID, now time.Time) (*models.WorldJoinToken, error) {
	var token models.WorldJoinToken

	err := wr.db.Conn.Transaction(func(tx *gorm.DB) error {
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("token_id = ?", tokenId).
			First(&token).Error; err != nil {
			if core_errors.IsRecordNotFound(err) {
				return player_errors.NewWorldJoinTokenNotFound("world join token not found")
			}
			return err
		}

		if token.ConsumedAt != nil {
			return player_errors.NewWorldJoinTokenConsumed("world join token already consumed")
		}

		if !token.ExpiresAt.After(now) {
			return player_errors.NewWorldJoinTokenExpired("world join token expired")
		}

		consumedAt := now
		if err := tx.Model(&token).Updates(map[string]interface{}{
			"consumed_at": consumedAt,
			"updated_at":  consumedAt,
		}).Error; err != nil {
			return err
		}

		token.ConsumedAt = &consumedAt
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &token, nil
}
