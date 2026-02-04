package bucket

import (
	"mime/multipart"
)

// BucketRepository defines the interface for bucket operations.
type BucketRepository interface {
	UploadFile(fileName string, file multipart.File) error

	DownloadFile(fileName string) multipart.File
}
