package models

import (
	"time"

	"github.com/google/uuid"
)

type CharacterInfo struct {
	UserId        uuid.UUID `gorm:"type:uuid;primaryKey"`
	CharacterName string    `gorm:"unique;not null"`
	CharacterBio  string    `gorm:"not null"`
	CreatedAt     time.Time `gorm:"autoCreateTime"`
	UpdatedAt     time.Time `gorm:"autoUpdateTime"`
}

type CategorySprite struct {
	UserId     uuid.UUID `gorm:"type:uuid;primaryKey"`
	CategoryId uuid.UUID `gorm:"type:uuid;primaryKey"`
	SpriteId   uuid.UUID `gorm:"type:uuid"`
	CreatedAt  time.Time `gorm:"autoCreateTime"`
	UpdatedAt  time.Time `gorm:"autoUpdateTime"`
}

// CategorySpritesToMap is a helper function to convert CategorySprite slice to map
func CategorySpritesToMap(sprites []CategorySprite) map[string]string {
	result := make(map[string]string)
	for _, sprite := range sprites {
		result[sprite.CategoryId.String()] = sprite.SpriteId.String()
	}
	return result
}

// MapToCategorySprites is a helper function to convert map to CategorySprite slice
func MapToCategorySprites(userId uuid.UUID, spriteMap map[string]string) []CategorySprite {
	sprites := make([]CategorySprite, 0, len(spriteMap))
	for categoryId, spriteId := range spriteMap {
		categoryUUID, errCateg := uuid.Parse(categoryId)
		spriteUUID, errSprite := uuid.Parse(spriteId)
		if errCateg != nil || errSprite != nil {
			continue
		}
		sprites = append(sprites, CategorySprite{
			UserId:     userId,
			CategoryId: categoryUUID,
			SpriteId:   spriteUUID,
		})
	}
	return sprites
}
