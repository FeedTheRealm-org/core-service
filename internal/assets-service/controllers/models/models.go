package sprites

import (
	"archive/zip"
	"fmt"
	"io"
	"io/fs"
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

// DownloadModelsByWorldId downloads all models and materials for a specific world as a ZIP file
// @Summary Download world models
// @Description Downloads all 3D models and their materials for a specific world as a ZIP archive
// @Tags Models
// @Accept json
// @Produce application/zip
// @Param world_id path string true "World ID" format(uuid)
// @Success 200 {file} binary "ZIP file containing all models and materials"
// @Failure 400 {object} map[string]interface{} "Bad request - invalid world_id format"
// @Failure 401 {object} map[string]interface{} "Unauthorized - invalid or missing JWT token"
// @Failure 404 {object} map[string]interface{} "Not found - no models found for this world"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /assets/models/{world_id} [get]
func (mc *modelsController) DownloadModelsByWorldId(c *gin.Context) {
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

	// Path: bucket/worlds/<worldId>/models
	modelsDir := filepath.Join("bucket", "worlds", worldID.String(), "models")
	info, err := os.Stat(modelsDir)
	if err != nil || !info.IsDir() {
		_ = c.Error(errors.NewNotFoundError("models folder not found for this world"))
		return
	}

	// Generate filename like: <uuid>.zip
	zipFilename := fmt.Sprintf("%s.zip", worldID.String())

	// Set headers for file download
	c.Header("Content-Type", "application/zip")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", zipFilename))
	c.Header("Transfer-Encoding", "chunked") // enables streaming

	zipWriter := zip.NewWriter(c.Writer)
	defer func() {
		_ = zipWriter.Close()
	}()

	// Walk through the models directory and write to ZIP
	err = filepath.WalkDir(modelsDir, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			return nil
		}

		// Keep only relative path inside zip (e.g., 6445f0a2/.../model.fbx)
		relPath, err := filepath.Rel(modelsDir, path)
		if err != nil {
			return err
		}

		return addFileToZip(zipWriter, path, filepath.ToSlash(relPath))
	})

	if err != nil {
		_ = c.Error(errors.NewInternalServerError("failed to create zip: " + err.Error()))
		return
	}
}

// Helper: Add file to ZIP
func addFileToZip(zipWriter *zip.Writer, filePath, zipPath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file %s: %w", filePath, err)
	}
	defer func() {
		_ = file.Close()
	}()

	// Get file info
	fileInfo, err := file.Stat()
	if err != nil {
		return fmt.Errorf("failed to get file info: %w", err)
	}

	// Create ZIP header
	header, err := zip.FileInfoHeader(fileInfo)
	if err != nil {
		return fmt.Errorf("failed to create zip header: %w", err)
	}

	header.Name = zipPath // relative path inside zip
	header.Method = zip.Deflate

	// Create ZIP writer
	writer, err := zipWriter.CreateHeader(header)
	if err != nil {
		return fmt.Errorf("failed to create zip writer: %w", err)
	}

	// Stream file into ZIP (efficient, no RAM loading)
	_, err = io.Copy(writer, file)
	if err != nil {
		return fmt.Errorf("failed to copy file content: %w", err)
	}

	return nil
}

// UploadModelsByWorldId uploads multiple 3D models and materials to a specific world
// @Summary Upload world models
// @Description Upload multiple 3D models with their materials to a specific world. Supports FBX, OBJ model files and various material formats.
// @Tags Models
// @Accept multipart/form-data
// @Produce json
// @Param world_id formData string true "World ID" format(uuid)
// @Param models[0].name formData string true "Name of the first model"
// @Param models[0].model_id formData string true "Unique ID for the first model" format(uuid)
// @Param models[0].model_file formData file true "3D model file (FBX, OBJ, etc.)"
// @Param models[0].material_file formData file true "Material file (MAT, MTL, etc.)"
// @Param models[1].name formData string false "Name of the second model"
// @Param models[1].model_id formData string false "Unique ID for the second model" format(uuid)
// @Param models[1].model_file formData file false "3D model file for second model"
// @Param models[1].material_file formData file false "Material file for second model"
// @Success 201 {object} dtos.ModelsPublishListResponse "Successfully uploaded models"
// @Failure 400 {object} map[string]interface{} "Bad request - missing required fields or invalid format"
// @Failure 401 {object} map[string]interface{} "Unauthorized - invalid or missing JWT token"
// @Failure 500 {object} map[string]interface{} "Internal server error - failed to save models"
// @Router /assets/models [post]
func (mc *modelsController) UploadModelsByWorldId(c *gin.Context) {

	_, err := common_handlers.GetUserIDFromSession(c)
	if err != nil {
		_ = c.Error(errors.NewUnauthorizedError(err.Error()))
		return
	}

	worldIDStr := c.PostForm("world_id")
	if worldIDStr == "" {
		_ = c.Error(errors.NewNotFoundError("world_id is required"))
		return
	}

	worldID, err := uuid.Parse(worldIDStr)
	if err != nil {
		_ = c.Error(errors.NewBadRequestError("invalid world_id format"))
		return
	}

	form := c.Request.MultipartForm
	if form == nil || len(form.File) == 0 {
		_ = c.Error(errors.NewBadRequestError("no files uploaded"))
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
		materialFile, err := c.FormFile(fmt.Sprintf("models[%d].material_file", i))
		if err != nil {
			_ = c.Error(errors.NewBadRequestError(fmt.Sprintf("material_file is required for model %d", i)))
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
			ModelID: model.ModelID,
			Name:    model.Name,
		}
	}

	modelsListResponse := dtos.ModelsPublishListResponse{
		WorldID: worldID,
		List:    modelResponses,
	}
	common_handlers.HandleSuccessResponse(c, http.StatusCreated, modelsListResponse)
}
