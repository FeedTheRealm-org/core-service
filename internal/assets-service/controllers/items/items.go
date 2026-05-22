package items

import (
	"fmt"
	"net/http"

	"github.com/FeedTheRealm-org/core-service/config"
	"github.com/FeedTheRealm-org/core-service/internal/assets-service/dtos"
	assets_errors "github.com/FeedTheRealm-org/core-service/internal/assets-service/errors"
	"github.com/FeedTheRealm-org/core-service/internal/assets-service/services/items"
	"github.com/FeedTheRealm-org/core-service/internal/common_handlers"
	internalErrors "github.com/FeedTheRealm-org/core-service/internal/errors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type itemController struct {
	conf    *config.Config
	service items.ItemService
}

// NewItemController creates a new instance of ItemController.
func NewItemController(conf *config.Config, service items.ItemService) ItemController {
	return &itemController{
		conf:    conf,
		service: service,
	}
}

// GetItemsListByWorld godoc
// @Summary      Get items by world
// @Description  Retrieves an items list specific to a world ID.
// @Tags         assets-service
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        world_id path string true "World UUID"
// @Success      200  {object}  dtos.ItemListResponse
// @Failure      400  {object} dtos.ErrorResponse
// @Failure      401  {object} dtos.ErrorResponse
// @Router       /assets/items/world/{world_id} [get]
func (ic *itemController) GetItemsListByWorld(c *gin.Context) {
	_, err := common_handlers.GetUserIDFromSession(c)
	if err != nil {
		_ = c.Error(internalErrors.NewUnauthorizedError(err.Error()))
		return
	}

	worldId, err := uuid.Parse(c.Param("world_id"))
	if err != nil {
		_ = c.Error(internalErrors.NewBadRequestError("invalid world_id: " + err.Error()))
		return
	}

	itemsList, err := ic.service.GetItemsListByWorld(worldId)
	if err != nil {
		_ = c.Error(err)
		return
	}

	res := &dtos.ItemListResponse{
		Items: make([]dtos.ItemResponse, len(itemsList)),
	}
	for idx, item := range itemsList {
		res.Items[idx] = dtos.ItemResponse{
			Id:        item.Id,
			Url:       item.Url,
			CreatedAt: item.CreatedAt,
			UpdatedAt: item.UpdatedAt,
		}
	}

	common_handlers.HandleSuccessResponse(c, http.StatusOK, res)
}

// GetItemById godoc
// @Summary      Get item by ID
// @Description  Retrieves a single item item by its ID.
// @Tags         assets-service
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id path string true "Item UUID"
// @Success      200  {object}  dtos.ItemResponse
// @Failure      400  {object} dtos.ErrorResponse
// @Failure      401  {object} dtos.ErrorResponse
// @Router       /assets/items/{id} [get]
func (ic *itemController) GetItemById(c *gin.Context) {
	_, err := common_handlers.GetUserIDFromSession(c)
	if err != nil {
		_ = c.Error(internalErrors.NewUnauthorizedError(err.Error()))
		return
	}

	itemId, err := uuid.Parse(c.Param("id"))
	if err != nil {
		_ = c.Error(internalErrors.NewBadRequestError("invalid item_id: " + err.Error()))
		return
	}

	item, err := ic.service.GetItemById(itemId)
	if err != nil {
		switch err.(type) {
		case *assets_errors.ItemSpriteNotFound:
			_ = c.Error(internalErrors.NewNotFoundError("item not found"))
		default:
			_ = c.Error(err)
		}
		return
	}

	res := &dtos.ItemResponse{
		Id:        item.Id,
		Url:       item.Url,
		CreatedAt: item.CreatedAt,
		UpdatedAt: item.UpdatedAt,
	}

	common_handlers.HandleSuccessResponse(c, http.StatusOK, res)
}

// UploadItems godoc
// @Summary      Upload items and sprites
// @Description  Upload item IDs alongside sprite files mapping to a world.
// @Tags         assets-service
// @Security     BearerAuth
// @Accept       multipart/form-data
// @Produce      json
// @Param        world_id path string true "World UUID"
// @Param        ids formData []string true "Array of exact item IDs"
// @Param        sprites formData file true "Muti-part chunk array of files"
// @Success      201  {object}  dtos.ItemListResponse
// @Failure      400  {object} dtos.ErrorResponse
// @Failure      401  {object} dtos.ErrorResponse
// @Router       /assets/items/world/{world_id} [put]
func (ic *itemController) UploadItems(c *gin.Context) {
	userId, err := common_handlers.GetUserIDFromSession(c)
	if err != nil {
		_ = c.Error(internalErrors.NewUnauthorizedError(err.Error()))
		return
	}

	worldId, err := uuid.Parse(c.Param("world_id"))
	if err != nil {
		_ = c.Error(internalErrors.NewBadRequestError("invalid world_id format"))
		return
	}

	responseSprites := make([]dtos.ItemResponse, 0)

	i := 0
	for {
		idKey := fmt.Sprintf("ids[%d]", i)
		idVal := c.PostForm(idKey)

		if idVal == "" {
			break
		}

		id, err := uuid.Parse(idVal)
		if err != nil {
			_ = c.Error(internalErrors.NewBadRequestError(fmt.Sprintf("invalid id format for ids[%d]: %s", i, err.Error())))
			return
		}

		spriteFile, err := c.FormFile(fmt.Sprintf("sprites[%d]", i))
		if err != nil {
			_ = c.Error(internalErrors.NewBadRequestError(fmt.Sprintf("Missing sprite file for ids[%d]", i)))
			return
		}

		item, err := ic.service.UploadSprite(worldId, id, spriteFile, userId)
		if err != nil {
			_ = c.Error(err)
			return
		}

		responseSprites = append(responseSprites, dtos.ItemResponse{
			Id:        item.Id,
			Url:       item.Url,
			CreatedAt: item.CreatedAt,
			UpdatedAt: item.UpdatedAt,
		})

		i++
	}

	res := &dtos.ItemListResponse{
		Items: responseSprites,
	}
	common_handlers.HandleSuccessResponse(c, http.StatusCreated, res)
}

// DeleteItem godoc
// @Summary      Delete item
// @Description  Delete a specific item by ID (requires ownership)
// @Tags         assets-service
// @Security     BearerAuth
// @Produce      json
// @Param        id path string true "Item UUID"
// @Success      204  {object}  dtos.ItemResponse
// @Failure      400  {object} dtos.ErrorResponse
// @Failure      401  {object} dtos.ErrorResponse
// @Router       /assets/items/{id} [delete]
func (ic *itemController) DeleteItem(c *gin.Context) {
	userId, err := common_handlers.GetUserIDFromSession(c)
	if err != nil {
		_ = c.Error(internalErrors.NewUnauthorizedError(err.Error()))
		return
	}

	itemId, err := uuid.Parse(c.Param("id"))
	if err != nil {
		_ = c.Error(internalErrors.NewBadRequestError("invalid item_id format"))
		return
	}

	item, err := ic.service.GetItemById(itemId)
	if err != nil {
		_ = c.Error(err)
		return
	}

	if item == nil {
		_ = c.Error(internalErrors.NewBadRequestError("item not found"))
		return
	}

	if item.CreatedBy != userId {
		_ = c.Error(internalErrors.NewUnauthorizedError("user is not authorized to delete this item"))
		return
	}

	if err := ic.service.DeleteItem(itemId); err != nil {
		_ = c.Error(err)
		return
	}

	res := &dtos.ItemResponse{
		Id: itemId,
	}

	common_handlers.HandleSuccessResponse(c, http.StatusOK, res)
}
