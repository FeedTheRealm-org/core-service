package exports

import (
	"mime/multipart"

	"github.com/FeedTheRealm-org/core-service/internal/exports-service/models"
)

// ExportsService defines export-related business logic.
type ExportsService interface {
	UploadZip(appName, version, osName string, zipFile multipart.File) (*models.ExportZip, error)
	GetZipPath(appName, version, osName string) (string, error)
	ListZipVersions(appName, osName string) ([]*models.ExportZip, error)
	DeleteZipVersion(appName, version, osName string) error
	SetLatestZipVersion(appName, version, osName string) (*models.ExportZip, error)
}
