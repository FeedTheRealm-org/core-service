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
