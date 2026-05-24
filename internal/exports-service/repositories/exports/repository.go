package exports

import "github.com/FeedTheRealm-org/core-service/internal/exports-service/models"

// ExportRepository defines the interface for export zip database operations.
type ExportRepository interface {
	CreateExportVersion(exportZip *models.ExportZip) error
	GetExportVersion(appName, version, osName string) (*models.ExportZip, error)
	GetLatestExportVersion(appName, osName string) (*models.ExportZip, error)
	ListExportVersions(appName, osName string) ([]*models.ExportZip, error)
	DeleteExportVersion(appName, version, osName string) error
	SetLatestExportVersion(appName, version, osName string) (*models.ExportZip, error)
}
