package cosmetics

import "github.com/gin-gonic/gin"

// CosmeticsController defines the interface for cosmetic-related HTTP operations.
type CosmeticsController interface {
	// GetCategoriesList retrieves a list of cosmetic categories.
	GetCategoriesList(c *gin.Context)

	// GetCosmeticsListByCategory retrieves a list of cosmetics for a given category.
	GetCosmeticsListByCategory(c *gin.Context)

	// GetCosmeticById retrieves a cosmetic by its ID.
	GetCosmeticById(c *gin.Context)

	// GetCosmeticsListByWorld retrieves a list of cosmetics for a given world.
	GetCosmeticsListByWorld(c *gin.Context)

	// UploadCosmeticData handles the upload of cosmetic file.
	UploadCosmeticData(c *gin.Context)

	// UploadCosmeticByID links an existing cosmetic sprite to another category.
	UploadCosmeticByID(c *gin.Context)

	// DeleteCosmetic handles the deletion of a cosmetic by its ID.
	DeleteCosmetic(c *gin.Context)

	// AddCategory handles the addition of a new sprite category.
	AddCategory(c *gin.Context)
}
