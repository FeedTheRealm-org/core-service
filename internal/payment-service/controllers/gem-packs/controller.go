package gem_packs

import (
	"github.com/gin-gonic/gin"
)

type GemPacksController interface {
	// GetGemPacks handles the request to retrieve a list of available Pack.
	GetAllGemPacks(c *gin.Context)

	// GetGemPackById handles the request to retrieve a package by its ID.
	GetGemPackById(c *gin.Context)

	// CreateGemPack handles the request to create a new package with the provided details.
	CreateGemPack(c *gin.Context)

	// UpdateGemPack handles the request to update the details of an existing package.
	UpdateGemPack(c *gin.Context)

	// DeleteGemPack handles the request to delete a package by its ID.
	DeleteGemPack(c *gin.Context)
}
