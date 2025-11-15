package errors

// CharacterInfoNotFound is returned when the character info for a given user ID is not found.
type CharacterInfoNotFound struct {
	details string
}

func (e *CharacterInfoNotFound) Error() string {
	return e.details
}

func NewCharacterInfoNotFound(details string) *CharacterInfoNotFound {
	return &CharacterInfoNotFound{
		details: details,
	}
}

// CharacterNameTaken is returned when trying to create or update a character with a name that is already taken.
type CharacterNameTaken struct {
	details string
}

func (e *CharacterNameTaken) Error() string {
	return e.details
}

func NewCharacterNameTaken(details string) *CharacterNameTaken {
	return &CharacterNameTaken{
		details: details,
	}
}

// CategorySpritesNotFound is returned when the category sprites for a given user ID are not found.
type CategorySpritesNotFound struct {
	details string
}

func (e *CategorySpritesNotFound) Error() string {
	return e.details
}

func NewCategorySpritesNotFound(details string) *CategorySpritesNotFound {
	return &CategorySpritesNotFound{
		details: details,
	}
}
