package items

import (
	"fmt"
	"mime/multipart"
	"net/http"
	"os"

	"github.com/FeedTheRealm-org/core-service/config"
	"github.com/FeedTheRealm-org/core-service/internal/assets-service/dtos"
	assets_errors "github.com/FeedTheRealm-org/core-service/internal/assets-service/errors"
	"github.com/FeedTheRealm-org/core-service/internal/assets-service/services/items"
	"github.com/FeedTheRealm-org/core-service/internal/common_handlers"
	"github.com/FeedTheRealm-org/core-service/internal/errors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type itemSpritesController struct {
	conf    *config.Config
	service items.ItemSpritesService
}

// NewItemSpritesController creates a new instance of ItemSpritesController.
func NewItemSpritesController(conf *config.Config, service items.ItemSpritesService) ItemSpritesController {
	return &itemSpritesController{
		conf:    conf,
		service: service,
	}
}

// @Summary UploadItemSprites
// @Description Uploads multiple item sprites. Each sprite must have a provided ID.
// @Tags assets-service
// @Accept multipart/form-data
// @Produce json
// @Param world_id path string true "World ID" format(uuid)
// @Param ids[] formData string true "Sprite IDs (UUIDs), one per file"
// @Param sprites[] formData file true "Item sprite files (PNG o JPEG)"
// @Success 201 {object} dtos.ItemSpritesListResponse "Uploaded item sprites"
// @Failure 400 {object} dtos.ErrorResponse "Bad request"
// @Failure 401 {object} dtos.ErrorResponse "Invalid credentials or invalid JWT token"
// @Router /assets/sprites/items/{world_id} [post]
func (isc *itemSpritesController) UploadItemSprite(c *gin.Context) {
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
	var ids []uuid.UUID
	var files []*multipart.FileHeader
	i := 0
	for {
		i++
		idKey := fmt.Sprintf("id[%d]", i)
		spriteKey := fmt.Sprintf("sprite[%d]", i)
		idVal := c.PostForm(idKey)
		if idVal == "" {
			break
		}
		spriteFile, err := c.FormFile(spriteKey)
		if err != nil {
			_ = c.Error(errors.NewBadRequestError(fmt.Sprintf("Missing sprite file for id[%d]", i)))
			return
		}
		id, err := uuid.Parse(idVal)
		if err != nil {
			_ = c.Error(errors.NewBadRequestError(fmt.Sprintf("invalid UUID in id[%d]: %s", i, idVal)))
			return
		}
		if spriteFile.Size > isc.conf.Assets.MaxUploadSizeBytes {
			_ = c.Error(errors.NewBadRequestError("file size exceeds the limit"))
			return
		}
		contentType := spriteFile.Header.Get("Content-Type")
		if contentType != "image/png" && contentType != "image/jpeg" && contentType != "application/octet-stream" {
			_ = c.Error(errors.NewBadRequestError("file must be PNG, JPEG, or octet-stream format"))
			return
		}
		ids = append(ids, id)
		files = append(files, spriteFile)
	}
	if len(ids) == 0 || len(files) == 0 || len(ids) != len(files) {
		_ = c.Error(errors.NewBadRequestError("must provide at least one id[N] and sprite[N] pair, and all pairs must be complete"))
		return
	}
	// Check for duplicate IDs in the same request
	seen := make(map[uuid.UUID]struct{})
	for _, id := range ids {
		if _, exists := seen[id]; exists {
			c.AbortWithStatusJSON(http.StatusConflict, gin.H{"error": fmt.Sprintf("duplicate id detected in request: %s", id)})
			return
		}
		seen[id] = struct{}{}
	}

	sprites, err := isc.service.UploadSprites(worldID, ids, files)
	if err != nil {
		_ = c.Error(err)
		return
	}
	responseSprites := make([]dtos.ItemSpriteResponse, len(sprites))
	for i, sprite := range sprites {
		responseSprites[i] = dtos.ItemSpriteResponse{
			Id:        sprite.Id,
			Url:       sprite.Url,
			CreatedAt: sprite.CreatedAt,
			UpdatedAt: sprite.UpdatedAt,
		}
	}
	res := &dtos.ItemSpritesListResponse{
		Sprites: responseSprites,
	}
	common_handlers.HandleSuccessResponse(c, http.StatusCreated, res)
}

// @Summary GetAllItemSprites
// @Description Retrieves all item sprites.
// @Tags assets-service
// @Produce json
// @Success 200 {object} dtos.ItemSpritesListResponse "List of item sprites"
// @Failure 401 {object} dtos.ErrorResponse "Invalid credentials or invalid JWT token"
// @Router /assets/sprites/items [get]
func (isc *itemSpritesController) GetAllItemSprites(c *gin.Context) {
	sprites, err := isc.service.GetAllSprites()
	if err != nil {
		_ = c.Error(err)
		return
	}

	responseSprites := make([]dtos.ItemSpriteResponse, len(sprites))
	for i, sprite := range sprites {
		responseSprites[i] = dtos.ItemSpriteResponse{
			Id:        sprite.Id,
			Url:       sprite.Url,
			CreatedAt: sprite.CreatedAt,
			UpdatedAt: sprite.UpdatedAt,
		}
	}

	res := &dtos.ItemSpritesListResponse{
		Sprites: responseSprites,
	}

	common_handlers.HandleSuccessResponse(c, http.StatusOK, res)
}

// @Summary DownloadItemSprite
// @Description Downloads an item sprite by its ID.
// @Tags assets-service
// @Produce octet-stream
// @Param sprite_id path string true "Item sprite ID"
// @Success 200 {file} file "Item sprite file"
// @Failure 400 {object} dtos.ErrorResponse "Bad request"
// @Failure 401 {object} dtos.ErrorResponse "Invalid credentials or invalid JWT token"
// @Failure 404 {object} dtos.ErrorResponse "Sprite not found"
// @Router /assets/sprites/items/{sprite_id} [get]
func (isc *itemSpritesController) DownloadItemSprite(c *gin.Context) {
	spriteIdStr := c.Param("sprite_id")
	spriteId, err := uuid.Parse(spriteIdStr)
	if err != nil {
		_ = c.Error(errors.NewBadRequestError("invalid sprite ID format"))
		return
	}

	sprite, err := isc.service.GetSpriteById(spriteId)
	if err != nil {
		if _, ok := err.(*assets_errors.ItemSpriteNotFound); ok {
			_ = c.Error(errors.NewNotFoundError("sprite not found"))
			return
		}
		_ = c.Error(err)
		return
	}

	if _, err := os.Stat(sprite.Url); os.IsNotExist(err) {
		_ = c.Error(errors.NewNotFoundError("sprite file not found on disk"))
		return
	}

	c.File(sprite.Url)
}

// @Summary DeleteItemSprite
// @Description Deletes an item sprite by its ID.
// @Tags assets-service
// @Param sprite_id path string true "Item sprite ID"
// @Success 200 {object} map[string]string "Sprite deleted successfully"
// @Failure 400 {object} dtos.ErrorResponse "Bad request"
// @Failure 401 {object} dtos.ErrorResponse "Invalid credentials or invalid JWT token"
// @Failure 404 {object} dtos.ErrorResponse "Sprite not found"
// @Router /assets/sprites/items/{sprite_id} [delete]
func (isc *itemSpritesController) DeleteItemSprite(c *gin.Context) {
	spriteIdStr := c.Param("sprite_id")
	spriteId, err := uuid.Parse(spriteIdStr)
	if err != nil {
		_ = c.Error(errors.NewBadRequestError("invalid sprite ID format"))
		return
	}

	if err := isc.service.DeleteSprite(spriteId); err != nil {
		if _, ok := err.(*assets_errors.ItemSpriteNotFound); ok {
			_ = c.Error(errors.NewNotFoundError("sprite not found"))
			return
		}
		_ = c.Error(err)
		return
	}

	common_handlers.HandleSuccessResponse(c, http.StatusOK, gin.H{"message": "sprite deleted successfully"})
}
