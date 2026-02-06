package items

import "github.com/gin-gonic/gin"

// ItemController defines the interface for item-related HTTP operations.
type ItemController interface {
	GetItemsListByCategory(c *gin.Context)

	GetItemById(c *gin.Context)

	// UploadItems handles item upload.
	UploadItems(c *gin.Context)
	AddCategory(c *gin.Context)
}
