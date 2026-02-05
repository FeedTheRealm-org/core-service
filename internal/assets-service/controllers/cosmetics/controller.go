package cosmetics

import "github.com/gin-gonic/gin"

// SpritesController defines the interface for sprite-related HTTP operations.
type SpritesController interface {

	// GetCategoriesList retrieves a list of sprite categories.
	GetCategoriesList(c *gin.Context)

	// GetSpritesList retrieves a list of sprites for a given category.
	GetSpritesListByCategory(c *gin.Context)

	// AddCategory handles the addition of a new sprite category.
	AddCategory(c *gin.Context)

	// UploadSpriteData handles the upload of sprite file.
	UploadSpriteData(c *gin.Context)
}
