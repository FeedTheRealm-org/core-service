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

	// UploadItemSprite handles sprite file upload for items.
	UploadItemSprite(c *gin.Context)

	// DownloadItemSprite handles sprite file download by ID.
	DownloadItemSprite(c *gin.Context)

	// DownloadItemSpriteByCategory handles sprite file download with category path.
	DownloadItemSpriteByCategory(c *gin.Context)

	// DeleteItemSprite deletes a sprite by ID.
	DeleteItemSprite(c *gin.Context)
}
