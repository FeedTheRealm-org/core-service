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

// GetCategoriesList godoc
// @Summary      Get cosmetics categories
// @Description  Retrieves a list of all cosmetic categories.
// @Tags         assets-service
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Success      200  {object}  dtos.CosmeticCategoryListResponse
// @Failure      401  {object} dtos.ErrorResponse
// @Router       /assets/cosmetics/categories [get]
func (cc *cosmeticsController) GetCategoriesList(c *gin.Context) {
	_, err := common_handlers.GetUserIDFromSession(c)
	if err != nil {
		_ = c.Error(errors.NewUnauthorizedError(err.Error()))
		return
	}

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

// GetCosmeticsListByCategory godoc
// @Summary      Get cosmetics by category
// @Description  Retrieves a list of cosmetics that belong to a specific category ID.
// @Tags         assets-service
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id path string true "Category UUID"
// @Success      200  {object}  dtos.CosmeticsListResponse
// @Failure      400  {object} dtos.ErrorResponse
// @Failure      401  {object} dtos.ErrorResponse
// @Router       /assets/cosmetics/categories/{id} [get]
func (cc *cosmeticsController) GetCosmeticsListByCategory(c *gin.Context) {
	_, err := common_handlers.GetUserIDFromSession(c)
	if err != nil {
		_ = c.Error(errors.NewUnauthorizedError(err.Error()))
		return
	}

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

// GetCosmeticById godoc
// @Summary      Get cosmetic by ID
// @Description  Retrieves a single cosmetic item by its ID.
// @Tags         assets-service
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id path string true "Cosmetic UUID"
// @Success      200  {object}  dtos.CosmeticResponse
// @Failure      400  {object} dtos.ErrorResponse
// @Failure      401  {object} dtos.ErrorResponse
// @Router       /assets/cosmetics/{id} [get]
func (cc *cosmeticsController) GetCosmeticById(c *gin.Context) {
	_, err := common_handlers.GetUserIDFromSession(c)
	if err != nil {
		_ = c.Error(errors.NewUnauthorizedError(err.Error()))
		return
	}

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

// UploadCosmeticData godoc
// @Summary      Upload cosmetic data
// @Description  Upload a cosmetic form-data payload containing category ID and cosmetic file.
// @Tags         assets-service
// @Security     BearerAuth
// @Accept       multipart/form-data
// @Produce      json
// @Param        category_id formData string true "Category UUID"
// @Param        sprite formData file true "Cosmetic File"
// @Success      201  {object}  dtos.CosmeticResponse
// @Failure      400  {object} dtos.ErrorResponse
// @Failure      401  {object} dtos.ErrorResponse
// @Router       /assets/cosmetics/categories/{id} [put]
func (cc *cosmeticsController) UploadCosmeticData(c *gin.Context) {
	userId, err := common_handlers.GetUserIDFromSession(c)
	if err != nil {
		_ = c.Error(errors.NewUnauthorizedError(err.Error()))
		return
	}

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

	cosmetic, err := cc.cosmeticsService.UploadCosmeticData(categoryId, file, filepath.Ext(reqFile.Filename), userId)
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

// DeleteCosmetic godoc
// @Summary      Delete cosmetic
// @Description  Delete a specific cosmetic by ID (requires ownership)
// @Tags         assets-service
// @Security     BearerAuth
// @Produce      json
// @Param        id path string true "Cosmetic UUID"
// @Success      204  {object}  dtos.CosmeticResponse
// @Failure      400  {object} dtos.ErrorResponse
// @Failure      401  {object} dtos.ErrorResponse
// @Router       /assets/cosmetics/{id} [delete]
func (cc *cosmeticsController) DeleteCosmetic(c *gin.Context) {
	userId, err := common_handlers.GetUserIDFromSession(c)
	if err != nil {
		_ = c.Error(errors.NewUnauthorizedError(err.Error()))
		return
	}

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

	if cosmetic == nil {
		_ = c.Error(errors.NewBadRequestError("cosmetic not found"))
		return
	}

	if cosmetic.CreatedBy != userId {
		_ = c.Error(errors.NewUnauthorizedError("user is not authorized to delete this cosmetic"))
		return
	}

	err = cc.cosmeticsService.DeleteCosmetic(cosmeticId)
	if err != nil {
		_ = c.Error(err)
		return
	}

	res := &dtos.CosmeticResponse{
		CosmeticId: cosmeticId,
	}
	common_handlers.HandleSuccessResponse(c, http.StatusNoContent, res)
}

// AddCategory godoc
// @Summary      Adds a new cosmetic category
// @Description  Creates a new cosmetic category globally.
// @Tags         assets-service
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        category body dtos.AddCosmeticCategoryRequest true "Category data"
// @Success      201  {object}  dtos.CosmeticCategoryResponse
// @Failure      400  {object} dtos.ErrorResponse
// @Failure      401  {object} dtos.ErrorResponse
// @Failure      409  {object} dtos.ErrorResponse
// @Router       /assets/cosmetics/categories [post]
func (cc *cosmeticsController) AddCategory(c *gin.Context) {
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
