package sprites

import (
	"net/http"
	"os"
	"path/filepath"

	"github.com/FeedTheRealm-org/core-service/config"
	"github.com/FeedTheRealm-org/core-service/internal/assets-service/dtos"
	assets_errors "github.com/FeedTheRealm-org/core-service/internal/assets-service/errors"
	"github.com/FeedTheRealm-org/core-service/internal/assets-service/services/sprites"
	"github.com/FeedTheRealm-org/core-service/internal/common_handlers"
	"github.com/FeedTheRealm-org/core-service/internal/errors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type spritesController struct {
	conf           *config.Config
	spritesService sprites.SpritesService
}

// NewSpritesController creates a new instance of SpritesController.
func NewSpritesController(conf *config.Config, spritesService sprites.SpritesService) SpritesController {
	return &spritesController{
		conf:           conf,
		spritesService: spritesService,
	}
}

// @Summary GetCategoriesList
// @Description Retrieves the list of existing categories UUIDs.
// @Tags assets-service
// @Produce  json
// @Success 200  {object}  dtos.SpriteCategoryListResponse "Category list"
// @Failure 401  {object}  dtos.ErrorResponse "Invalid credentials or invalid JWT token"
// @Router /assets/sprites/categories [get]
func (sc *spritesController) GetCategoriesList(c *gin.Context) {
	// userId, err := common_handlers.GetUserIDFromSession(ctx)
	// if err != nil {
	// 	_ = ctx.Error(errors.NewUnauthorizedError(err.Error()))
	// 	return
	// }

	categories, err := sc.spritesService.GetCategoriesList()
	if err != nil {
		_ = c.Error(err)
		return
	}

	res := &dtos.SpriteCategoryListResponse{
		CategoryList: make([]dtos.SpriteCategoryResponse, len(categories)),
	}
	for idx, c := range categories {
		res.CategoryList[idx] = dtos.SpriteCategoryResponse{
			CategoryId:   c.Id,
			CategoryName: c.Name,
		}
	}

	common_handlers.HandleSuccessResponse(c, http.StatusOK, res)
}

// @Summary GetSpritesListByCategory
// @Description Retrieves the list of existing sprites UUIDs for a given category UUID.
// @Tags assets-service
// @Produce  json
// @Param category_id path string true "Category UUID"
// @Success 200  {object}  dtos.SpritesListResponse "Sprite list"
// @Failure 401  {object}  dtos.ErrorResponse "Invalid credentials or invalid JWT token"
// @Router /assets/sprites/categories/{category_id} [get]
func (sc *spritesController) GetSpritesListByCategory(c *gin.Context) {
	// userId, err := common_handlers.GetUserIDFromSession(ctx)
	// if err != nil {
	// 	_ = ctx.Error(errors.NewUnauthorizedError(err.Error()))
	// 	return
	// }

	categoryId, err := uuid.Parse(c.Param("id"))
	if err != nil {
		_ = c.Error(errors.NewBadRequestError("invalid category_id: " + err.Error()))
		return
	}

	spritesList, err := sc.spritesService.GetSpritesListByCategory(categoryId)
	if err != nil {
		_ = c.Error(err)
		return
	}

	res := &dtos.SpritesListResponse{
		SpritesList: make([]dtos.SpriteResponse, len(spritesList)),
	}
	for idx, sprite := range spritesList {
		res.SpritesList[idx] = dtos.SpriteResponse{
			SpriteId:  sprite.Id,
			SpriteUrl: sprite.Url,
		}
	}

	common_handlers.HandleSuccessResponse(c, http.StatusOK, res)
}

// @Summary DownloadSpriteData
// @Description Downloads the sprite file for a given sprite UUID.
// @Tags assets-service
// @Produce  octet-stream
// @Param sprite_id path string true "Sprite UUID"
// @Success 200  {file}  byte "Sprite file"
// @Failure 401  {object}  dtos.ErrorResponse "Invalid credentials or invalid JWT token"
// @Router /assets/sprites/{sprite_id} [get]
func (sc *spritesController) DownloadSpriteData(c *gin.Context) {
	// userId, err := common_handlers.GetUserIDFromSession(ctx)
	// if err != nil {
	// 	_ = ctx.Error(errors.NewUnauthorizedError(err.Error()))
	// 	return
	// }

	spriteIdStr := c.Param("id")
	spriteId, err := uuid.Parse(spriteIdStr)
	if err != nil {
		_ = c.Error(errors.NewBadRequestError("invalid sprite_id format"))
		return
	}

	filePath, err := sc.spritesService.GetSpriteUrl(spriteId)
	if err != nil {
		if _, ok := err.(*assets_errors.SpriteNotFound); ok {
			_ = c.Error(errors.NewNotFoundError("sprite not found"))
			return
		}
		_ = c.Error(err)
		return
	}

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		_ = c.Error(errors.NewNotFoundError("sprite file not found on disk"))
		return
	}

	c.File(filePath)
}

// @Summary AddCategory
// @Description Adds a new sprite category.
// @Tags assets-service
// @Accept  json
// @Produce  json
// @Param category body dtos.AddSpriteCategoryRequest true "Category data"
// @Success 201  {object}  dtos.SpriteCategoryResponse "Created category"
// @Failure 400  {object}  dtos.ErrorResponse "Bad request body"
// @Failure 401  {object}  dtos.ErrorResponse "Invalid credentials or invalid JWT token"
// @Router /assets/sprites/categories [post]
func (sc *spritesController) AddCategory(c *gin.Context) {
	// userId, err := common_handlers.GetUserIDFromSession(ctx)
	// if err != nil {
	// 	_ = ctx.Error(errors.NewUnauthorizedError(err.Error()))
	// 	return
	// }

	req := &dtos.AddSpriteCategoryRequest{}
	if err := c.ShouldBindJSON(req); err != nil {
		_ = c.Error(errors.NewBadRequestError(err.Error()))
		return
	}

	if len(req.CategoryName) < 3 || len(req.CategoryName) > 32 {
		_ = c.Error(errors.NewBadRequestError("category name must be between 3 and 32 characters"))
		return
	}

	category, err := sc.spritesService.AddCategory(req.CategoryName)
	if err != nil {
		if _, ok := err.(*assets_errors.CategoryConflict); ok {
			_ = c.Error(errors.NewConflictError("sprite category name already exists"))
			return
		}
		_ = c.Error(err)
		return
	}

	res := &dtos.SpriteCategoryResponse{
		CategoryId:   category.Id,
		CategoryName: category.Name,
	}
	common_handlers.HandleSuccessResponse(c, http.StatusCreated, res)
}

// @Summary UploadSpriteData
// @Description Uploads a sprite file.
// @Tags assets-service
// @Accept  multipart/form-data
// @Produce  json
// @Param sprite formData file true "Sprite file"
// @Param category_id formData string true "Category ID"
// @Success 201  {object}  dtos.SpriteResponse "Uploaded sprite"
// @Failure 400  {object}  dtos.ErrorResponse "Bad request body"
// @Failure 401  {object}  dtos.ErrorResponse "Invalid credentials or invalid JWT token"
// @Router /assets/sprites [put]
func (sc *spritesController) UploadSpriteData(c *gin.Context) {
	// userId, err := common_handlers.GetUserIDFromSession(ctx)
	// if err != nil {
	// 	_ = ctx.Error(errors.NewUnauthorizedError(err.Error()))
	// 	return
	// }

	categoryId, err := uuid.Parse(c.PostForm("category_id"))
	if err != nil {
		_ = c.Error(errors.NewBadRequestError("invalid category_id: " + err.Error()))
		return
	}

	reqFile, err := c.FormFile("sprite")
	if err != nil {
		_ = c.Error(errors.NewBadRequestError("failed to get sprite file from request: " + err.Error()))
		return
	}

	if reqFile.Size > sc.conf.Assets.MaxUploadSizeBytes {
		_ = c.Error(errors.NewBadRequestError("file size exceeds the limit"))
		return
	}

	contentType := reqFile.Header.Get("Content-Type")
	if contentType != "image/png" && contentType != "image/jpeg" {
		_ = c.Error(errors.NewBadRequestError("file must be PNG or JPEG format"))
		return
	}

	file, err := reqFile.Open()
	if err != nil {
		_ = c.Error(errors.NewBadRequestError("failed to open sprite file: " + err.Error()))
		return
	}
	defer func() {
		_ = file.Close()
	}()

	sprite, err := sc.spritesService.UploadSpriteData(categoryId, file, filepath.Ext(reqFile.Filename))
	if err != nil {
		_ = c.Error(err)
		return
	}

	res := &dtos.SpriteResponse{
		SpriteId:  sprite.Id,
		SpriteUrl: sprite.Url,
	}
	common_handlers.HandleSuccessResponse(c, http.StatusCreated, res)
}
