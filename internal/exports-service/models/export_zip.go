package models

import (
	"time"

	"github.com/google/uuid"
)

type ExportZip struct {
	Id          uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	AppName     string    `gorm:"not null;index:idx_exports_app_version_os,unique"`
	Version     string    `gorm:"not null;index:idx_exports_app_version_os,unique"`
	OS          string    `gorm:"not null;index:idx_exports_app_version_os,unique"`
	Path        string    `gorm:"not null"`
	ReleaseNote string    `gorm:"not null;default:'no release note provided.'"`
	IsLatest    bool      `gorm:"not null;default:false"`
	CreatedAt   time.Time `gorm:"autoCreateTime"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime"`
}

func (ExportZip) TableName() string {
	return "exports_versions"
}
