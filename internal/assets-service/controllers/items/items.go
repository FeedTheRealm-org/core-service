package items

import (
	"fmt"
	"mime/multipart"
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
	categoryId, err := uuid.Parse(c.Param("id"))
	if err != nil {
		_ = c.Error(errors.NewBadRequestError("invalid category_id: " + err.Error()))
		return
	}

	itemsList, err := ic.service.GetItemsListByCategory(categoryId)
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
// @Param ids[] formData string true "Item IDs (UUIDs), one per file"
// @Param sprites[] formData file true "Item sprite files (PNG o JPEG)"
// @Success 201 {object} dtos.ItemListResponse "Uploaded item sprites"
// @Failure 400 {object} dtos.ErrorResponse "Bad request"
// @Failure 401 {object} dtos.ErrorResponse "Invalid credentials or invalid JWT token"
// @Router /assets/sprites/items/{world_id} [post]
func (ic *itemController) UploadItems(c *gin.Context) {
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
		if spriteFile.Size > ic.conf.Assets.MaxUploadSizeBytes {
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

	sprites, err := ic.service.UploadSprites(worldID, ids, files)
	if err != nil {
		_ = c.Error(err)
		return
	}
	responseSprites := make([]dtos.ItemResponse, len(sprites))
	for i, sprite := range sprites {
		responseSprites[i] = dtos.ItemResponse{
			Id:        sprite.Id,
			Url:       sprite.Url,
			CreatedAt: sprite.CreatedAt,
			UpdatedAt: sprite.UpdatedAt,
		}
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
