package bucket

import (
	"mime/multipart"
)

// BucketRepository defines the interface for bucket operations.
type BucketRepository interface {
	UploadFile(fileName, mimeType string, file multipart.File) error

	DeleteFile(fileName string) error
}
