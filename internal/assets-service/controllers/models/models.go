package sprites

import (
	"fmt"
	"net/http"

	"github.com/FeedTheRealm-org/core-service/config"
	"github.com/FeedTheRealm-org/core-service/internal/assets-service/dtos"
	service "github.com/FeedTheRealm-org/core-service/internal/assets-service/services/models"
	"github.com/FeedTheRealm-org/core-service/internal/common_handlers"
	"github.com/FeedTheRealm-org/core-service/internal/errors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type modelsController struct {
	conf         *config.Config
	modelService service.ModelsService
}

// NewModelsController creates a new instance of ModelsController.
func NewModelsController(conf *config.Config, modelService service.ModelsService) ModelsController {
	return &modelsController{
		conf:         conf,
		modelService: modelService,
	}
}

func (mc *modelsController) GetModelsList(c *gin.Context) {
	_, err := common_handlers.GetUserIDFromSession(c)
	if err != nil {
		_ = c.Error(errors.NewUnauthorizedError(err.Error()))
		return
	}

	worldID, err := parseUUIDParam(c, "world_id")
	if err != nil {
		_ = c.Error(err)
		return
	}

	modelsList, err := mc.modelService.GetModelsByWorld(worldID)
	if err != nil {
		_ = c.Error(errors.NewInternalServerError(err.Error()))
		return
	}

	// Also fetch global default models (nil world ID)
	if worldID != uuid.Nil {
		defaultModels, err := mc.modelService.GetModelsByWorld(uuid.Nil)
		if err != nil {
			_ = c.Error(errors.NewInternalServerError(err.Error()))
			return
		}
		modelsList = append(modelsList, defaultModels...)
	}

	modelResponseList := make([]dtos.ModelResponse, len(modelsList))
	for i, model := range modelsList {
		modelResponseList[i] = dtos.ModelResponse{
			ModelID: model.Id,
			Url:     model.Url,
		}
	}

	common_handlers.HandleSuccessResponse(c, http.StatusOK, dtos.ModelsListResponse{
		WorldID: worldID,
		List:    modelResponseList,
	})
}

// UploadModel godoc
// @Summary      Upload a 3D model
// @Description  Submit a single custom model bounded to a specific world.
// @Tags         assets-service
// @Security     BearerAuth
// @Accept       multipart/form-data
// @Produce      json
// @Param        world_id  path      string true "World UUID"
// @Param        model_id  formData  string true "Model UUID"
// @Param        model_file formData file   true "Model GLB file"
// @Success      201  {object}  dtos.ModelResponse
// @Failure      400  {object}  dtos.ErrorResponse
// @Failure      401  {object}  dtos.ErrorResponse
// @Failure      500  {object}  dtos.ErrorResponse
// @Router       /assets/models/world/{world_id} [put]
func (mc *modelsController) UploadModel(c *gin.Context) {
	userId, err := common_handlers.GetUserIDFromSession(c)
	if err != nil {
		_ = c.Error(errors.NewUnauthorizedError(err.Error()))
		return
	}

	worldID, err := parseUUIDParam(c, "world_id")
	if err != nil {
		_ = c.Error(err)
		return
	}

	if worldID == uuid.Nil {
		if err := common_handlers.IsAdminSession(c); err != nil {
			_ = c.Error(errors.NewUnauthorizedError("only admins can upload global models"))
			return
		}
	} else {
		if err := common_handlers.IsAdminSession(c); err != nil {
			if err := common_handlers.CheckWorldOwnership(c, mc.conf.Server.Port, worldID, userId); err != nil {
				_ = c.Error(err)
				return
			}
		}
	}

	modelID, err := parseUUIDForm(c, "model_id")
	if err != nil {
		_ = c.Error(err)
		return
	}

	modelFile, err := c.FormFile("model_file")
	if err != nil {
		_ = c.Error(errors.NewBadRequestError("model_file is required"))
		return
	}

	savedModel, err := mc.modelService.UploadModel(dtos.ModelRequest{
		WorldID:   worldID,
		Id:        modelID,
		CreatedBy: userId,
		ModelFile: modelFile,
	})
	if err != nil {
		_ = c.Error(errors.NewInternalServerError(err.Error()))
		return
	}

	common_handlers.HandleSuccessResponse(c, http.StatusCreated, dtos.ModelResponse{
		ModelID: savedModel.Id,
		Url:     savedModel.Url,
	})
}

func parseUUIDParam(c *gin.Context, key string) (uuid.UUID, error) {
	val := c.Param(key)
	if val == "" {
		return uuid.Nil, errors.NewBadRequestError(fmt.Sprintf("%s is required", key))
	}
	id, err := uuid.Parse(val)
	if err != nil {
		return uuid.Nil, errors.NewBadRequestError(fmt.Sprintf("invalid %s format", key))
	}
	return id, nil
}

func parseUUIDForm(c *gin.Context, key string) (uuid.UUID, error) {
	val := c.PostForm(key)
	if val == "" {
		return uuid.Nil, errors.NewBadRequestError(fmt.Sprintf("%s is required", key))
	}
	id, err := uuid.Parse(val)
	if err != nil {
		return uuid.Nil, errors.NewBadRequestError(fmt.Sprintf("invalid %s format", key))
	}
	return id, nil
}
