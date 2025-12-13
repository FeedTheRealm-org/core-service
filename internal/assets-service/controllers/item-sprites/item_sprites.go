package itemsprites

import (
	"net/http"
	"os"

	"github.com/FeedTheRealm-org/core-service/config"
	"github.com/FeedTheRealm-org/core-service/internal/assets-service/dtos"
	assets_errors "github.com/FeedTheRealm-org/core-service/internal/assets-service/errors"
	itemsprites "github.com/FeedTheRealm-org/core-service/internal/assets-service/services/item-sprites"
	"github.com/FeedTheRealm-org/core-service/internal/common_handlers"
	"github.com/FeedTheRealm-org/core-service/internal/errors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type itemSpritesController struct {
	conf    *config.Config
	service itemsprites.ItemSpritesService
}

// NewItemSpritesController creates a new instance of ItemSpritesController.
func NewItemSpritesController(conf *config.Config, service itemsprites.ItemSpritesService) ItemSpritesController {
	return &itemSpritesController{
		conf:    conf,
		service: service,
	}
}

// @Summary UploadItemSprite
// @Description Uploads a new item sprite.
// @Tags assets-service
// @Accept multipart/form-data
// @Produce json
// @Param sprite formData file true "Item sprite file (PNG or JPEG)"
// @Success 201 {object} dtos.ItemSpriteResponse "Uploaded item sprite"
// @Failure 400 {object} dtos.ErrorResponse "Bad request"
// @Failure 401 {object} dtos.ErrorResponse "Invalid credentials or invalid JWT token"
// @Router /assets/sprites/items [post]
func (isc *itemSpritesController) UploadItemSprite(c *gin.Context) {
	reqFile, err := c.FormFile("sprite")
	if err != nil {
		_ = c.Error(errors.NewBadRequestError("failed to get sprite file from request: " + err.Error()))
		return
	}

	if reqFile.Size > isc.conf.Assets.MaxUploadSizeBytes {
		_ = c.Error(errors.NewBadRequestError("file size exceeds the limit"))
		return
	}

	contentType := reqFile.Header.Get("Content-Type")
	if contentType != "image/png" && contentType != "image/jpeg" {
		_ = c.Error(errors.NewBadRequestError("file must be PNG or JPEG format"))
		return
	}

	sprite, err := isc.service.UploadSprite(reqFile)
	if err != nil {
		_ = c.Error(err)
		return
	}

	res := &dtos.ItemSpriteResponse{
		Id:        sprite.Id,
		Url:       sprite.Url,
		CreatedAt: sprite.CreatedAt,
		UpdatedAt: sprite.UpdatedAt,
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

// (Item sprite categories endpoint removed)
