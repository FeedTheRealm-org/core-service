package character

import "github.com/FeedTheRealm-org/core-service/internal/players-service/repositories/character"

type characterService struct {
	characterRepository character.CharacterRepository
}

// NewCharacterService creates a new instance of CharacteService.
func NewCharacterService(characterRepository character.CharacterRepository) CharacterService {
	return &characterService{
		characterRepository: characterRepository,
	}
}

func (cs *characterService) UpdateCharacterInfo() {}

func (cs *characterService) GetCharacterInfo() {}
