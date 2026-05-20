package exports

import (
	"mime/multipart"
	"net/http"
	"path/filepath"
	"regexp"

	"github.com/FeedTheRealm-org/core-service/config"
	"github.com/FeedTheRealm-org/core-service/internal/common_handlers"
	"github.com/FeedTheRealm-org/core-service/internal/errors"
	exports_dtos "github.com/FeedTheRealm-org/core-service/internal/exports-service/dtos"
	exports_errors "github.com/FeedTheRealm-org/core-service/internal/exports-service/errors"
	exports_service "github.com/FeedTheRealm-org/core-service/internal/exports-service/services/exports"
	"github.com/gin-gonic/gin"
)

var versionPattern = regexp.MustCompile(`^v\d+\.\d+\.\d+$`)

type exportsController struct {
	conf    *config.Config
	service exports_service.ExportsService
}

// NewExportsController creates a new instance of ExportsController.
func NewExportsController(conf *config.Config, service exports_service.ExportsService) ExportsController {
	return &exportsController{
		conf:    conf,
		service: service,
	}
}

// UploadZip godoc
// @Summary      Upload export zip
// @Description  Uploads a zip file for a given app, version, and OS.
// @Tags         exports-service
// @Security     BearerAuth
// @Accept       multipart/form-data
// @Produce      json
// @Param        app formData string true "App name"
// @Param        version formData string true "Version (vX.Y.Z)"
// @Param        os formData string true "OS name"
// @Param        file formData file true "Zip file"
// @Success      201  {object}  dtos.ExportZipResponse
// @Failure      400  {object}  dtos.ErrorResponse
// @Failure      401  {object}  dtos.ErrorResponse
// @Failure      409  {object}  dtos.ErrorResponse
// @Router       /exports/zip [put]
func (ec *exportsController) UploadZip(c *gin.Context) {
	_, err := common_handlers.GetUserIDFromSession(c)
	if err != nil {
		_ = c.Error(errors.NewUnauthorizedError(err.Error()))
		return
	}

	appNameStr := firstNonEmpty(c.PostForm("app"), c.PostForm("app_name"))
	version := firstNonEmpty(c.PostForm("version"), c.PostForm("versiion"))
	osNameStr := c.PostForm("os")

	// Validate inputs
	appName := exports_dtos.AppName(appNameStr)
	if !appName.Valid() {
		_ = c.Error(errors.NewBadRequestError("app must be one of: ftr_world_editor, ftr_game"))
		return
	}

	osName := exports_dtos.OSName(osNameStr)
	if !osName.Valid() {
		_ = c.Error(errors.NewBadRequestError("os must be one of: linux, windows"))
		return
	}

	if err := validateVersion(version); err != nil {
		_ = c.Error(err)
		return
	}

	fileHeader, err := c.FormFile("file")
	if err != nil {
		_ = c.Error(errors.NewBadRequestError("missing zip file"))
		return
	}

	if filepath.Ext(fileHeader.Filename) != ".zip" {
		_ = c.Error(errors.NewBadRequestError("file must be a .zip"))
		return
	}

	zipFile, err := openMultipartFile(fileHeader)
	if err != nil {
		_ = c.Error(errors.NewBadRequestError("failed to open zip file"))
		return
	}
	defer func() {
		_ = zipFile.Close()
	}()

	exportZip, err := ec.service.UploadZip(appNameStr, version, osNameStr, zipFile)
	if err != nil {
		switch err.(type) {
		case *exports_errors.ExportVersionConflict:
			_ = c.Error(errors.NewConflictError("export version already exists"))
		default:
			_ = c.Error(err)
		}
		return
	}

	res := &exports_dtos.ExportZipResponse{
		AppName: exportZip.AppName,
		Version: exportZip.Version,
		OS:      exportZip.OS,
		Path:    exportZip.Path,
	}

	common_handlers.HandleSuccessResponse(c, http.StatusCreated, res)
}

// GetZipPath godoc
// @Summary      Get export zip path
// @Description  Retrieves the path for a given app, version, and OS.
// @Tags         exports-service
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        app query string true "App name"
// @Param        version query string true "Version (vX.Y.Z)"
// @Param        os query string true "OS name"
// @Success      200  {object}  dtos.ExportZipPathResponse
// @Failure      400  {object}  dtos.ErrorResponse
// @Failure      401  {object}  dtos.ErrorResponse
// @Failure      404  {object}  dtos.ErrorResponse
// @Router       /exports/zip [get]
func (ec *exportsController) GetZipPath(c *gin.Context) {
	_, err := common_handlers.GetUserIDFromSession(c)
	if err != nil {
		_ = c.Error(errors.NewUnauthorizedError(err.Error()))
		return
	}

	appNameStr := firstNonEmpty(c.Query("app"), c.Query("app_name"))
	version := firstNonEmpty(c.Query("version"), c.Query("versiion"))
	osNameStr := c.Query("os")

	// Validate inputs
	appName := exports_dtos.AppName(appNameStr)
	if !appName.Valid() {
		_ = c.Error(errors.NewBadRequestError("app must be one of: ftr_world_editor, ftr_game"))
		return
	}

	osName := exports_dtos.OSName(osNameStr)
	if !osName.Valid() {
		_ = c.Error(errors.NewBadRequestError("os must be one of: linux, windows"))
		return
	}

	if err := validateVersion(version); err != nil {
		_ = c.Error(err)
		return
	}

	path, err := ec.service.GetZipPath(appNameStr, version, osNameStr)
	if err != nil {
		switch err.(type) {
		case *exports_errors.ExportNotFound:
			_ = c.Error(errors.NewNotFoundError("export zip not found"))
		default:
			_ = c.Error(err)
		}
		return
	}

	res := &exports_dtos.ExportZipPathResponse{
		Path: path,
	}

	common_handlers.HandleSuccessResponse(c, http.StatusOK, res)
}

func validateVersion(version string) error {
	if version == "" {
		return errors.NewBadRequestError("version is required")
	}
	if !versionPattern.MatchString(version) {
		return errors.NewBadRequestError("version must match vX.Y.Z")
	}
	return nil
}

func openMultipartFile(fileHeader *multipart.FileHeader) (multipart.File, error) {
	return fileHeader.Open()
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}
