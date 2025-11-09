package errors

// SpriteNotFound is returned when a requested sprite cannot be found.
type SpriteNotFound struct {
	details string
}

func (e *SpriteNotFound) Error() string {
	return e.details
}

func NewSpriteNotFound(details string) *SpriteNotFound {
	return &SpriteNotFound{
		details: details,
	}
}

// CategoryNotFound is returned when a requested category cannot be found.
type CategoryNotFound struct {
	details string
}

func (e *CategoryNotFound) Error() string {
	return e.details
}

func NewCategoryNotFound(details string) *CategoryNotFound {
	return &CategoryNotFound{
		details: details,
	}
}
