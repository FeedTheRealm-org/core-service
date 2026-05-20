package exports

import (
	"time"

	"github.com/FeedTheRealm-org/core-service/config"
	"github.com/FeedTheRealm-org/core-service/internal/errors"
	exports_errors "github.com/FeedTheRealm-org/core-service/internal/exports-service/errors"
	"github.com/FeedTheRealm-org/core-service/internal/exports-service/models"
	"gorm.io/gorm"
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

func (er *exportRepository) GetLatestExportVersion(appName, osName string) (*models.ExportZip, error) {
	var exportZip models.ExportZip
	if err := er.db.Conn.
		Where("app_name = ? AND os = ? AND is_latest = ?", appName, osName, true).
		Order("created_at DESC").
		First(&exportZip).Error; err == nil {
		return &exportZip, nil
	} else if !errors.IsRecordNotFound(err) {
		return nil, err
	}

	if err := er.db.Conn.
		Where("app_name = ? AND os = ?", appName, osName).
		Order("created_at DESC").
		First(&exportZip).Error; err != nil {
		if errors.IsRecordNotFound(err) {
			return nil, exports_errors.NewExportNotFound("export zip not found")
		}
		return nil, err
	}

	return &exportZip, nil
}

func (er *exportRepository) ListExportVersions(appName, osName string) ([]*models.ExportZip, error) {
	var exportZips []*models.ExportZip
	query := er.db.Conn.Order("created_at DESC")
	if appName != "" {
		query = query.Where("app_name = ?", appName)
	}
	if osName != "" {
		query = query.Where("os = ?", osName)
	}
	if err := query.Find(&exportZips).Error; err != nil {
		return nil, err
	}
	return exportZips, nil
}

func (er *exportRepository) DeleteExportVersion(appName, version, osName string) error {
	return er.db.Conn.Transaction(func(tx *gorm.DB) error {
		var exportZip models.ExportZip
		if err := tx.Where("app_name = ? AND version = ? AND os = ?", appName, version, osName).First(&exportZip).Error; err != nil {
			if errors.IsRecordNotFound(err) {
				return exports_errors.NewExportNotFound("export zip not found")
			}
			return err
		}

		if err := tx.Delete(&exportZip).Error; err != nil {
			return err
		}

		if exportZip.IsLatest {
			if err := tx.Model(&models.ExportZip{}).
				Where("app_name = ? AND os = ?", appName, osName).
				Update("is_latest", false).Error; err != nil {
				return err
			}

			var replacement models.ExportZip
			if err := tx.Where("app_name = ? AND os = ?", appName, osName).
				Order("created_at DESC").
				First(&replacement).Error; err == nil {
				if err := tx.Model(&models.ExportZip{}).Where("id = ?", replacement.Id).Updates(map[string]interface{}{
					"is_latest":  true,
					"updated_at": time.Now().UTC(),
				}).Error; err != nil {
					return err
				}
			}
		}

		return nil
	})
}

func (er *exportRepository) SetLatestExportVersion(appName, version, osName string) (*models.ExportZip, error) {
	var updatedExport *models.ExportZip
	err := er.db.Conn.Transaction(func(tx *gorm.DB) error {
		var exportZip models.ExportZip
		if err := tx.Where("app_name = ? AND version = ? AND os = ?", appName, version, osName).First(&exportZip).Error; err != nil {
			if errors.IsRecordNotFound(err) {
				return exports_errors.NewExportNotFound("export zip not found")
			}
			return err
		}

		if err := tx.Model(&models.ExportZip{}).
			Where("app_name = ? AND os = ?", appName, osName).
			Update("is_latest", false).Error; err != nil {
			return err
		}

		if err := tx.Model(&models.ExportZip{}).Where("id = ?", exportZip.Id).Updates(map[string]interface{}{
			"is_latest":  true,
			"updated_at": time.Now().UTC(),
		}).Error; err != nil {
			return err
		}

		exportZip.IsLatest = true
		updatedExport = &exportZip
		return nil
	})
	if err != nil {
		return nil, err
	}
	return updatedExport, nil
}
