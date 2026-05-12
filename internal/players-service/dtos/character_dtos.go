package dtos

import (
	"time"

	"github.com/FeedTheRealm-org/core-service/internal/players-service/models"
)

// PatchCharacterInfoRequest represents the request payload for updating character information.
type PatchCharacterInfoRequest struct {
	CharacterName   string                    `json:"character_name"`
	CharacterBio    string                    `json:"character_bio"`
	SkinColor       *models.CharacterColorHsv `json:"skin_color"`
	HairColor       *models.CharacterColorHsv `json:"hair_color"`
	EyeColor        *models.CharacterColorHsv `json:"eye_color"`
	CategorySprites map[string]string         `json:"category_sprites"`
}

// CharacterInfoResponse represents the response payload for retrieving character information.
type CharacterInfoResponse struct {
	CharacterName   string                   `json:"character_name"`
	CharacterBio    string                   `json:"character_bio"`
	SkinColor       models.CharacterColorHsv `json:"skin_color"`
	HairColor       models.CharacterColorHsv `json:"hair_color"`
	EyeColor        models.CharacterColorHsv `json:"eye_color"`
	CategorySprites map[string]string        `json:"category_sprites"`
	CreatedAt       time.Time                `json:"created_at"`
	UpdatedAt       time.Time                `json:"updated_at"`
}
