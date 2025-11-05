package character

type characterRepository struct {
}

// NewCharacterRepository creates a new instance of CharacteRepository.
func NewCharacterRepository() CharacterRepository {
	return &characterRepository{}
}

func (cr *characterRepository) UpdateCharacterInfo() {}

func (cr *characterRepository) GetCharacterInfo() {}
