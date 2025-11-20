package sprites

import (
	"fmt"
	"net/http"

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

func (sc *modelsController) DownloadModelsByWorldId(c *gin.Context) {
	// TODO: implement!!
}
func (mc *modelsController) UploadModelsByWorldId(ctx *gin.Context) {

	_, err := common_handlers.GetUserIDFromSession(ctx)
	if err != nil {
		_ = ctx.Error(errors.NewUnauthorizedError(err.Error()))
		return
	}

	worldIDStr := ctx.PostForm("world_id")
	if worldIDStr == "" {
		_ = ctx.Error(errors.NewNotFoundError("world_id is required"))
		return
	}

	worldID, err := uuid.Parse(worldIDStr)
	if err != nil {
		_ = ctx.Error(errors.NewBadRequestError("invalid world_id format"))
		return
	}

	form := ctx.Request.MultipartForm
	if form == nil || len(form.File) == 0 {
		_ = ctx.Error(errors.NewBadRequestError("no files uploaded"))
		return
	}

	modelsRequest := []models.Model{}

	for i := 0; ; i++ {
		name := ctx.PostForm(fmt.Sprintf("models[%d].name", i))
		if name == "" {
			break
		}
		modelIDStr := ctx.PostForm(fmt.Sprintf("models[%d].model_id", i))
		if modelIDStr == "" {
			_ = ctx.Error(errors.NewBadRequestError(fmt.Sprintf("model_id is required for model %d", i)))
			return
		}

		modelID, err := uuid.Parse(modelIDStr)
		if err != nil {
			_ = ctx.Error(errors.NewBadRequestError(fmt.Sprintf("invalid model_id format for model %d", i)))
			return
		}
		modelFile, err := ctx.FormFile(fmt.Sprintf("models[%d].model_file", i))
		if err != nil {
			_ = ctx.Error(errors.NewBadRequestError(fmt.Sprintf("model_file is required for model %d", i)))
			return
		}
		materialFile, err := ctx.FormFile(fmt.Sprintf("models[%d].material_file", i))
		if err != nil {
			_ = ctx.Error(errors.NewBadRequestError(fmt.Sprintf("material_file is required for model %d", i)))
			return
		}
		// Collect metadata and file headers
		modelsRequest = append(modelsRequest, models.Model{
			Name:         name,
			ModelID:      modelID,
			ModelFile:    modelFile,
			MaterialFile: materialFile,
		})
	}

	if len(modelsRequest) == 0 {
		_ = ctx.Error(errors.NewBadRequestError("no models uploaded"))
		return
	}

	savedModels, err := mc.modelService.PublishModels(worldID, modelsRequest)
	if err != nil {
		_ = ctx.Error(errors.NewInternalServerError(err.Error()))
		return
	}

	modelResponses := make([]dtos.ModelResponse, len(savedModels))
	for i, model := range savedModels {
		modelResponses[i] = dtos.ModelResponse{
			ModelID: model.ModelID,
			Name:    model.Name,
		}
	}

	modelsListResponse := dtos.ModelsListResponse{
		WorldID: worldID,
		List:    modelResponses,
	}
	common_handlers.HandleSuccessResponse(ctx, http.StatusCreated, modelsListResponse)
}
