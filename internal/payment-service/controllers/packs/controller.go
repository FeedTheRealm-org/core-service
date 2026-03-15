package packs

import (
	"github.com/gin-gonic/gin"
)

type PacksController interface {
	// GetPacks handles the request to retrieve a list of available Pack.
	GetAllPacks(c *gin.Context)

	// GetPackById handles the request to retrieve a package by its ID.
	GetPackById(c *gin.Context)

	// CreatePack handles the request to create a new package with the provided details.
	CreatePack(c *gin.Context)

	// UpdatePack handles the request to update the details of an existing package.
	UpdatePack(c *gin.Context)

	// DeletePack handles the request to delete a package by its ID.
	DeletePack(c *gin.Context)
}
