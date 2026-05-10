package models

import "github.com/gin-gonic/gin"

// ModelsController defines the interface for model-related HTTP operations.
type ModelsController interface {

	// RetrieveWorldAssetIds retrieves asset IDs associated with a specific world.
	GetModelsList(c *gin.Context)

	// UploadModel handles the uploading of a new model to the system.
	UploadModel(c *gin.Context)
}
