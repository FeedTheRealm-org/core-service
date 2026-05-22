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
	"github.com/FeedTheRealm-org/core-service/internal/exports-service/models"
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

	if err := common_handlers.PrepareMultipartRequest(c); err != nil {
		_ = c.Error(err)
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

	common_handlers.HandleSuccessResponse(c, http.StatusCreated, buildExportZipResponse(exportZip))
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

	if err := validateVersionOptional(version); err != nil {
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

// ListZipVersions godoc
// @Summary      List export zip versions
// @Description  Retrieves all export versions, optionally filtered by app and OS.
// @Tags         exports-service
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        app query string false "App name"
// @Param        os query string false "OS name"
// @Success      200  {array}  dtos.ExportZipResponse
// @Failure      400  {object}  dtos.ErrorResponse
// @Failure      401  {object}  dtos.ErrorResponse
// @Router       /exports/zip/versions [get]
func (ec *exportsController) ListZipVersions(c *gin.Context) {
	_, err := common_handlers.GetUserIDFromSession(c)
	if err != nil {
		_ = c.Error(errors.NewUnauthorizedError(err.Error()))
		return
	}

	appNameStr := c.Query("app")
	osNameStr := c.Query("os")

	if appNameStr != "" {
		appName := exports_dtos.AppName(appNameStr)
		if !appName.Valid() {
			_ = c.Error(errors.NewBadRequestError("app must be one of: ftr_world_editor, ftr_game"))
			return
		}
	}

	if osNameStr != "" {
		osName := exports_dtos.OSName(osNameStr)
		if !osName.Valid() {
			_ = c.Error(errors.NewBadRequestError("os must be one of: linux, windows"))
			return
		}
	}

	exports, err := ec.service.ListZipVersions(appNameStr, osNameStr)
	if err != nil {
		_ = c.Error(err)
		return
	}

	res := make([]exports_dtos.ExportZipResponse, 0, len(exports))
	for _, exportZip := range exports {
		res = append(res, buildExportZipResponse(exportZip))
	}

	common_handlers.HandleSuccessResponse(c, http.StatusOK, res)
}

// DeleteZipVersion godoc
// @Summary      Delete export zip version
// @Description  Deletes a specific export version.
// @Tags         exports-service
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        app query string true "App name"
// @Param        version query string true "Version"
// @Param        os query string true "OS name"
// @Success      204
// @Failure      400  {object}  dtos.ErrorResponse
// @Failure      401  {object}  dtos.ErrorResponse
// @Failure      404  {object}  dtos.ErrorResponse
// @Router       /exports/zip [delete]
func (ec *exportsController) DeleteZipVersion(c *gin.Context) {
	_, err := common_handlers.GetUserIDFromSession(c)
	if err != nil {
		_ = c.Error(errors.NewUnauthorizedError(err.Error()))
		return
	}

	appNameStr := c.Query("app")
	version := c.Query("version")
	osNameStr := c.Query("os")

	if appNameStr == "" || version == "" || osNameStr == "" {
		_ = c.Error(errors.NewBadRequestError("app, version and os are required"))
		return
	}

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

	if err := ec.service.DeleteZipVersion(appNameStr, version, osNameStr); err != nil {
		switch err.(type) {
		case *exports_errors.ExportNotFound:
			_ = c.Error(errors.NewNotFoundError("export zip not found"))
		default:
			_ = c.Error(err)
		}
		return
	}

	common_handlers.HandleBodilessResponse(c, http.StatusNoContent)
}

// SetLatestZipVersion godoc
// @Summary      Set export zip version as latest
// @Description  Marks one export version as the latest for its app and OS.
// @Tags         exports-service
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        request body dtos.ExportZipSetLatestRequest true "Latest version payload"
// @Success      200  {object}  dtos.ExportZipResponse
// @Failure      400  {object}  dtos.ErrorResponse
// @Failure      401  {object}  dtos.ErrorResponse
// @Failure      404  {object}  dtos.ErrorResponse
// @Router       /exports/zip/latest [patch]
func (ec *exportsController) SetLatestZipVersion(c *gin.Context) {
	_, err := common_handlers.GetUserIDFromSession(c)
	if err != nil {
		_ = c.Error(errors.NewUnauthorizedError(err.Error()))
		return
	}

	var req exports_dtos.ExportZipSetLatestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(errors.NewBadRequestError("invalid request body: " + err.Error()))
		return
	}

	appName := exports_dtos.AppName(req.AppName)
	if !appName.Valid() {
		_ = c.Error(errors.NewBadRequestError("app must be one of: ftr_world_editor, ftr_game"))
		return
	}

	osName := exports_dtos.OSName(req.OS)
	if !osName.Valid() {
		_ = c.Error(errors.NewBadRequestError("os must be one of: linux, windows"))
		return
	}

	if err := validateVersion(req.Version); err != nil {
		_ = c.Error(err)
		return
	}

	exportZip, err := ec.service.SetLatestZipVersion(req.AppName, req.Version, req.OS)
	if err != nil {
		switch err.(type) {
		case *exports_errors.ExportNotFound:
			_ = c.Error(errors.NewNotFoundError("export zip not found"))
		default:
			_ = c.Error(err)
		}
		return
	}

	common_handlers.HandleSuccessResponse(c, http.StatusOK, buildExportZipResponse(exportZip))
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

func validateVersionOptional(version string) error {
	if version == "" {
		return nil
	}
	return validateVersion(version)
}

func buildExportZipResponse(exportZip *models.ExportZip) exports_dtos.ExportZipResponse {
	return exports_dtos.ExportZipResponse{
		AppName:   exportZip.AppName,
		Version:   exportZip.Version,
		OS:        exportZip.OS,
		Path:      exportZip.Path,
		IsLatest:  exportZip.IsLatest,
		CreatedAt: exportZip.CreatedAt,
		UpdatedAt: exportZip.UpdatedAt,
	}
}

func openMultipartFile(fileHeader *multipart.FileHeader) (multipart.File, error) {
	return common_handlers.OpenMultipartFile(fileHeader)
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}
