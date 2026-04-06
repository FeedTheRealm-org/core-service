package items

import "github.com/gin-gonic/gin"

// ItemController defines the interface for item-related HTTP operations.
type ItemController interface {
	GetItemsListByWorld(c *gin.Context)

	GetItemById(c *gin.Context)

	UploadItems(c *gin.Context)

	DeleteItem(c *gin.Context)
}
