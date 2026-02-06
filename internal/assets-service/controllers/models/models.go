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

func (mc *modelsController) ListAssets(c *gin.Context) {
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
		}
	}

	response := dtos.ModelsListResponse{
		WorldID: worldID,
		List:    modelResponseList,
	}
	common_handlers.HandleSuccessResponse(c, http.StatusOK, response)
}

// DownloadModel downloads a single model file for a specific asset in a world
// @Summary Download model file
// @Description Downloads a single 3D model file (GLB) for a specific asset in a world
// @Tags Models
// @Accept json
// @Produce application/gltf-binary
// @Param world_id path string true "World ID" format(uuid)
// @Param model_id path string true "Asset ID" format(uuid)
// @Security BearerAuth
// @Success 200 {file} binary "GLB model file"
// @Failure 400 {object} map[string]interface{} "Bad request - missing world_id or model_id"
// @Failure 404 {object} map[string]interface{} "Not found - model file not found for this asset"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /assets/models/{world_id}/assets/{model_id}/model [get]
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

// UploadModels uploads multiple 3D models and materials to a specific world
// @Summary Upload world models
// @Description Upload multiple 3D models with their materials to a specific world. Supports GLB, FBX, OBJ model files and various material formats. Material files are optional.
// @Tags Models
// @Accept multipart/form-data
// @Produce json
// @Param world_id path string true "World ID" format(uuid)
// @Param models[0].name formData string true "Name of the first model"
// @Param models[0].model_id formData string true "Unique ID for the first model" format(uuid)
// @Param models[0].model_file formData file true "3D model file (GLB, FBX, OBJ, etc.)"
// @Param models[0].material_file formData file false "Material file (optional)"
// @Param models[1].name formData string false "Name of the second model"
// @Param models[1].model_id formData string false "Unique ID for the second model" format(uuid)
// @Param models[1].model_file formData file false "3D model file for second model"
// @Param models[1].material_file formData file false "Material file for second model (optional)"
// @Security BearerAuth
// @Success 201 {object} dtos.ModelsPublishListResponse "Successfully uploaded models"
// @Failure 400 {object} map[string]interface{} "Bad request - missing required fields or invalid format"
// @Failure 401 {object} map[string]interface{} "Unauthorized - invalid or missing JWT token"
// @Failure 500 {object} map[string]interface{} "Internal server error - failed to save models"
// @Router /assets/models/{world_id} [post]
// @Router /assets/models [post]
func (mc *modelsController) UploadModels(c *gin.Context) {
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

	modelsRequest := []models.Model{}
	for i := 0; ; i++ {
		name := c.PostForm(fmt.Sprintf("models[%d].name", i))
		if name == "" {
			break
		}
		modelIDStr := c.PostForm(fmt.Sprintf("models[%d].model_id", i))
		if modelIDStr == "" {
			_ = c.Error(errors.NewBadRequestError(fmt.Sprintf("model_id is required for model %d", i)))
			return
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
		})
	}

	if len(modelsRequest) == 0 {
		_ = c.Error(errors.NewBadRequestError("no models uploaded"))
		return
	}

	savedModels, err := mc.modelService.PublishModels(worldID, modelsRequest)
	if err != nil {
		_ = c.Error(errors.NewInternalServerError(err.Error()))
		return
	}

	modelResponses := make([]dtos.ModelPublishResponse, len(savedModels))
	for i, model := range savedModels {
		modelResponses[i] = dtos.ModelPublishResponse{
			ModelID: model.Id,
		}
	}

	modelsListResponse := dtos.ModelsPublishListResponse{
		WorldID: worldID,
		List:    modelResponses,
	}
	common_handlers.HandleSuccessResponse(c, http.StatusCreated, modelsListResponse)
}
