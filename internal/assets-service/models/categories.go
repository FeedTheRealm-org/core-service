package models

import (
	"time"

	"github.com/google/uuid"
)

type Category[T any] struct {
	Id        uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	Name      string    `gorm:"not null;unique"`
	Assets    []T       `gorm:"many2many:categories;"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}

func (Category[T]) TableName() string {
	return "categories"
}

type ItemCategory = Category[Item]
type ModelCategory = Category[Model]
