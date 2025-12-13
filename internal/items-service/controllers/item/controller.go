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

	// UpdateItemSprite updates the sprite associated to an item.
	UpdateItemSprite(c *gin.Context)

	// DeleteItem deletes an item by ID.
	DeleteItem(c *gin.Context)
}
