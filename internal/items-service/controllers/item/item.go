package item

import (
	"net/http"
	"os"

	"github.com/FeedTheRealm-org/core-service/config"
	"github.com/FeedTheRealm-org/core-service/internal/common_handlers"
	"github.com/FeedTheRealm-org/core-service/internal/errors"
	"github.com/FeedTheRealm-org/core-service/internal/items-service/dtos"
	item_errors "github.com/FeedTheRealm-org/core-service/internal/items-service/errors"
	"github.com/FeedTheRealm-org/core-service/internal/items-service/models"
	"github.com/FeedTheRealm-org/core-service/internal/items-service/services/item"
	"github.com/FeedTheRealm-org/core-service/internal/utils/logger"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type itemController struct {
	conf              *config.Config
	itemService       item.ItemService
	itemSpriteService item.ItemSpriteService
}

// NewItemController creates a new instance of ItemController.
func NewItemController(conf *config.Config, itemService item.ItemService, itemSpriteService item.ItemSpriteService) ItemController {
	return &itemController{
		conf:              conf,
		itemService:       itemService,
		itemSpriteService: itemSpriteService,
	}
}

// @Summary CreateItem
// @Description Creates a new game item with metadata
// @Tags items-service
// @Accept   json
// @Produce  json
// @Param   request body dtos.CreateItemRequest true "Item data"
// @Success 201  {object}  dtos.ItemMetadataResponse "Item created"
// @Failure 400  {object}  dtos.ErrorResponse "Bad request body"
// @Failure 401  {object}  dtos.ErrorResponse "Invalid credentials or invalid JWT token"
// @Failure 409  {object}  dtos.ErrorResponse "Item already exists"
// @Router /items [post]
func (ic *itemController) CreateItem(ctx *gin.Context) {
	// _, err := common_handlers.GetUserIDFromSession(ctx)
	// if err != nil {
	// 	_ = ctx.Error(errors.NewUnauthorizedError(err.Error()))
	// 	return
	// }

	req := &dtos.CreateItemRequest{}
	if err := ctx.ShouldBindJSON(req); err != nil {
		_ = ctx.Error(errors.NewBadRequestError(err.Error()))
		return
	}

	// Validations
	if len(req.Name) < 3 || len(req.Name) > 64 {
		_ = ctx.Error(errors.NewBadRequestError("item name must be between 3 and 64 characters"))
		return
	}
	if len(req.Description) > 256 {
		_ = ctx.Error(errors.NewBadRequestError("item description must be less than 256 characters"))
		return
	}
	if len(req.Category) < 3 || len(req.Category) > 32 {
		_ = ctx.Error(errors.NewBadRequestError("category must be between 3 and 32 characters"))
		return
	}

	newItem := &models.Item{
		Name:        req.Name,
		Description: req.Description,
		Category:    req.Category,
		SpriteId:    req.SpriteId,
	}

	if err := ic.itemService.CreateItem(newItem); err != nil {
		if _, ok := err.(*item_errors.ItemAlreadyExists); ok {
			_ = ctx.Error(errors.NewConflictError("item already exists"))
			return
		}
		_ = ctx.Error(err)
		return
	}

	res := &dtos.ItemMetadataResponse{
		Id:          newItem.Id,
		Name:        newItem.Name,
		Description: newItem.Description,
		Category:    newItem.Category,
		SpriteId:    newItem.SpriteId,
		CreatedAt:   newItem.CreatedAt,
		UpdatedAt:   newItem.UpdatedAt,
	}

	common_handlers.HandleSuccessResponse(ctx, http.StatusCreated, res)
}

// @Summary CreateItemsBatch
// @Description Creates multiple game items at once
// @Tags items-service
// @Accept   json
// @Produce  json
// @Param   request body dtos.CreateItemBatchRequest true "Batch of items"
// @Success 201  {object}  dtos.ItemsListResponse "Items created"
// @Failure 400  {object}  dtos.ErrorResponse "Bad request body"
// @Failure 401  {object}  dtos.ErrorResponse "Invalid credentials or invalid JWT token"
// @Router /items/batch [post]
func (ic *itemController) CreateItemsBatch(ctx *gin.Context) {
	// _, err := common_handlers.GetUserIDFromSession(ctx)
	// if err != nil {
	// 	_ = ctx.Error(errors.NewUnauthorizedError(err.Error()))
	// 	return
	// }

	req := &dtos.CreateItemBatchRequest{}
	if err := ctx.ShouldBindJSON(req); err != nil {
		_ = ctx.Error(errors.NewBadRequestError(err.Error()))
		return
	}

	if len(req.Items) == 0 {
		_ = ctx.Error(errors.NewBadRequestError("items list cannot be empty"))
		return
	}

	// Convert DTOs to models
	items := make([]models.Item, len(req.Items))
	for i, itemReq := range req.Items {
		// Basic validation
		if len(itemReq.Name) < 3 || len(itemReq.Name) > 64 {
			_ = ctx.Error(errors.NewBadRequestError("item name must be between 3 and 64 characters"))
			return
		}

		items[i] = models.Item{
			Name:        itemReq.Name,
			Description: itemReq.Description,
			Category:    itemReq.Category,
			SpriteId:    itemReq.SpriteId,
		}
	}

	if err := ic.itemService.CreateItems(items); err != nil {
		_ = ctx.Error(err)
		return
	}

	// Build response
	responseItems := make([]dtos.ItemMetadataResponse, len(items))
	for i, item := range items {
		responseItems[i] = dtos.ItemMetadataResponse{
			Id:          item.Id,
			Name:        item.Name,
			Description: item.Description,
			Category:    item.Category,
			SpriteId:    item.SpriteId,
			CreatedAt:   item.CreatedAt,
			UpdatedAt:   item.UpdatedAt,
		}
	}

	res := &dtos.ItemsListResponse{
		Items: responseItems,
	}

	common_handlers.HandleSuccessResponse(ctx, http.StatusCreated, res)
}

// @Summary GetItemsMetadata
// @Description Retrieves all items metadata (for Unity client initialization)
// @Tags items-service
// @Produce  json
// @Success 200  {object}  dtos.ItemsListResponse "Items list"
// @Failure 401  {object}  dtos.ErrorResponse "Invalid credentials or invalid JWT token"
// @Router /items/metadata [get]
func (ic *itemController) GetItemsMetadata(ctx *gin.Context) {
	// _, err := common_handlers.GetUserIDFromSession(ctx)
	// if err != nil {
	// 	_ = ctx.Error(errors.NewUnauthorizedError(err.Error()))
	// 	return
	// }

	items, err := ic.itemService.GetAllItems()
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	responseItems := make([]dtos.ItemMetadataResponse, len(items))
	for i, item := range items {
		responseItems[i] = dtos.ItemMetadataResponse{
			Id:          item.Id,
			Name:        item.Name,
			Description: item.Description,
			Category:    item.Category,
			SpriteId:    item.SpriteId,
			CreatedAt:   item.CreatedAt,
			UpdatedAt:   item.UpdatedAt,
		}
	}

	res := &dtos.ItemsListResponse{
		Items: responseItems,
	}

	common_handlers.HandleSuccessResponse(ctx, http.StatusOK, res)
}

// @Summary GetItemById
// @Description Retrieves a single item by its ID
// @Tags items-service
// @Produce  json
// @Param id path string true "Item UUID"
// @Success 200  {object}  dtos.ItemMetadataResponse "Item metadata"
// @Failure 400  {object}  dtos.ErrorResponse "Invalid item ID"
// @Failure 401  {object}  dtos.ErrorResponse "Invalid credentials or invalid JWT token"
// @Failure 404  {object}  dtos.ErrorResponse "Item not found"
// @Router /items/{id} [get]
func (ic *itemController) GetItemById(ctx *gin.Context) {
	// _, err := common_handlers.GetUserIDFromSession(ctx)
	// if err != nil {
	// 	_ = ctx.Error(errors.NewUnauthorizedError(err.Error()))
	// 	return
	// }

	itemIdStr := ctx.Param("id")
	itemId, err := uuid.Parse(itemIdStr)
	if err != nil {
		_ = ctx.Error(errors.NewBadRequestError("invalid item ID format"))
		return
	}

	item, err := ic.itemService.GetItemById(itemId)
	if err != nil {
		if _, ok := err.(*item_errors.ItemNotFound); ok {
			_ = ctx.Error(errors.NewNotFoundError("item not found"))
			return
		}
		_ = ctx.Error(err)
		return
	}

	res := &dtos.ItemMetadataResponse{
		Id:          item.Id,
		Name:        item.Name,
		Description: item.Description,
		Category:    item.Category,
		SpriteId:    item.SpriteId,
		CreatedAt:   item.CreatedAt,
		UpdatedAt:   item.UpdatedAt,
	}

	common_handlers.HandleSuccessResponse(ctx, http.StatusOK, res)
}

// @Summary UploadItemSprite
// @Description Uploads a sprite file for items
// @Tags items-service
// @Accept  multipart/form-data
// @Produce  json
// @Param sprite formData file true "Sprite file"
// @Param category formData string true "Category (armor/weapon/consumable)"
// @Success 201  {object}  dtos.ItemSpriteResponse "Uploaded sprite"
// @Failure 400  {object}  dtos.ErrorResponse "Bad request"
// @Failure 401  {object}  dtos.ErrorResponse "Invalid credentials or invalid JWT token"
// @Router /assets/sprites/items [post]
func (ic *itemController) UploadItemSprite(ctx *gin.Context) {
	// _, err := common_handlers.GetUserIDFromSession(ctx)
	// if err != nil {
	// 	_ = ctx.Error(errors.NewUnauthorizedError(err.Error()))
	// 	return
	// }

	category := ctx.PostForm("category")
	if category == "" {
		_ = ctx.Error(errors.NewBadRequestError("category is required"))
		return
	}

	reqFile, err := ctx.FormFile("sprite")
	if err != nil {
		_ = ctx.Error(errors.NewBadRequestError("failed to get sprite file from request: " + err.Error()))
		return
	}

	// Validate file size (using same config as assets-service)
	if reqFile.Size > ic.conf.Assets.MaxUploadSizeBytes {
		_ = ctx.Error(errors.NewBadRequestError("file size exceeds the limit"))
		return
	}

	// Validate content type
	contentType := reqFile.Header.Get("Content-Type")
	if contentType != "image/png" && contentType != "image/jpeg" {
		_ = ctx.Error(errors.NewBadRequestError("file must be PNG or JPEG format"))
		return
	}

	sprite, err := ic.itemSpriteService.UploadSprite(category, reqFile)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	res := &dtos.ItemSpriteResponse{
		Id:        sprite.Id,
		Category:  sprite.Category,
		Url:       sprite.Url,
		CreatedAt: sprite.CreatedAt,
		UpdatedAt: sprite.UpdatedAt,
	}

	common_handlers.HandleSuccessResponse(ctx, http.StatusCreated, res)
}

// @Summary DownloadItemSprite
// @Description Downloads a sprite file by sprite ID (optionally filtered by category via query param)
// @Tags items-service
// @Produce  octet-stream
// @Param sprite_id path string true "Sprite UUID"
// @Param category query string false "Category (armor/weapon/consumable) - optional validation"
// @Success 200  {file}  byte "Sprite file"
// @Failure 400  {object}  dtos.ErrorResponse "Invalid sprite ID or category mismatch"
// @Failure 401  {object}  dtos.ErrorResponse "Invalid credentials or invalid JWT token"
// @Failure 404  {object}  dtos.ErrorResponse "Sprite not found"
// @Router /assets/sprites/items/{sprite_id} [get]
func (ic *itemController) DownloadItemSprite(ctx *gin.Context) {
	// _, err := common_handlers.GetUserIDFromSession(ctx)
	// if err != nil {
	// 	_ = ctx.Error(errors.NewUnauthorizedError(err.Error()))
	// 	return
	// }

	spriteIdStr := ctx.Param("sprite_id")
	spriteId, err := uuid.Parse(spriteIdStr)
	if err != nil {
		_ = ctx.Error(errors.NewBadRequestError("invalid sprite ID format"))
		return
	}

	// Get sprite from database
	sprite, err := ic.itemSpriteService.GetSpriteById(spriteId)
	if err != nil {
		if _, ok := err.(*item_errors.ItemSpriteNotFound); ok {
			_ = ctx.Error(errors.NewNotFoundError("sprite not found"))
			return
		}
		_ = ctx.Error(err)
		return
	}

	// Optional: validate category if provided as query parameter
	category := ctx.Query("category")
	if category != "" && sprite.Category != category {
		_ = ctx.Error(errors.NewBadRequestError("sprite does not belong to specified category"))
		return
	}

	if _, err := os.Stat(sprite.Url); os.IsNotExist(err) {
		_ = ctx.Error(errors.NewNotFoundError("sprite file not found on disk"))
		return
	}

	ctx.File(sprite.Url)
}

// @Summary DeleteItem
// @Description Deletes an item by its ID
// @Tags items-service
// @Produce  json
// @Param id path string true "Item UUID"
// @Success 200  {object}  map[string]interface{} "Item deleted successfully"
// @Failure 400  {object}  dtos.ErrorResponse "Invalid item ID"
// @Failure 401  {object}  dtos.ErrorResponse "Invalid credentials or invalid JWT token"
// @Failure 404  {object}  dtos.ErrorResponse "Item not found"
// @Router /items/{id} [delete]
func (ic *itemController) DeleteItem(ctx *gin.Context) {
	// _, err := common_handlers.GetUserIDFromSession(ctx)
	// if err != nil {
	// 	_ = ctx.Error(errors.NewUnauthorizedError(err.Error()))
	// 	return
	// }

	itemIdStr := ctx.Param("id")
	itemId, err := uuid.Parse(itemIdStr)
	if err != nil {
		_ = ctx.Error(errors.NewBadRequestError("invalid item ID format"))
		return
	}

	// Check if item exists
	_, err = ic.itemService.GetItemById(itemId)
	if err != nil {
		if _, ok := err.(*item_errors.ItemNotFound); ok {
			_ = ctx.Error(errors.NewNotFoundError("item not found"))
			return
		}
		_ = ctx.Error(err)
		return
	}

	if err := ic.itemService.DeleteItem(itemId); err != nil {
		_ = ctx.Error(err)
		return
	}

	common_handlers.HandleSuccessResponse(ctx, http.StatusOK, gin.H{"message": "item deleted successfully"})
}

// @Summary DeleteItemSprite
// @Description Deletes a sprite by its ID and removes the file from disk
// @Tags items-service
// @Produce  json
// @Param sprite_id path string true "Sprite UUID"
// @Success 200  {object}  map[string]interface{} "Sprite deleted successfully"
// @Failure 400  {object}  dtos.ErrorResponse "Invalid sprite ID"
// @Failure 401  {object}  dtos.ErrorResponse "Invalid credentials or invalid JWT token"
// @Failure 404  {object}  dtos.ErrorResponse "Sprite not found"
// @Router /assets/sprites/items/{sprite_id} [delete]
func (ic *itemController) DeleteItemSprite(ctx *gin.Context) {
	// _, err := common_handlers.GetUserIDFromSession(ctx)
	// if err != nil {
	// 	_ = ctx.Error(errors.NewUnauthorizedError(err.Error()))
	// 	return
	// }

	spriteIdStr := ctx.Param("sprite_id")
	spriteId, err := uuid.Parse(spriteIdStr)
	if err != nil {
		_ = ctx.Error(errors.NewBadRequestError("invalid sprite ID format"))
		return
	}

	// Get sprite to get file path before deleting
	sprite, err := ic.itemSpriteService.GetSpriteById(spriteId)
	if err != nil {
		if _, ok := err.(*item_errors.ItemSpriteNotFound); ok {
			_ = ctx.Error(errors.NewNotFoundError("sprite not found"))
			return
		}
		_ = ctx.Error(err)
		return
	}

	// Delete from database
	if err := ic.itemSpriteService.DeleteSprite(spriteId); err != nil {
		_ = ctx.Error(err)
		return
	}

	// Delete file from disk
	if err := os.Remove(sprite.Url); err != nil {
		// Log error but don't fail the request
		// The database record is already deleted
		logger.Logger.Warnf("Failed to remove sprite file from disk: %v", err)
	}

	common_handlers.HandleSuccessResponse(ctx, http.StatusOK, gin.H{"message": "sprite deleted successfully"})
}
