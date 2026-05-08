package materials

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/FeedTheRealm-org/core-service/config"
	"github.com/FeedTheRealm-org/core-service/internal/assets-service/dtos"
	assets_errors "github.com/FeedTheRealm-org/core-service/internal/assets-service/errors"
	"github.com/FeedTheRealm-org/core-service/internal/assets-service/models"
	"github.com/FeedTheRealm-org/core-service/internal/assets-service/services/materials"
	"github.com/FeedTheRealm-org/core-service/internal/common_handlers"
	"github.com/FeedTheRealm-org/core-service/internal/errors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type materialsController struct {
	conf    *config.Config
	service materials.MaterialsService
}

// NewMaterialsController creates a new instance of MaterialsController.
func NewMaterialsController(conf *config.Config, service materials.MaterialsService) MaterialsController {
	return &materialsController{
		conf:    conf,
		service: service,
	}
}

// GetMaterialsList godoc
// @Summary      Get materials list
// @Description  Retrieves a materials list. If world_id query param is provided, retrieves materials specific to that world along with default materials. Otherwise retrieves only default materials.
// @Tags         assets-service
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        world_id query string false "World UUID"
// @Param        material_type query int false "Material type filter (0 for GroundMaterial, 1 for SkyBoxMaterial)"
// @Param        offset query int false "Pagination offset" default(0)
// @Param        limit query int false "Pagination limit" default(24)
// @Success      200  {array}   dtos.MaterialResponse
// @Failure      400  {object} dtos.ErrorResponse
// @Failure      401  {object} dtos.ErrorResponse
// @Router       /assets/materials [get]
func (mc *materialsController) GetMaterialsList(c *gin.Context) {
	_, err := common_handlers.GetUserIDFromSession(c)
	if err != nil {
		_ = c.Error(errors.NewUnauthorizedError(err.Error()))
		return
	}

	var worldId uuid.UUID
	worldIdStr := c.Query("world_id")
	if worldIdStr != "" {
		worldId, err = uuid.Parse(worldIdStr)
		if err != nil {
			_ = c.Error(errors.NewBadRequestError("invalid world_id: " + err.Error()))
			return
		}
	} else {
		worldId = uuid.Nil
	}

	var materialType models.MaterialType
	materialTypeStr := c.Query("material_type")
	if materialTypeStr != "" {
		materialTypeInt, err := strconv.Atoi(materialTypeStr)
		if err != nil {
			_ = c.Error(errors.NewBadRequestError("invalid material_type: " + err.Error()))
			return
		}

		materialType, err = models.ParseMaterialType(materialTypeInt)
		if err != nil {
			_ = c.Error(errors.NewBadRequestError("invalid material_type: " + err.Error()))
			return
		}
	}

	offset, err := strconv.Atoi(c.DefaultQuery("offset", "0"))
	if err != nil || offset < 0 {
		_ = c.Error(errors.NewBadRequestError("offset must be a non-negative integer"))
		return
	}

	limit, err := strconv.Atoi(c.DefaultQuery("limit", "200"))
	if err != nil || limit <= 0 {
		_ = c.Error(errors.NewBadRequestError("limit must be a positive integer"))
		return
	}

	if limit > 200 {
		limit = 200
	}

	if materialTypeStr != "" {
		materialsList, err := mc.service.GetMaterialsListByWorldAndType(worldId, materialType, offset, limit)
		if err != nil {
			_ = c.Error(err)
			return
		}

		res := make([]dtos.MaterialResponse, len(materialsList))
		for idx, material := range materialsList {
			res[idx] = dtos.MaterialResponse{
				ID:        material.ID,
				Name:      material.Name,
				URL:       material.URL,
				WorldID:   material.WorldID,
				CreatedAt: material.CreatedAt.String(),
				UpdatedAt: material.UpdatedAt.String(),
			}
		}

		common_handlers.HandleSuccessResponse(c, http.StatusOK, res)
		return
	}

	materialsList, err := mc.service.GetMaterialsListByWorld(worldId, offset, limit)
	if err != nil {
		_ = c.Error(err)
		return
	}

	res := make([]dtos.MaterialResponse, len(materialsList))
	for idx, material := range materialsList {
		res[idx] = dtos.MaterialResponse{
			ID:        material.ID,
			Name:      material.Name,
			URL:       material.URL,
			WorldID:   material.WorldID,
			CreatedAt: material.CreatedAt.String(),
			UpdatedAt: material.UpdatedAt.String(),
		}
	}

	common_handlers.HandleSuccessResponse(c, http.StatusOK, res)
}

// GetMaterialByID godoc
// @Summary      Get material by ID
// @Description  Retrieves a single material by its ID.
// @Tags         assets-service
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id path string true "Material UUID"
// @Success      200  {object}  dtos.MaterialResponse
// @Failure      400  {object} dtos.ErrorResponse
// @Failure      401  {object} dtos.ErrorResponse
// @Router       /assets/materials/{id} [get]
func (mc *materialsController) GetMaterialByID(c *gin.Context) {
	_, err := common_handlers.GetUserIDFromSession(c)
	if err != nil {
		_ = c.Error(errors.NewUnauthorizedError(err.Error()))
		return
	}

	materialId, err := uuid.Parse(c.Param("id"))
	if err != nil {
		_ = c.Error(errors.NewBadRequestError("invalid material_id: " + err.Error()))
		return
	}

	material, err := mc.service.GetMaterialByID(materialId)
	if err != nil {
		switch err.(type) {
		case *assets_errors.MaterialNotFound:
			_ = c.Error(errors.NewNotFoundError("material not found"))
		default:
			_ = c.Error(err)
		}
		return
	}

	res := &dtos.MaterialResponse{
		ID:        material.ID,
		Name:      material.Name,
		URL:       material.URL,
		WorldID:   material.WorldID,
		CreatedAt: material.CreatedAt.String(),
		UpdatedAt: material.UpdatedAt.String(),
	}

	common_handlers.HandleSuccessResponse(c, http.StatusOK, res)
}

// UploadMaterials godoc
// @Summary      Upload materials
// @Description  Upload material IDs alongside material files mapping to a world.
// @Tags         assets-service
// @Security     BearerAuth
// @Accept       multipart/form-data
// @Produce      json
// @Param        world_id path string true "World UUID"
// @Param        ids formData []string true "Array of exact material IDs"
// @Param        materials formData file true "Muti-part chunk array of files"
// @Param        material_types formData []int true "Array of material types"
// @Success      201  {array}   dtos.MaterialResponse
// @Failure      400  {object} dtos.ErrorResponse
// @Failure      401  {object} dtos.ErrorResponse
// @Router       /assets/materials/world/{world_id} [put]
func (mc *materialsController) UploadMaterials(c *gin.Context) {
	userId, err := common_handlers.GetUserIDFromSession(c)
	if err != nil {
		_ = c.Error(errors.NewUnauthorizedError(err.Error()))
		return
	}

	worldId, err := uuid.Parse(c.Param("world_id"))
	if err != nil {
		_ = c.Error(errors.NewBadRequestError("invalid world_id format"))
		return
	}

	if worldId == uuid.Nil {
		if err := common_handlers.IsAdminSession(c); err != nil {
			_ = c.Error(errors.NewForbiddenError("only admins can upload global default materials"))
			return
		}
	}

	responseMaterials := make([]dtos.MaterialResponse, 0)

	i := 0
	for {
		idKey := fmt.Sprintf("ids[%d]", i)
		idVal := c.PostForm(idKey)

		if idVal == "" {
			break
		}

		id, err := uuid.Parse(idVal)
		if err != nil {
			_ = c.Error(errors.NewBadRequestError(fmt.Sprintf("invalid id format for ids[%d]: %s", i, err.Error())))
			return
		}

		materialTypeStr := c.PostForm(fmt.Sprintf("material_types[%d]", i))
		if materialTypeStr == "" {
			_ = c.Error(errors.NewBadRequestError(fmt.Sprintf("missing material_type for ids[%d]", i)))
			return
		}

		materialTypeInt, err := strconv.Atoi(materialTypeStr)
		if err != nil {
			_ = c.Error(errors.NewBadRequestError(fmt.Sprintf("invalid material_type for ids[%d]: %s", i, err.Error())))
			return
		}

		materialType, err := models.ParseMaterialType(materialTypeInt)
		if err != nil {
			_ = c.Error(errors.NewBadRequestError(fmt.Sprintf("invalid material_type for ids[%d]: %s", i, err.Error())))
			return
		}

		nameKey := fmt.Sprintf("names[%d]", i)
		name := c.PostForm(nameKey)
		if name == "" {
			_ = c.Error(errors.NewBadRequestError(fmt.Sprintf("missing name for ids[%d]", i)))
			return
		}

		file, err := c.FormFile(fmt.Sprintf("materials[%d]", i))
		if err != nil {
			_ = c.Error(errors.NewBadRequestError(fmt.Sprintf("Missing material file for ids[%d]", i)))
			return
		}

		material, err := mc.service.UploadMaterial(worldId, id, materialType, name, file, userId)
		if err != nil {
			_ = c.Error(err)
			return
		}

		responseMaterials = append(responseMaterials, dtos.MaterialResponse{
			ID:        material.ID,
			Name:      material.Name,
			URL:       material.URL,
			WorldID:   material.WorldID,
			CreatedAt: material.CreatedAt.String(),
			UpdatedAt: material.UpdatedAt.String(),
		})

		i++
	}

	common_handlers.HandleSuccessResponse(c, http.StatusCreated, responseMaterials)
}

// DeleteMaterial godoc
// @Summary      Delete material
// @Description  Delete a specific material by ID (requires ownership)
// @Tags         assets-service
// @Security     BearerAuth
// @Produce      json
// @Param        id path string true "Material UUID"
// @Success      204  {object}  dtos.MaterialResponse
// @Failure      400  {object} dtos.ErrorResponse
// @Failure      401  {object} dtos.ErrorResponse
// @Router       /assets/materials/{id} [delete]
func (mc *materialsController) DeleteMaterial(c *gin.Context) {
	userId, err := common_handlers.GetUserIDFromSession(c)
	if err != nil {
		_ = c.Error(errors.NewUnauthorizedError(err.Error()))
		return
	}

	materialId, err := uuid.Parse(c.Param("id"))
	if err != nil {
		_ = c.Error(errors.NewBadRequestError("invalid material_id format"))
		return
	}

	material, err := mc.service.GetMaterialByID(materialId)
	if err != nil {
		_ = c.Error(err)
		return
	}

	if material == nil {
		_ = c.Error(errors.NewBadRequestError("material not found"))
		return
	}

	if material.CreatedBy != userId {
		_ = c.Error(errors.NewUnauthorizedError("user is not authorized to delete this material"))
		return
	}

	if err := mc.service.DeleteMaterial(materialId); err != nil {
		_ = c.Error(err)
		return
	}

	res := &dtos.MaterialResponse{
		ID: materialId,
	}

	common_handlers.HandleSuccessResponse(c, http.StatusOK, res)
}
