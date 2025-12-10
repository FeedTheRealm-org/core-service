package bundles

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/FeedTheRealm-org/core-service/config"
	"github.com/FeedTheRealm-org/core-service/internal/assets-service/dtos"
	"github.com/FeedTheRealm-org/core-service/internal/assets-service/models"
	service "github.com/FeedTheRealm-org/core-service/internal/assets-service/services/bundles"
	"github.com/FeedTheRealm-org/core-service/internal/common_handlers"
	"github.com/FeedTheRealm-org/core-service/internal/errors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type bundleController struct {
	conf          *config.Config
	bundleService service.BundleService
}

// NewBundleController creates a new instance of BundleController.
func NewBundleController(conf *config.Config, bundleService service.BundleService) BundleController {
	return &bundleController{
		conf:          conf,
		bundleService: bundleService,
	}
}

// UploadWorldBundle uploads a single bundle file for a specific world
// @Summary Upload world bundle
// @Description Upload a bundle file (ZIP) for a specific world containing all models and materials
// @Tags Bundles
// @Accept multipart/form-data
// @Produce json
// @Param world_id path string true "World ID" format(uuid)
// @Param bundle formData file true "Bundle file (ZIP)"
// @Security BearerAuth
// @Success 201 {object} dtos.BundlePublishResponse "Successfully uploaded bundle"
// @Failure 400 {object} map[string]interface{} "Bad request - missing required fields or invalid format"
// @Failure 401 {object} map[string]interface{} "Unauthorized - invalid or missing JWT token"
// @Failure 500 {object} map[string]interface{} "Internal server error - failed to save bundle"
// @Router /assets/bundles/{world_id} [post]
func (bc *bundleController) UploadWorldBundle(c *gin.Context) {
	// Authenticate user
	_, err := common_handlers.GetUserIDFromSession(c)
	if err != nil {
		_ = c.Error(errors.NewUnauthorizedError(err.Error()))
		return
	}

	// Get world ID from path parameter
	worldIDStr := c.Param("world_id")
	if worldIDStr == "" {
		_ = c.Error(errors.NewBadRequestError("world_id is required"))
		return
	}

	worldID, err := uuid.Parse(worldIDStr)
	if err != nil {
		_ = c.Error(errors.NewBadRequestError("invalid world_id format"))
		return
	}

	// Get bundle file from form
	bundleFile, err := c.FormFile("bundle")
	if err != nil {
		_ = c.Error(errors.NewBadRequestError("bundle file is required"))
		return
	}

	// Create bundle model with the file
	bundle := models.Bundle{
		WorldID:    worldID,
		BundleFile: bundleFile,
	}

	// Save bundle via service
	savedBundle, err := bc.bundleService.PublishWorldBundle(bundle)
	if err != nil {
		_ = c.Error(errors.NewInternalServerError("failed to publish bundle: " + err.Error()))
		return
	}

	// Create response
	response := dtos.BundlePublishResponse{
		WorldID:   savedBundle.WorldID,
		BundleURL: savedBundle.BundleURL,
	}

	common_handlers.HandleSuccessResponse(c, http.StatusCreated, response)
}

// DownloadWorldBundle downloads the bundle file for a specific world
// @Summary Download world bundle
// @Description Downloads the bundle file (ZIP) containing all models and materials for a specific world
// @Tags Bundles
// @Accept json
// @Produce application/octet-stream
// @Param world_id path string true "World ID" format(uuid)
// @Security BearerAuth
// @Success 200 {file} binary "Bundle file"
// @Failure 400 {object} map[string]interface{} "Bad request - invalid world_id format"
// @Failure 401 {object} map[string]interface{} "Unauthorized - invalid or missing JWT token"
// @Failure 404 {object} map[string]interface{} "Not found - bundle not found for this world"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /assets/bundles/{world_id} [get]
func (bc *bundleController) DownloadWorldBundle(c *gin.Context) {
	worldIDStr := c.Param("world_id")
	if worldIDStr == "" {
		_ = c.Error(errors.NewBadRequestError("world_id is required"))
		return
	}

	worldID, err := uuid.Parse(worldIDStr)
	if err != nil {
		_ = c.Error(errors.NewBadRequestError("invalid world_id format"))
		return
	}

	// Get bundle from service
	bundle, err := bc.bundleService.GetWorldBundle(worldID)
	if err != nil {
		_ = c.Error(errors.NewNotFoundError("bundle not found for this world"))
		return
	}

	// Check if file exists
	if _, err := os.Stat(bundle.BundleURL); os.IsNotExist(err) {
		_ = c.Error(errors.NewNotFoundError("bundle file not found on disk"))
		return
	}

	// Set headers for file download
	c.Header("Content-Type", "application/octet-stream")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filepath.Base(bundle.BundleURL)))

	// Serve the file
	c.File(bundle.BundleURL)
}
