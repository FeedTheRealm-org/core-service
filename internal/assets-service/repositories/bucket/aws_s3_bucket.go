package bucket

import (
	"mime/multipart"

	"github.com/FeedTheRealm-org/core-service/config"
)

type awsS3BucketRepository struct {
	conf *config.Config
}

// NewItemSpritesRepository creates a new instance of ItemSpritesRepository.
func NewAwsS3BucketRepository(conf *config.Config) BucketRepository {
	return &awsS3BucketRepository{
		conf: conf,
	}
}

func (r *awsS3BucketRepository) UploadFile(fileName string, file multipart.File) error {
	return nil
}

func (r *awsS3BucketRepository) DownloadFile(fileName string) multipart.File {
	return nil
}
