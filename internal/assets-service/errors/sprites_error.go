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

// CategoryConflict is returned in conflict scenarios, such as when attempting to create a category that already exists.
type CategoryConflict struct {
	details string
}

func (e *CategoryConflict) Error() string {
	return e.details
}

func NewCategoryConflict(details string) *CategoryConflict {
	return &CategoryConflict{
		details: details,
	}
}

// ItemSpriteNotFound is returned when a requested item sprite cannot be found.
type ItemSpriteNotFound struct {
	details string
}

func (e *ItemSpriteNotFound) Error() string {
	return e.details
}

func NewItemSpriteNotFound(details string) *ItemSpriteNotFound {
	return &ItemSpriteNotFound{
		details: details,
	}
}

// ItemCategoryNotFound is returned when a requested item category cannot be found.
type ItemCategoryNotFound struct {
	CategoryId string
}

func (e *ItemCategoryNotFound) Error() string {
	return "Category with ID " + e.CategoryId + " does not exist"
}

func NewItemCategoryNotFound(categoryId string) *ItemCategoryNotFound {
	return &ItemCategoryNotFound{
		CategoryId: categoryId,
	}
}
