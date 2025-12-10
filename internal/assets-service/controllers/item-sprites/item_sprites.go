package itemsprites

import (
	"net/http"
	"os"

	"github.com/FeedTheRealm-org/core-service/config"
	"github.com/FeedTheRealm-org/core-service/internal/assets-service/dtos"
	assets_errors "github.com/FeedTheRealm-org/core-service/internal/assets-service/errors"
	"github.com/FeedTheRealm-org/core-service/internal/assets-service/models"
	"github.com/FeedTheRealm-org/core-service/internal/assets-service/services/item-sprites"
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

func (isc *itemSpritesController) UploadItemSprite(c *gin.Context) {
	categoryIdStr := c.PostForm("category_id")
	if categoryIdStr == "" {
		_ = c.Error(errors.NewBadRequestError("category_id is required"))
		return
	}

	categoryId, err := uuid.Parse(categoryIdStr)
	if err != nil {
		_ = c.Error(errors.NewBadRequestError("invalid category_id format"))
		return
	}

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

	sprite, err := isc.service.UploadSprite(categoryId, reqFile)
	if err != nil {
		if _, ok := err.(*assets_errors.ItemCategoryNotFound); ok {
			_ = c.Error(errors.NewBadRequestError(err.Error()))
			return
		}
		_ = c.Error(err)
		return
	}

	res := &dtos.ItemSpriteResponse{
		Id:         sprite.Id,
		CategoryId: sprite.CategoryId,
		Url:        sprite.Url,
		CreatedAt:  sprite.CreatedAt,
		UpdatedAt:  sprite.UpdatedAt,
	}

	common_handlers.HandleSuccessResponse(c, http.StatusCreated, res)
}

func (isc *itemSpritesController) GetAllItemSprites(c *gin.Context) {
	// Optional category filter
	categoryIdStr := c.Query("category_id")

	var sprites []models.ItemSprite
	var err error

	if categoryIdStr != "" {
		categoryId, parseErr := uuid.Parse(categoryIdStr)
		if parseErr != nil {
			_ = c.Error(errors.NewBadRequestError("invalid category_id format"))
			return
		}
		sprites, err = isc.service.GetSpritesByCategory(categoryId)
	} else {
		sprites, err = isc.service.GetAllSprites()
	}

	if err != nil {
		if _, ok := err.(*assets_errors.ItemCategoryNotFound); ok {
			_ = c.Error(errors.NewBadRequestError(err.Error()))
			return
		}
		_ = c.Error(err)
		return
	}

	responseSprites := make([]dtos.ItemSpriteResponse, len(sprites))
	for i, sprite := range sprites {
		responseSprites[i] = dtos.ItemSpriteResponse{
			Id:         sprite.Id,
			CategoryId: sprite.CategoryId,
			Url:        sprite.Url,
			CreatedAt:  sprite.CreatedAt,
			UpdatedAt:  sprite.UpdatedAt,
		}
	}

	res := &dtos.ItemSpritesListResponse{
		Sprites: responseSprites,
	}

	common_handlers.HandleSuccessResponse(c, http.StatusOK, res)
}

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

func (isc *itemSpritesController) GetItemCategories(c *gin.Context) {
	categories, err := isc.service.GetAllCategories()
	if err != nil {
		_ = c.Error(err)
		return
	}

	responseCategories := make([]dtos.ItemCategoryResponse, len(categories))
	for i, cat := range categories {
		responseCategories[i] = dtos.ItemCategoryResponse{
			Id:        cat.Id,
			Name:      cat.Name,
			CreatedAt: cat.CreatedAt,
			UpdatedAt: cat.UpdatedAt,
		}
	}

	res := &dtos.ItemCategoriesListResponse{
		Categories: responseCategories,
	}

	common_handlers.HandleSuccessResponse(c, http.StatusOK, res)
}
