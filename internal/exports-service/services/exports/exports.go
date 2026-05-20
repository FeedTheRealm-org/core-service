package exports

import (
	"fmt"
	"mime/multipart"

	"github.com/FeedTheRealm-org/core-service/config"
	"github.com/FeedTheRealm-org/core-service/internal/exports-service/models"
	"github.com/FeedTheRealm-org/core-service/internal/exports-service/repositories/bucket"
	exports_repo "github.com/FeedTheRealm-org/core-service/internal/exports-service/repositories/exports"
	"github.com/FeedTheRealm-org/core-service/internal/utils/logger"
)

type exportsService struct {
	conf       *config.Config
	repository exports_repo.ExportRepository
	bucketRepo bucket.BucketRepository
}

// NewExportsService creates a new instance of ExportsService.
func NewExportsService(conf *config.Config, repository exports_repo.ExportRepository, bucketRepo bucket.BucketRepository) ExportsService {
	return &exportsService{
		conf:       conf,
		repository: repository,
		bucketRepo: bucketRepo,
	}
}

func (es *exportsService) UploadZip(appName, version, osName string, zipFile multipart.File) (*models.ExportZip, error) {
	filePath := buildExportFilePath(appName, version, osName)

	if err := es.bucketRepo.UploadFile(filePath, "application/zip", zipFile); err != nil {
		logger.Logger.Errorf("Error uploading export zip: %v", err)
		return nil, err
	}

	exportZip := &models.ExportZip{
		AppName:  appName,
		Version:  version,
		OS:       osName,
		Path:     fmt.Sprintf("/%s", filePath),
		IsLatest: false,
	}

	if err := es.repository.CreateExportVersion(exportZip); err != nil {
		logger.Logger.Errorf("Error creating export version: %v", err)
		return nil, err
	}

	return exportZip, nil
}

func (es *exportsService) GetZipPath(appName, version, osName string) (string, error) {
	var exportZip *models.ExportZip
	var err error
	if version == "" {
		exportZip, err = es.repository.GetLatestExportVersion(appName, osName)
	} else {
		exportZip, err = es.repository.GetExportVersion(appName, version, osName)
	}
	if err != nil {
		return "", err
	}
	return exportZip.Path, nil
}

func (es *exportsService) ListZipVersions(appName, osName string) ([]*models.ExportZip, error) {
	return es.repository.ListExportVersions(appName, osName)
}

func (es *exportsService) DeleteZipVersion(appName, version, osName string) error {
	if err := es.repository.DeleteExportVersion(appName, version, osName); err != nil {
		logger.Logger.Errorf("Error deleting export version: %v", err)
		return err
	}

	filePath := buildExportFilePath(appName, version, osName)
	if err := es.bucketRepo.DeleteFile(filePath); err != nil {
		logger.Logger.Errorf("Error deleting export zip from bucket: %v", err)
		return err
	}

	return nil
}

func (es *exportsService) SetLatestZipVersion(appName, version, osName string) (*models.ExportZip, error) {
	return es.repository.SetLatestExportVersion(appName, version, osName)
}

func buildExportFilePath(appName, version, osName string) string {
	fileName := fmt.Sprintf("%s-%s.zip", appName, version)
	return fmt.Sprintf("exports/%s/%s/%s", appName, osName, fileName)
}
