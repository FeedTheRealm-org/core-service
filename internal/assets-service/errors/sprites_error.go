package errors

// CosmeticNotFound is returned when a requested sprite cannot be found.
type CosmeticNotFound struct {
	details string
}

func (e *CosmeticNotFound) Error() string {
	return e.details
}

func NewCosmeticNotFound(details string) *CosmeticNotFound {
	return &CosmeticNotFound{
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

// WorldNotFound is returned when a requested world cannot be found.
type WorldNotFound struct {
	WorldId string
}

func (e *WorldNotFound) Error() string {
	return "World with ID " + e.WorldId + " does not exist"
}

func NewWorldNotFound(worldId string) *WorldNotFound {
	return &WorldNotFound{
		WorldId: worldId,
	}
}
