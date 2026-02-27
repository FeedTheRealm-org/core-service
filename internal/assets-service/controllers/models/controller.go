package sprites

import "github.com/gin-gonic/gin"

// ModelsController defines the interface for model-related HTTP operations.
type ModelsController interface {

	// RetrieveWorldAssetIds retrieves asset IDs associated with a specific world.
	GetModelsList(c *gin.Context)

	// UploadSpriteData handles the upload of sprite file.
	UploadModels(c *gin.Context)
}
