package bucket

import (
	"fmt"
	"mime/multipart"
	"os"
	"path/filepath"

	"github.com/FeedTheRealm-org/core-service/config"
)

const localBucketFolder = "./local_buckets"

type onDiskBucketRepository struct {
	bucketName string
	bucketPath string
	conf       *config.Config
}

// NewOnDiskBucketRepository creates a new instance of the bucket repository connected to on-disk storage.
func NewOnDiskBucketRepository(bucketName string, conf *config.Config) (BucketRepository, error) {
	r := &onDiskBucketRepository{
		bucketName: bucketName,
		bucketPath: fmt.Sprintf("%s/%s", localBucketFolder, bucketName),
		conf:       conf,
	}

	if err := os.MkdirAll(r.bucketPath, os.ModePerm); err != nil {
		return nil, err
	}

	return r, nil
}

func (r *onDiskBucketRepository) UploadFile(filePath, mimeType string, file multipart.File) error {
	destPath := filepath.Join(r.bucketPath, filePath)

	destFile, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer destFile.Close()

	return nil
}

func (r *onDiskBucketRepository) DeleteFile(filePath string) error {
	destPath := filepath.Join(r.bucketPath, filePath)

	if err := os.Remove(destPath); err != nil {
		return err
	}

	return nil
}
