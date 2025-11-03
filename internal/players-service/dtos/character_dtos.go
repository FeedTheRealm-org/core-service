package dtos

// UpdateCharacterInfoRequest represents the request payload for updating character information.
type UpdateCharacterInfoRequest struct {
	CharacterName string `json:"character_name"`
	CharacterBio  string `json:"character_bio"`
}

// CharacterInfoResponse represents the response payload for retrieving character information.
type CharacterInfoResponse struct {
	CharacterName string `json:"character_name"`
	CharacterBio  string `json:"character_bio"`
}
