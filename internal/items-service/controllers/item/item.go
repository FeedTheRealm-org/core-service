package item

import (
	"net/http"

	"github.com/FeedTheRealm-org/core-service/config"
	"github.com/FeedTheRealm-org/core-service/internal/common_handlers"
	"github.com/FeedTheRealm-org/core-service/internal/errors"
	"github.com/FeedTheRealm-org/core-service/internal/items-service/dtos"
	item_errors "github.com/FeedTheRealm-org/core-service/internal/items-service/errors"
	"github.com/FeedTheRealm-org/core-service/internal/items-service/models"
	"github.com/FeedTheRealm-org/core-service/internal/items-service/services/item"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type itemController struct {
	conf        *config.Config
	itemService item.ItemService
}

// NewItemController creates a new instance of ItemController.
func NewItemController(conf *config.Config, itemService item.ItemService) ItemController {
	return &itemController{
		conf:        conf,
		itemService: itemService,
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

	newItem := &models.Item{
		Name:        req.Name,
		Description: req.Description,
	}
	// Only set SpriteId if provided (zero UUID means no sprite yet)
	newItem.SpriteId = req.SpriteId

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
		}
		items[i].SpriteId = itemReq.SpriteId
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
		SpriteId:    item.SpriteId,
		CreatedAt:   item.CreatedAt,
		UpdatedAt:   item.UpdatedAt,
	}

	common_handlers.HandleSuccessResponse(ctx, http.StatusOK, res)
}

// @Summary UpdateItemSprite
// @Description Updates the sprite associated to an item
// @Tags items-service
// @Accept   json
// @Produce  json
// @Param id path string true "Item UUID"
// @Param request body dtos.UpdateItemSpriteRequest true "Sprite data"
// @Success 200  {object}  dtos.ItemMetadataResponse "Item updated"
// @Failure 400  {object}  dtos.ErrorResponse "Invalid item ID or bad request body"
// @Failure 401  {object}  dtos.ErrorResponse "Invalid credentials or invalid JWT token"
// @Failure 404  {object}  dtos.ErrorResponse "Item not found"
// @Router /items/{id}/sprite [patch]
func (ic *itemController) UpdateItemSprite(ctx *gin.Context) {
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

	req := &dtos.UpdateItemSpriteRequest{}
	if err := ctx.ShouldBindJSON(req); err != nil {
		_ = ctx.Error(errors.NewBadRequestError(err.Error()))
		return
	}

	if err := ic.itemService.UpdateItemSprite(itemId, req.SpriteId); err != nil {
		if _, ok := err.(*item_errors.ItemNotFound); ok {
			_ = ctx.Error(errors.NewNotFoundError("item not found"))
			return
		}
		_ = ctx.Error(err)
		return
	}

	updatedItem, err := ic.itemService.GetItemById(itemId)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	res := &dtos.ItemMetadataResponse{
		Id:          updatedItem.Id,
		Name:        updatedItem.Name,
		Description: updatedItem.Description,
		SpriteId:    updatedItem.SpriteId,
		CreatedAt:   updatedItem.CreatedAt,
		UpdatedAt:   updatedItem.UpdatedAt,
	}

	common_handlers.HandleSuccessResponse(ctx, http.StatusOK, res)
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
