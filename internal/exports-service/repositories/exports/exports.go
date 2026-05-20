package exports

import (
	"github.com/FeedTheRealm-org/core-service/config"
	"github.com/FeedTheRealm-org/core-service/internal/errors"
	exports_errors "github.com/FeedTheRealm-org/core-service/internal/exports-service/errors"
	"github.com/FeedTheRealm-org/core-service/internal/exports-service/models"
)

type exportRepository struct {
	conf *config.Config
	db   *config.DB
}

// NewExportRepository creates a new instance of ExportRepository.
func NewExportRepository(conf *config.Config, db *config.DB) ExportRepository {
	return &exportRepository{
		conf: conf,
		db:   db,
	}
}

func (er *exportRepository) CreateExportVersion(exportZip *models.ExportZip) error {
	if err := er.db.Conn.Create(exportZip).Error; err != nil {
		if errors.IsDuplicateEntryError(err) {
			return exports_errors.NewExportVersionConflict("export version already exists")
		}
		return err
	}
	return nil
}

func (er *exportRepository) GetExportVersion(appName, version, osName string) (*models.ExportZip, error) {
	var exportZip models.ExportZip
	if err := er.db.Conn.
		Where("app_name = ? AND version = ? AND os = ?", appName, version, osName).
		First(&exportZip).Error; err != nil {
		if errors.IsRecordNotFound(err) {
			return nil, exports_errors.NewExportNotFound("export zip not found")
		}
		return nil, err
	}
	return &exportZip, nil
}
