package items

import (
	"fmt"
	"net/http"

	"github.com/FeedTheRealm-org/core-service/config"
	"github.com/FeedTheRealm-org/core-service/internal/assets-service/dtos"
	assets_errors "github.com/FeedTheRealm-org/core-service/internal/assets-service/errors"
	"github.com/FeedTheRealm-org/core-service/internal/assets-service/services/items"
	"github.com/FeedTheRealm-org/core-service/internal/common_handlers"
	"github.com/FeedTheRealm-org/core-service/internal/errors"
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

func (ic *itemController) GetItemsListByCategory(c *gin.Context) {
	worldId, err := uuid.Parse(c.Param("world_id"))
	if err != nil {
		_ = c.Error(errors.NewBadRequestError("invalid world_id: " + err.Error()))
		return
	}

	categoryId, err := uuid.Parse(c.Param("category_id"))
	if err != nil {
		_ = c.Error(errors.NewBadRequestError("invalid category_id: " + err.Error()))
		return
	}

	itemsList, err := ic.service.GetItemsListByCategory(worldId, categoryId)
	if err != nil {
		_ = c.Error(err)
		return
	}

	res := &dtos.ItemListResponse{
		Items: make([]dtos.ItemResponse, len(itemsList)),
	}
	for idx, item := range itemsList {
		res.Items[idx] = dtos.ItemResponse{
			Id:  item.Id,
			Url: item.Url,
		}
	}

	common_handlers.HandleSuccessResponse(c, http.StatusOK, res)
}

func (ic *itemController) GetItemById(c *gin.Context) {
	itemId, err := uuid.Parse(c.Param("id"))
	if err != nil {
		_ = c.Error(errors.NewBadRequestError("invalid item_id: " + err.Error()))
		return
	}

	item, err := ic.service.GetItemById(itemId)
	if err != nil {
		_ = c.Error(err)
		return
	}

	res := &dtos.ItemResponse{
		Id:  item.Id,
		Url: item.Url,
	}

	common_handlers.HandleSuccessResponse(c, http.StatusOK, res)
}

// @Summary UploadItemSprites
// @Description Uploads multiple item sprites. Each sprite must have a provided ID.
// @Tags assets-service
// @Accept multipart/form-data
// @Produce json
// @Param world_id path string true "World ID" format(uuid)
// @Param category_id path string true "Category ID" format(uuid)
// @Param ids[] formData string true "Item IDs (UUIDs), one per file"
// @Param sprites[] formData file true "Item sprite files (PNG o JPEG)"
// @Success 201 {object} dtos.ItemListResponse "Uploaded item sprites"
// @Failure 400 {object} dtos.ErrorResponse "Bad request"
// @Failure 401 {object} dtos.ErrorResponse "Invalid credentials or invalid JWT token"
// @Router /assets/sprites/items/{world_id}/{category_id} [post]
func (ic *itemController) UploadItems(c *gin.Context) {
	worldId, err := uuid.Parse(c.Param("world_id"))
	if err != nil {
		_ = c.Error(errors.NewBadRequestError("invalid world_id format"))
		return
	}

	categoryId, err := uuid.Parse(c.Param("category_id"))
	if err != nil {
		_ = c.Error(errors.NewBadRequestError("invalid category_id format"))
		return
	}

	responseSprites := make([]dtos.ItemResponse, 0)
	if len(c.Request.Form["ids[]"]) == 0 {
		_ = c.Error(errors.NewBadRequestError("must provide at least one ids[N] and sprites[N] pair"))
		return
	}

	for i, idStr := range c.Request.Form["ids[]"] {
		id, err := uuid.Parse(idStr)
		if err != nil {
			_ = c.Error(errors.NewBadRequestError(fmt.Sprintf("invalid id format for id[%d]: %s", i, err.Error())))
			return
		}

		spriteFile, err := c.FormFile(fmt.Sprintf("sprites[%d]", i))
		if err != nil {
			_ = c.Error(errors.NewBadRequestError(fmt.Sprintf("Missing sprite file for id[%d]", i)))
			return
		}

		item, err := ic.service.UploadSprite(worldId, categoryId, id, spriteFile)
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
	}

	res := &dtos.ItemListResponse{
		Items: responseSprites,
	}
	common_handlers.HandleSuccessResponse(c, http.StatusCreated, res)
}

func (ic *itemController) AddCategory(c *gin.Context) {
	req := &dtos.AddItemCategoryRequest{}
	if err := c.ShouldBindJSON(req); err != nil {
		_ = c.Error(errors.NewBadRequestError(err.Error()))
		return
	}

	if len(req.CategoryName) < 3 || len(req.CategoryName) > 32 {
		_ = c.Error(errors.NewBadRequestError("category name must be between 3 and 32 characters"))
		return
	}

	category, err := ic.service.AddCategory(req.CategoryName)
	if err != nil {
		if _, ok := err.(*assets_errors.CategoryConflict); ok {
			_ = c.Error(errors.NewConflictError("item category name already exists"))
			return
		}
		_ = c.Error(err)
		return
	}

	res := &dtos.ItemCategoryResponse{
		CategoryId:   category.Id,
		CategoryName: category.Name,
	}
	common_handlers.HandleSuccessResponse(c, http.StatusCreated, res)
}
