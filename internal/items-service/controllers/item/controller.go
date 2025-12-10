package item

import "github.com/gin-gonic/gin"

// ItemController defines the interface for item-related HTTP operations.
type ItemController interface {
	// CreateItem handles the creation of a new item.
	CreateItem(c *gin.Context)

	// CreateItemsBatch handles the creation of multiple items.
	CreateItemsBatch(c *gin.Context)

	// GetItemsMetadata retrieves all items metadata.
	GetItemsMetadata(c *gin.Context)

	// GetItemById retrieves a single item by ID.
	GetItemById(c *gin.Context)

	// DeleteItem deletes an item by ID.
	DeleteItem(c *gin.Context)

	// CreateItemCategory creates a new item category.
	CreateItemCategory(c *gin.Context)

	// GetItemCategories retrieves all item categories.
	GetItemCategories(c *gin.Context)

	// DeleteItemCategory deletes an item category by ID.
	DeleteItemCategory(c *gin.Context)
}
