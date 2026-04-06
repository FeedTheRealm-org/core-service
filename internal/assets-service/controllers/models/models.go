package sprites

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/FeedTheRealm-org/core-service/config"
	"github.com/FeedTheRealm-org/core-service/internal/assets-service/dtos"
	models "github.com/FeedTheRealm-org/core-service/internal/assets-service/models"
	service "github.com/FeedTheRealm-org/core-service/internal/assets-service/services/models"
	"github.com/FeedTheRealm-org/core-service/internal/common_handlers"
	"github.com/FeedTheRealm-org/core-service/internal/errors"
	"github.com/FeedTheRealm-org/core-service/internal/utils/logger"
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

// GetModelsList godoc
// @Summary      Get 3D models list
// @Description  Retrieve the metadata models available in a given world id.
// @Tags         assets-service
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        world_id path string true "World UUID"
// @Success      200  {object}  dtos.ModelsListResponse
// @Failure      400  {object} dtos.ErrorResponse
// @Failure      401  {object} dtos.ErrorResponse
// @Failure      404  {object} dtos.ErrorResponse
// @Failure      500  {object} dtos.ErrorResponse
// @Router       /assets/models/world/{world_id} [get]
func (mc *modelsController) GetModelsList(c *gin.Context) {
	_, err := common_handlers.GetUserIDFromSession(c)
	if err != nil {
		_ = c.Error(errors.NewUnauthorizedError(err.Error()))
		return
	}
	worldIDStr := c.Param("world_id")
	if worldIDStr == "" {
		_ = c.Error(errors.NewNotFoundError("world_id is required"))
		return
	}

	worldID, err := uuid.Parse(worldIDStr)
	if err != nil {
		_ = c.Error(errors.NewBadRequestError("invalid world_id format"))
		return
	}

	modelsList, err := mc.modelService.GetModelsByWorld(worldID)
	if err != nil {
		_ = c.Error(errors.NewInternalServerError(err.Error()))
		return
	}

	modelResponseList := make([]dtos.ModelResponse, len(modelsList))
	for i, model := range modelsList {
		modelResponseList[i] = dtos.ModelResponse{
			ModelID: model.Id,
			Url:     model.Url,
		}
	}

	response := dtos.ModelsListResponse{
		WorldID: worldID,
		List:    modelResponseList,
	}
	common_handlers.HandleSuccessResponse(c, http.StatusOK, response)
}

// DownloadModel godoc
// @Summary      Download binary model
// @Description  Stream/Download a .glb model file.
// @Tags         assets-service
// @Security     BearerAuth
// @Produce      application/gltf-binary
// @Param        world_id path string true "World UUID"
// @Param        model_id path string true "Model UUID"
// @Success      200  {string}  string "GLB file body"
// @Failure      400  {object} dtos.ErrorResponse
// @Failure      404  {object} dtos.ErrorResponse
// @Failure      500  {object} dtos.ErrorResponse
func (mc *modelsController) DownloadModel(c *gin.Context) {
	// _, err := common_handlers.GetUserIDFromSession(c)
	// if err != nil {
	// 	_ = c.Error(errors.NewUnauthorizedError(err.Error()))
	// 	return
	// }
	worldID := c.Param("world_id")
	assetID := c.Param("model_id")

	if worldID == "" || assetID == "" {
		_ = c.Error(errors.NewBadRequestError("world_id and model_id are required"))
		return
	}
	// Path: bucket/worlds/<worldId>/models/<modelId>/model.glb
	modelPath := filepath.Join(
		"bucket",
		"worlds",
		worldID,
		"models",
		assetID,
		"model.glb",
	)

	// Validate file exists
	if _, err := os.Stat(modelPath); err != nil {
		_ = c.Error(errors.NewNotFoundError("model not found for this asset"))
		return
	}

	// Stream model file directly
	c.Header("Content-Type", "model/gltf-binary") // official GLB MIME
	c.Header("Content-Disposition",
		fmt.Sprintf(`attachment; filename="%s_model.glb"`, assetID))

	c.File(modelPath)
}

// UploadModels godoc
// @Summary      Upload batch 3D models
// @Description  Submit multiple custom models bounded to a specific world.
// @Tags         assets-service
// @Security     BearerAuth
// @Accept       multipart/form-data
// @Produce      json
// @Param        world_id path string true "World UUID"
// @Param        models formData file true "Multipart upload array for model bindings"
// @Success      201  {object}  dtos.ModelsListResponse
// @Failure      400  {object} dtos.ErrorResponse
// @Failure      401  {object} dtos.ErrorResponse
// @Failure      500  {object} dtos.ErrorResponse
// @Router       /assets/models/world/{world_id} [put]
func (mc *modelsController) UploadModels(c *gin.Context) {
	userId, err := common_handlers.GetUserIDFromSession(c)
	if err != nil {
		_ = c.Error(errors.NewUnauthorizedError(err.Error()))
		return
	}

	worldIDStr := c.Param("world_id")
	if worldIDStr == "" {
		_ = c.Error(errors.NewNotFoundError("world_id is required"))
		return
	}

	worldID, err := uuid.Parse(worldIDStr)
	if err != nil {
		_ = c.Error(errors.NewBadRequestError("invalid world_id format"))
		return
	}

	modelsRequest := []models.Model{}
	hasModels := true
	for i := 0; hasModels; i++ {

		modelIDStr := c.PostForm(fmt.Sprintf("models[%d].model_id", i))
		if modelIDStr == "" {
			hasModels = false
			continue
		}

		modelID, err := uuid.Parse(modelIDStr)
		if err != nil {
			_ = c.Error(errors.NewBadRequestError(fmt.Sprintf("invalid model_id format for model %d", i)))
			return
		}
		modelFile, err := c.FormFile(fmt.Sprintf("models[%d].model_file", i))
		if err != nil {
			_ = c.Error(errors.NewBadRequestError(fmt.Sprintf("model_file is required for model %d", i)))
			return
		}
		// Collect metadata and file headers
		modelsRequest = append(modelsRequest, models.Model{
			Id:        modelID,
			ModelFile: modelFile,
			CreatedBy: userId,
		})
	}
	logger.Logger.Infof("CONTROLLER: Models request parsed with %d models", len(modelsRequest))

	if len(modelsRequest) == 0 {
		_ = c.Error(errors.NewBadRequestError("no models uploaded"))
		return
	}

	savedModels, err := mc.modelService.UploadModels(worldID, modelsRequest)
	if err != nil {
		_ = c.Error(errors.NewInternalServerError(err.Error()))
		return
	}

	logger.Logger.Infof("CONTROLLER: Models uploaded and metadata saved for %d models", len(savedModels))

	modelResponses := make([]dtos.ModelResponse, len(savedModels))
	for i, model := range savedModels {
		modelResponses[i] = dtos.ModelResponse{
			ModelID: model.Id,
			Url:     model.Url,
		}
	}

	modelsListResponse := dtos.ModelsListResponse{
		WorldID: worldID,
		List:    modelResponses,
	}
	common_handlers.HandleSuccessResponse(c, http.StatusCreated, modelsListResponse)
}
