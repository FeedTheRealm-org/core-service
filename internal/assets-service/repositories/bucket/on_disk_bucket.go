package bucket

import (
	"mime/multipart"

	"github.com/FeedTheRealm-org/core-service/config"
)

type onDiskBucketRepository struct {
	conf *config.Config
}

// NewItemSpritesRepository creates a new instance of ItemSpritesRepository.
func NewOnDiskBucketRepository(conf *config.Config) BucketRepository {
	return &onDiskBucketRepository{
		conf: conf,
	}
}

func (r *onDiskBucketRepository) UploadFile(fileName string, file multipart.File) error {
	return nil
}

func (r *onDiskBucketRepository) DownloadFile(fileName string) multipart.File {
	return nil
}
