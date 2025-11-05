package errors

import (
	"errors"

	"gorm.io/gorm"
)

func IsRecordNotFound(err error) bool {
	return errors.Is(err, gorm.ErrRecordNotFound)
}
