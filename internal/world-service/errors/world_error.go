package errors

// WorldInfoNotFound is returned when the character info for a given user ID is not found.
type WorldInfoNotFound struct {
	details string
}

func (e *WorldInfoNotFound) Error() string {
	return e.details
}

func NewWorldNotFound(details string) *WorldInfoNotFound {
	return &WorldInfoNotFound{
		details: details,
	}
}

// WorldNameTaken is returned when trying to create or update a character with a name that is already taken.
type WorldNameTaken struct {
	details string
}

func (e *WorldNameTaken) Error() string {
	return e.details
}

func NewWorldNameTaken(details string) *WorldNameTaken {
	return &WorldNameTaken{
		details: details,
	}
}
