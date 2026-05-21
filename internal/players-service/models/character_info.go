package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type CharacterColorHsv struct {
	H float32 `json:"h"`
	S float32 `json:"s"`
	V float32 `json:"v"`
}

func DefaultCharacterColorHsv() CharacterColorHsv {
	return CharacterColorHsv{H: 0, S: 0, V: 100}
}

func (c CharacterColorHsv) Value() (driver.Value, error) {
	return json.Marshal(c)
}

func (c *CharacterColorHsv) Scan(value any) error {
	if value == nil {
		*c = DefaultCharacterColorHsv()
		return nil
	}

	var bytes []byte
	switch typed := value.(type) {
	case []byte:
		bytes = typed
	case string:
		bytes = []byte(typed)
	default:
		return fmt.Errorf("unsupported color value type %T", value)
	}

	if len(bytes) == 0 {
		*c = DefaultCharacterColorHsv()
		return nil
	}

	if err := json.Unmarshal(bytes, c); err != nil {
		return err
	}

	return nil
}

type CharacterInfo struct {
	UserId        uuid.UUID         `gorm:"type:uuid;primaryKey"`
	CharacterName string            `gorm:"unique;not null"`
	CharacterBio  string            `gorm:"not null"`
	SkinColor     CharacterColorHsv `gorm:"type:jsonb;not null" json:"skin_color"`
	HairColor     CharacterColorHsv `gorm:"type:jsonb;not null" json:"hair_color"`
	EyeColor      CharacterColorHsv `gorm:"type:jsonb;not null" json:"eye_color"`
	CreatedAt     time.Time         `gorm:"autoCreateTime"`
	UpdatedAt     time.Time         `gorm:"autoUpdateTime"`
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

// MapToCategorySpriteDeletes returns category IDs that should be unequipped.
func MapToCategorySpriteDeletes(spriteMap map[string]string) []uuid.UUID {
	categoryIds := make([]uuid.UUID, 0)
	for categoryId, spriteId := range spriteMap {
		if spriteId != "" {
			continue
		}

		categoryUUID, err := uuid.Parse(categoryId)
		if err != nil {
			continue
		}

		categoryIds = append(categoryIds, categoryUUID)
	}

	return categoryIds
}
