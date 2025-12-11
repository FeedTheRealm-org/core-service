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
	conf                *config.Config
	itemService         item.ItemService
	itemCategoryService item.ItemCategoryService
}

// NewItemController creates a new instance of ItemController.
func NewItemController(conf *config.Config, itemService item.ItemService, itemCategoryService item.ItemCategoryService) ItemController {
	return &itemController{
		conf:                conf,
		itemService:         itemService,
		itemCategoryService: itemCategoryService,
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

	// Validate that category exists
	_, err := ic.itemCategoryService.GetCategoryById(req.CategoryId)
	if err != nil {
		if _, ok := err.(*item_errors.ItemCategoryNotFound); ok {
			_ = ctx.Error(errors.NewBadRequestError(err.Error()))
			return
		}
		_ = ctx.Error(err)
		return
	}

	newItem := &models.Item{
		Name:        req.Name,
		Description: req.Description,
		CategoryId:  req.CategoryId,
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
		CategoryId:  newItem.CategoryId,
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
			CategoryId:  itemReq.CategoryId,
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
			CategoryId:  item.CategoryId,
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
			CategoryId:  item.CategoryId,
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
		CategoryId:  item.CategoryId,
		SpriteId:    item.SpriteId,
		CreatedAt:   item.CreatedAt,
		UpdatedAt:   item.UpdatedAt,
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

// CreateItemCategory creates a new item category.
func (ic *itemController) CreateItemCategory(ctx *gin.Context) {
	req := &dtos.CreateItemCategoryRequest{}
	if err := ctx.ShouldBindJSON(req); err != nil {
		_ = ctx.Error(errors.NewBadRequestError(err.Error()))
		return
	}

	if len(req.Name) < 3 || len(req.Name) > 32 {
		_ = ctx.Error(errors.NewBadRequestError("category name must be between 3 and 32 characters"))
		return
	}

	category, err := ic.itemCategoryService.CreateCategory(req.Name)
	if err != nil {
		if _, ok := err.(*item_errors.ItemCategoryConflict); ok {
			_ = ctx.Error(errors.NewConflictError("category name already exists"))
			return
		}
		_ = ctx.Error(err)
		return
	}

	res := &dtos.ItemCategoryResponse{
		Id:        category.Id,
		Name:      category.Name,
		CreatedAt: category.CreatedAt,
		UpdatedAt: category.UpdatedAt,
	}

	common_handlers.HandleSuccessResponse(ctx, http.StatusCreated, res)
}

// GetItemCategories retrieves all item categories.
func (ic *itemController) GetItemCategories(ctx *gin.Context) {
	categories, err := ic.itemCategoryService.GetAllCategories()
	if err != nil {
		_ = ctx.Error(err)
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

	common_handlers.HandleSuccessResponse(ctx, http.StatusOK, res)
}

// DeleteItemCategory deletes an item category by ID.
func (ic *itemController) DeleteItemCategory(ctx *gin.Context) {
	categoryIdStr := ctx.Param("id")
	categoryId, err := uuid.Parse(categoryIdStr)
	if err != nil {
		_ = ctx.Error(errors.NewBadRequestError("invalid category ID format"))
		return
	}

	if err := ic.itemCategoryService.DeleteCategory(categoryId); err != nil {
		if _, ok := err.(*item_errors.ItemCategoryNotFound); ok {
			_ = ctx.Error(errors.NewNotFoundError(err.Error()))
			return
		}
		if categoryInUse, ok := err.(*item_errors.ItemCategoryInUse); ok {
			_ = ctx.Error(errors.NewBadRequestError(categoryInUse.Error()))
			return
		}
		_ = ctx.Error(err)
		return
	}

	common_handlers.HandleSuccessResponse(ctx, http.StatusOK, gin.H{"message": "category deleted successfully"})
}
