package errors

import "fmt"

// ItemNotFound represents an error when an item is not found.
type ItemNotFound struct {
	Message string
}

func (e *ItemNotFound) Error() string {
	return e.Message
}

func NewItemNotFound(message string) *ItemNotFound {
	return &ItemNotFound{
		Message: message,
	}
}

// InvalidCategory represents an error when an invalid category is provided.
type InvalidCategory struct {
	Category string
}

func (e *InvalidCategory) Error() string {
	return fmt.Sprintf("invalid category: %s", e.Category)
}

func NewInvalidCategory(category string) *InvalidCategory {
	return &InvalidCategory{
		Category: category,
	}
}

// ItemAlreadyExists represents an error when trying to create an item that already exists.
type ItemAlreadyExists struct {
	ItemName string
}

func (e *ItemAlreadyExists) Error() string {
	return fmt.Sprintf("item already exists: %s", e.ItemName)
}

func NewItemAlreadyExists(itemName string) *ItemAlreadyExists {
	return &ItemAlreadyExists{
		ItemName: itemName,
	}
}

// ItemCategoryNotFound represents an error when an item category is not found.
type ItemCategoryNotFound struct {
	CategoryId string
}

func (e *ItemCategoryNotFound) Error() string {
	return fmt.Sprintf("Category with ID %s does not exist", e.CategoryId)
}

func NewItemCategoryNotFound(categoryId string) *ItemCategoryNotFound {
	return &ItemCategoryNotFound{
		CategoryId: categoryId,
	}
}

// ItemCategoryConflict represents an error when a category name already exists.
type ItemCategoryConflict struct {
	Message string
}

func (e *ItemCategoryConflict) Error() string {
	return e.Message
}

func NewItemCategoryConflict(message string) *ItemCategoryConflict {
	return &ItemCategoryConflict{
		Message: message,
	}
}

// ItemCategoryInUse represents an error when trying to delete a category that's in use.
type ItemCategoryInUse struct {
	CategoryName string
	ItemCount    int64
	SpriteCount  int64
}

func (e *ItemCategoryInUse) Error() string {
	totalCount := e.ItemCount + e.SpriteCount
	return fmt.Sprintf("cannot delete category '%s' - it is referenced by %d items/sprites",
		e.CategoryName, totalCount)
}

func NewItemCategoryInUse(categoryName string, itemCount, spriteCount int64) *ItemCategoryInUse {
	return &ItemCategoryInUse{
		CategoryName: categoryName,
		ItemCount:    itemCount,
		SpriteCount:  spriteCount,
	}
}
