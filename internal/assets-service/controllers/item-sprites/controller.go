package itemsprites

import "github.com/gin-gonic/gin"

// ItemSpritesController defines the interface for item sprite-related HTTP operations.
type ItemSpritesController interface {
	// UploadItemSprite handles sprite file upload for items.
	UploadItemSprite(c *gin.Context)

	// GetAllItemSprites retrieves all item sprites.
	GetAllItemSprites(c *gin.Context)

	// DownloadItemSprite handles sprite file download by ID.
	DownloadItemSprite(c *gin.Context)

	// DeleteItemSprite deletes a sprite by ID.
	DeleteItemSprite(c *gin.Context)
}
