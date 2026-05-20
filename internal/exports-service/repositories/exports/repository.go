package exports

import "github.com/FeedTheRealm-org/core-service/internal/exports-service/models"

// ExportRepository defines the interface for export zip database operations.
type ExportRepository interface {
	CreateExportVersion(exportZip *models.ExportZip) error
	GetExportVersion(appName, version, osName string) (*models.ExportZip, error)
}
