package materials

import "github.com/gin-gonic/gin"

type MaterialsController interface {
	// GetMaterialsList retrieves a list of materials.
	GetMaterialsList(c *gin.Context)

	// for a given world and saves its metadata.
	UploadMaterials(c *gin.Context)

	// DeleteMaterial deletes a material by its ID.
	DeleteMaterial(c *gin.Context)
}
