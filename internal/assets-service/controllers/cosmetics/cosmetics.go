package cosmetics

import (
	"net/http"
	"path/filepath"

	"github.com/FeedTheRealm-org/core-service/config"
	"github.com/FeedTheRealm-org/core-service/internal/assets-service/dtos"
	assets_errors "github.com/FeedTheRealm-org/core-service/internal/assets-service/errors"
	"github.com/FeedTheRealm-org/core-service/internal/assets-service/services/cosmetics"
	"github.com/FeedTheRealm-org/core-service/internal/common_handlers"
	"github.com/FeedTheRealm-org/core-service/internal/errors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type cosmeticsController struct {
	conf             *config.Config
	cosmeticsService cosmetics.CosmeticsService
}

// NewCosmeticsController creates a new instance of CosmeticsController.
func NewCosmeticsController(conf *config.Config, cosmeticsService cosmetics.CosmeticsService) CosmeticsController {
	return &cosmeticsController{
		conf:             conf,
		cosmeticsService: cosmeticsService,
	}
}

// @Summary GetCategoriesList
// @Description Retrieves the list of existing categories UUIDs.
// @Tags assets-service
// @Produce  json
// @Success 200  {object}  dtos.CosmeticCategoryListResponse "Category list"
// @Failure 401  {object}  dtos.ErrorResponse "Invalid credentials or invalid JWT token"
// @Router /assets/cosmetics/categories [get]
func (cc *cosmeticsController) GetCategoriesList(c *gin.Context) {
	categories, err := cc.cosmeticsService.GetCategoriesList()
	if err != nil {
		_ = c.Error(err)
		return
	}

	res := &dtos.CosmeticCategoryListResponse{
		CategoryList: make([]dtos.CosmeticCategoryResponse, len(categories)),
	}
	for idx, c := range categories {
		res.CategoryList[idx] = dtos.CosmeticCategoryResponse{
			CategoryId:   c.Id,
			CategoryName: c.Name,
		}
	}

	common_handlers.HandleSuccessResponse(c, http.StatusOK, res)
}

// @Summary GetCosmeticsListByCategory
// @Description Retrieves the list of existing cosmetics UUIDs for a given category UUID.
// @Tags assets-service
// @Produce  json
// @Param category_id path string true "Category UUID"
// @Success 200  {object}  dtos.CosmeticsListResponse "Cosmetic list"
// @Failure 401  {object}  dtos.ErrorResponse "Invalid credentials or invalid JWT token"
// @Router /assets/cosmetics/categories/{category_id} [get]
func (cc *cosmeticsController) GetCosmeticsListByCategory(c *gin.Context) {
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

	cosmeticsList, err := cc.cosmeticsService.GetCosmeticsListByCategory(categoryId)
	if err != nil {
		_ = c.Error(err)
		return
	}

	res := &dtos.CosmeticsListResponse{
		CosmeticsList: make([]dtos.CosmeticResponse, len(cosmeticsList)),
	}
	for idx, cosmetic := range cosmeticsList {
		res.CosmeticsList[idx] = dtos.CosmeticResponse{
			CosmeticId:  cosmetic.Id,
			CosmeticUrl: cosmetic.Url,
		}
	}

	common_handlers.HandleSuccessResponse(c, http.StatusOK, res)
}

func (cc *cosmeticsController) GetCosmeticById(c *gin.Context) {
	cosmeticId, err := uuid.Parse(c.Param("id"))
	if err != nil {
		_ = c.Error(errors.NewBadRequestError("invalid cosmetic_id: " + err.Error()))
		return
	}

	cosmetic, err := cc.cosmeticsService.GetCosmeticById(cosmeticId)
	if err != nil {
		_ = c.Error(err)
		return
	}

	res := &dtos.CosmeticResponse{
		CosmeticId:  cosmetic.Id,
		CosmeticUrl: cosmetic.Url,
	}

	common_handlers.HandleSuccessResponse(c, http.StatusOK, res)
}

// @Summary UploadCosmeticData
// @Description Uploads a cosmetic file.
// @Tags assets-service
// @Accept  multipart/form-data
// @Produce  json
// @Param cosmetic formData file true "Cosmetic file"
// @Param category_id formData string true "Category ID"
// @Success 201  {object}  dtos.CosmeticResponse "Uploaded cosmetic"
// @Failure 400  {object}  dtos.ErrorResponse "Bad request body"
// @Failure 401  {object}  dtos.ErrorResponse "Invalid credentials or invalid JWT token"
// @Router /assets/cosmetics [put]
func (cc *cosmeticsController) UploadCosmeticData(c *gin.Context) {
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

	if reqFile.Size > cc.conf.Assets.MaxUploadSizeBytes {
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

	cosmetic, err := cc.cosmeticsService.UploadCosmeticData(categoryId, file, filepath.Ext(reqFile.Filename))
	if err != nil {
		_ = c.Error(err)
		return
	}

	res := &dtos.CosmeticResponse{
		CosmeticId:  cosmetic.Id,
		CosmeticUrl: cosmetic.Url,
	}
	common_handlers.HandleSuccessResponse(c, http.StatusCreated, res)
}

// @Summary AddCategory
// @Description Adds a new cosmetic category.
// @Tags assets-service
// @Accept  json
// @Produce  json
// @Param category body dtos.AddCosmeticCategoryRequest true "Category data"
// @Success 201  {object}  dtos.CosmeticCategoryResponse "Created category"
// @Failure 400  {object}  dtos.ErrorResponse "Bad request body"
// @Failure 401  {object}  dtos.ErrorResponse "Invalid credentials or invalid JWT token"
// @Router /assets/cosmetics/categories [post]
func (cc *cosmeticsController) AddCategory(c *gin.Context) {
	// userId, err := common_handlers.GetUserIDFromSession(ctx)
	// if err != nil {
	// 	_ = ctx.Error(errors.NewUnauthorizedError(err.Error()))
	// 	return
	// }

	req := &dtos.AddCosmeticCategoryRequest{}
	if err := c.ShouldBindJSON(req); err != nil {
		_ = c.Error(errors.NewBadRequestError(err.Error()))
		return
	}

	if len(req.CategoryName) < 3 || len(req.CategoryName) > 32 {
		_ = c.Error(errors.NewBadRequestError("category name must be between 3 and 32 characters"))
		return
	}

	category, err := cc.cosmeticsService.AddCategory(req.CategoryName)
	if err != nil {
		if _, ok := err.(*assets_errors.CategoryConflict); ok {
			_ = c.Error(errors.NewConflictError("cosmetic category name already exists"))
			return
		}
		_ = c.Error(err)
		return
	}

	res := &dtos.CosmeticCategoryResponse{
		CategoryId:   category.Id,
		CategoryName: category.Name,
	}
	common_handlers.HandleSuccessResponse(c, http.StatusCreated, res)
}
