package cosmetics

import (
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strconv"

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
// @Param        offset query int false "Pagination offset" default(0)
// @Param        limit query int false "Pagination limit" default(24)
// @Param        world_id query string false "World UUID" default(null)
// @Param        player_id query string false "Player UUID" default(null)
// @Success      200  {object}  dtos.CosmeticsListResponse
// @Failure      400  {object} dtos.ErrorResponse
// @Failure      401  {object} dtos.ErrorResponse
// @Router       /assets/cosmetics/categories/{id} [get]
func (cc *cosmeticsController) GetCosmeticsListByCategory(c *gin.Context) {
	userID, err := common_handlers.GetUserIDFromSession(c)
	if err != nil {
		_ = c.Error(errors.NewUnauthorizedError(err.Error()))
		return
	}

	categoryId, err := uuid.Parse(c.Param("id"))
	if err != nil {
		_ = c.Error(errors.NewBadRequestError("invalid category_id: " + err.Error()))
		return
	}

	offset, err := strconv.Atoi(c.DefaultQuery("offset", "0"))
	if err != nil || offset < 0 {
		_ = c.Error(errors.NewBadRequestError("offset must be a non-negative integer"))
		return
	}

	limit, err := strconv.Atoi(c.DefaultQuery("limit", "24"))
	if err != nil || limit <= 0 {
		_ = c.Error(errors.NewBadRequestError("limit must be a positive integer"))
		return
	}

	var worldId *uuid.UUID
	if raw := c.DefaultQuery("world_id", ""); raw != "" {
		*worldId, err = uuid.Parse(raw)
		if err != nil {
			_ = c.Error(errors.NewBadRequestError("invalid world_id: " + err.Error()))
			return
		}
	}

	var playerId *uuid.UUID
	if raw := c.DefaultQuery("player_id", ""); raw != "" {
		*playerId, err = uuid.Parse(raw)
		if err != nil {
			_ = c.Error(errors.NewBadRequestError("invalid player_id: " + err.Error()))
			return
		}
	}

	if *playerId != uuid.Nil && *playerId != userID {
		if err := common_handlers.IsAdminSession(c); err != nil {
			_ = c.Error(errors.NewUnauthorizedError("invalid player_id"))
			return
		}
	}

	if limit > 200 {
		limit = 200
	}

	cosmeticsList, totalCount, err := cc.cosmeticsService.GetCosmeticsListByCategory(categoryId, worldId, playerId, offset, limit)
	if err != nil {
		switch err.(type) {
		case *assets_errors.CategoryNotFound:
			_ = c.Error(errors.NewNotFoundError("category not found"))
		default:
			_ = c.Error(err)
		}
		return
	}

	res := &dtos.CosmeticsListResponse{
		CosmeticsList: make([]dtos.CosmeticResponse, len(cosmeticsList)),
		TotalCount:    totalCount,
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
		switch err.(type) {
		case *assets_errors.CosmeticNotFound:
			_ = c.Error(errors.NewNotFoundError("cosmetic not found"))
		default:
			_ = c.Error(err)
		}
		return
	}

	res := &dtos.CosmeticResponse{
		CosmeticId:  cosmetic.Id,
		CosmeticUrl: cosmetic.Url,
	}

	common_handlers.HandleSuccessResponse(c, http.StatusOK, res)
}

// GetCosmeticsListByWorld godoc
// @Summary      Get cosmetics by world
// @Description  Retrieves a list of cosmetics that belong to a specific world ID.
// @Tags         assets-service
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        world_id path string true "World UUID"
// @Param        offset query int false "Pagination offset" default(0)
// @Param        limit query int false "Pagination limit" default(24)
// @Success      200  {object}  dtos.CosmeticsListResponse
// @Failure      400  {object} dtos.ErrorResponse
// @Failure      401  {object} dtos.ErrorResponse
// @Router       /assets/cosmetics/worlds/{world_id} [get]
func (cc *cosmeticsController) GetCosmeticsListByWorld(c *gin.Context) {
	_, err := common_handlers.GetUserIDFromSession(c)
	if err != nil {
		_ = c.Error(errors.NewUnauthorizedError(err.Error()))
		return
	}

	worldId, err := uuid.Parse(c.Param("world_id"))
	if err != nil {
		_ = c.Error(errors.NewBadRequestError("invalid world_id: " + err.Error()))
		return
	}

	offset, err := strconv.Atoi(c.DefaultQuery("offset", "0"))
	if err != nil || offset < 0 {
		_ = c.Error(errors.NewBadRequestError("offset must be a non-negative integer"))
		return
	}

	limit, err := strconv.Atoi(c.DefaultQuery("limit", "24"))
	if err != nil || limit <= 0 {
		_ = c.Error(errors.NewBadRequestError("limit must be a positive integer"))
		return
	}

	if limit > 200 {
		limit = 200
	}

	cosmeticsList, totalCount, err := cc.cosmeticsService.GetCosmeticsListByWorld(worldId, offset, limit)
	if err != nil {
		_ = c.Error(err)
		return
	}

	res := &dtos.CosmeticsListResponse{
		CosmeticsList: make([]dtos.CosmeticResponse, len(cosmeticsList)),
		TotalCount:    totalCount,
	}
	for idx, cosmetic := range cosmeticsList {
		res.CosmeticsList[idx] = dtos.CosmeticResponse{
			CosmeticId:  cosmetic.Id,
			CosmeticUrl: cosmetic.Url,
		}
	}

	common_handlers.HandleSuccessResponse(c, http.StatusOK, res)
}

func (cc *cosmeticsController) checkWorldOwnership(c *gin.Context, worldId uuid.UUID, userId uuid.UUID) error {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("http://127.0.0.1:%d/world/%s", cc.conf.Server.Port, worldId.String()), nil)
	if err != nil {
		return errors.NewInternalServerError("failed to check world ownership")
	}
	req.Header.Set("Authorization", c.GetHeader("Authorization"))
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return errors.NewInternalServerError("failed to check world ownership")
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return errors.NewBadRequestError("invalid world_id or world not found")
	}

	var envelope struct {
		Data struct {
			UserId string `json:"user_id"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&envelope); err != nil {
		return errors.NewInternalServerError("failed to parse world data")
	}
	if envelope.Data.UserId != userId.String() {
		return errors.NewUnauthorizedError("user does not own this world")
	}
	return nil
}

// UploadCosmeticData godoc
// @Summary      Upload cosmetic data
// @Description  Upload a cosmetic form-data payload containing category ID and cosmetic file.
// @Tags         assets-service
// @Security     BearerAuth
// @Accept       multipart/form-data
// @Produce      json
// @Param        category_id path string true "Category UUID"
// @Param        world_id formData string true "World UUID"
// @Param        price formData number false "Cosmetic price" default(0.00)
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

	categoryId, err := uuid.Parse(c.Param("category_id"))
	if err != nil {
		_ = c.Error(errors.NewBadRequestError("invalid category_id: " + err.Error()))
		return
	}

	worldId, err := uuid.Parse(c.PostForm("world_id"))
	if err != nil {
		_ = c.Error(errors.NewBadRequestError("invalid world_id: " + err.Error()))
		return
	}

	price, err := strconv.ParseFloat(c.PostForm("price"), 64)
	if err != nil {
		_ = c.Error(errors.NewBadRequestError("invalid price: " + err.Error()))
		return
	}

	if worldId == uuid.Nil {
		if err := common_handlers.IsAdminSession(c); err != nil {
			_ = c.Error(errors.NewUnauthorizedError("invalid world_id"))
			return
		}
	} else {
		if err := common_handlers.IsAdminSession(c); err != nil {
			if err := cc.checkWorldOwnership(c, worldId, userId); err != nil {
				_ = c.Error(err)
				return
			}
		}
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

	cosmetic, err := cc.cosmeticsService.UploadCosmeticData(categoryId, worldId, price, file, filepath.Ext(reqFile.Filename), userId)
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

// UploadCosmeticByID godoc
// @Summary      Link existing cosmetic by ID
// @Description  Creates a cosmetic entry in the target category using the URL from an existing sprite ID.
// @Tags         assets-service
// @Security     BearerAuth
// @Produce      json
// @Param        category_id path string true "Category UUID"
// @Param        sprite_id path string true "Existing Sprite UUID"
// @Param				 world_id formData string true "World UUID"
// @Param				 price formData number false "Cosmetic price" default(0.00)
// @Success      201  {object}  dtos.CosmeticResponse
// @Failure      400  {object} dtos.ErrorResponse
// @Failure      401  {object} dtos.ErrorResponse
// @Router       /assets/cosmetics/categories/{id}/sprites/{sprite_id} [put]
func (cc *cosmeticsController) UploadCosmeticByID(c *gin.Context) {
	userId, err := common_handlers.GetUserIDFromSession(c)
	if err != nil {
		_ = c.Error(errors.NewUnauthorizedError(err.Error()))
		return
	}

	categoryId, err := uuid.Parse(c.Param("category_id"))
	if err != nil {
		_ = c.Error(errors.NewBadRequestError("invalid category_id: " + err.Error()))
		return
	}

	worldId, err := uuid.Parse(c.PostForm("world_id"))
	if err != nil {
		_ = c.Error(errors.NewBadRequestError("invalid world_id: " + err.Error()))
		return
	}

	if worldId == uuid.Nil {
		if err := common_handlers.IsAdminSession(c); err != nil {
			_ = c.Error(errors.NewUnauthorizedError("invalid world_id"))
			return
		}
	} else {
		if err := common_handlers.IsAdminSession(c); err != nil {
			if err := cc.checkWorldOwnership(c, worldId, userId); err != nil {
				_ = c.Error(err)
				return
			}
		}
	}

	price, err := strconv.ParseFloat(c.PostForm("price"), 64)
	if err != nil {
		_ = c.Error(errors.NewBadRequestError("invalid price: " + err.Error()))
		return
	}

	spriteId, err := uuid.Parse(c.Param("sprite_id"))
	if err != nil {
		_ = c.Error(errors.NewBadRequestError("invalid sprite_id: " + err.Error()))
		return
	}

	var cosmeticFile multipart.File
	var ext string
	reqFile, err := c.FormFile("sprite")
	if err != nil && err != http.ErrMissingFile {
		_ = c.Error(errors.NewBadRequestError("failed to get sprite file from request: " + err.Error()))
		return
	}
	if reqFile != nil {
		if reqFile.Size > cc.conf.Assets.MaxUploadSizeBytes {
			_ = c.Error(errors.NewBadRequestError("file size exceeds the limit"))
			return
		}
		contentType := reqFile.Header.Get("Content-Type")
		if contentType != "image/png" && contentType != "image/jpeg" {
			_ = c.Error(errors.NewBadRequestError("file must be PNG or JPEG format"))
			return
		}
		cosmeticFile, err = reqFile.Open()
		if err != nil {
			_ = c.Error(errors.NewBadRequestError("failed to open sprite file: " + err.Error()))
			return
		}
		defer func() { _ = cosmeticFile.Close() }()
		ext = filepath.Ext(reqFile.Filename)
	}

	cosmetic, err := cc.cosmeticsService.UploadCosmeticByID(categoryId, worldId, price, spriteId, userId, cosmeticFile, ext)
	if err != nil {
		switch err.(type) {
		case *assets_errors.CategoryNotFound:
			_ = c.Error(errors.NewNotFoundError("category not found"))
		case *assets_errors.CosmeticNotFound:
			_ = c.Error(errors.NewNotFoundError("cosmetic not found"))
		default:
			_ = c.Error(err)
		}
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

// GetCosmeticByIdInternal godoc
// @Summary      Get cosmetic by ID (Internal)
// @Description  Retrieves a single cosmetic item by its ID. Intended for internal service communication.
// @Tags         assets-service-internal
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        cosmetic_id path string true "Cosmetic UUID"
// @Success      200  {object}  dtos.InternalCosmeticResponse
// @Failure      400  {object} dtos.ErrorResponse
// @Failure      404  {object} dtos.ErrorResponse
// @Router       /assets/internal/cosmetics/{cosmetic_id} [get]
func (cc *cosmeticsController) GetCosmeticByIdInternal(c *gin.Context) {
	cosmeticId, err := uuid.Parse(c.Param("cosmetic_id"))
	if err != nil {
		_ = c.Error(errors.NewBadRequestError("invalid cosmetic_id: " + err.Error()))
		return
	}

	cosmetic, err := cc.cosmeticsService.GetCosmeticById(cosmeticId)
	if err != nil {
		switch err.(type) {
		case *assets_errors.CosmeticNotFound:
			_ = c.Error(errors.NewNotFoundError("cosmetic not found"))
		default:
			_ = c.Error(err)
		}
		return
	}

	res := &dtos.InternalCosmeticResponse{
		CosmeticId:    cosmetic.Id,
		CosmeticPrice: cosmetic.Price,
	}

	common_handlers.HandleSuccessResponse(c, http.StatusOK, res)
}

// PurshaseCosmeticForUserInternal godoc
// @Summary      Purchase cosmetic for user (Internal)
// @Description  Records a cosmetic purchase for a specific user. Intended for internal service communication.
// @Tags         assets-service-internal
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        user_id path string true "User UUID"
// @Param        request body dtos.InternalPurchaseCosmeticForUserRequest true "Purchase Details"
// @Success      201  {object}  dtos.InternalPurchaseCosmeticForUserResponse
// @Failure      400  {object} dtos.ErrorResponse
// @Failure      409  {object} dtos.ErrorResponse
// @Router       /assets/internal/users/{user_id}/cosmetics [post]
func (cc *cosmeticsController) PurshaseCosmeticForUserInternal(c *gin.Context) {
	userId, err := uuid.Parse(c.Param("user_id"))
	if err != nil {
		_ = c.Error(errors.NewBadRequestError("invalid user_id: " + err.Error()))
		return
	}

	req := &dtos.InternalPurchaseCosmeticForUserRequest{}
	if err := c.ShouldBindJSON(req); err != nil {
		_ = c.Error(errors.NewBadRequestError(err.Error()))
		return
	}

	err = cc.cosmeticsService.PurchaseCosmeticForUserInternal(userId, req.CosmeticId)
	if err != nil {
		if _, ok := err.(*assets_errors.CosmeticsWasPurchasedBefore); ok {
			_ = c.Error(errors.NewConflictError("cosmetic was already purchased by the user"))
			return
		}
		_ = c.Error(err)
		return
	}

	res := &dtos.InternalPurchaseCosmeticForUserResponse{
		UserId:     userId,
		CosmeticId: req.CosmeticId,
	}
	common_handlers.HandleSuccessResponse(c, http.StatusCreated, res)
}
