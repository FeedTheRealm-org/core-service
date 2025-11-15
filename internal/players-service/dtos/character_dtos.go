package dtos

import "time"

// PatchCharacterInfoRequest represents the request payload for updating character information.
type PatchCharacterInfoRequest struct {
	CharacterName   string            `json:"character_name"`
	CharacterBio    string            `json:"character_bio"`
	CategorySprites map[string]string `json:"category_sprites"`
}

// CharacterInfoResponse represents the response payload for retrieving character information.
type CharacterInfoResponse struct {
	CharacterName   string            `json:"character_name"`
	CharacterBio    string            `json:"character_bio"`
	CategorySprites map[string]string `json:"category_sprites"`
	CreatedAt       time.Time         `json:"created_at"`
	UpdatedAt       time.Time         `json:"updated_at"`
}
