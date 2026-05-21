package models

import (
	"time"

	"github.com/google/uuid"
)

type PasswordReset struct {
	Id                  uuid.UUID  `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	UserId              uuid.UUID  `gorm:"column:user_id;not null;index"`
	OTPHash             string     `gorm:"column:otp_hash;not null"`
	ResetTokenHash      string     `gorm:"column:reset_token_hash"`
	Attempts            int        `gorm:"not null;default:0"`
	OTPExpiresAt        time.Time  `gorm:"column:otp_expires_at;not null"`
	ResetTokenExpiresAt *time.Time `gorm:"column:reset_token_expires_at"`
	OTPVerified         bool       `gorm:"column:otp_verified;not null;default:false"`
	Used                bool       `gorm:"not null;default:false"`
	CreatedAt           time.Time  `gorm:"autoCreateTime"`

	User User `gorm:"foreignKey:UserId"`
}
