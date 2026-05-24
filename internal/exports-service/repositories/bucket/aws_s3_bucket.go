package bucket

import (
	"context"
	"mime/multipart"

	"github.com/FeedTheRealm-org/core-service/config"
	"github.com/FeedTheRealm-org/core-service/internal/utils/logger"

	"github.com/aws/aws-sdk-go-v2/aws"
	aws_config "github.com/aws/aws-sdk-go-v2/config"
	aws_s3 "github.com/aws/aws-sdk-go-v2/service/s3"
)

type awsS3BucketRepository struct {
	bucketName  string
	conf        *config.Config
	awsS3Client *aws_s3.Client
}

// NewAwsS3BucketRepository creates a new instance of the bucket repository connected to AWS S3.
func NewAwsS3BucketRepository(bucketName string, conf *config.Config) (BucketRepository, error) {
	r := &awsS3BucketRepository{
		conf: conf,
	}

	ctx := context.Background()

	awsCfg, err := aws_config.LoadDefaultConfig(ctx, aws_config.WithRegion("us-east-2"))
	if err != nil {
		logger.Logger.Errorf("Error loading AWS config: %v", err)
		return nil, err
	}

	logger.Logger.Info("AWS S3 client configured successfully")

	r.awsS3Client = aws_s3.NewFromConfig(awsCfg)
	r.bucketName = bucketName

	return r, nil
}

func (r *awsS3BucketRepository) GetBaseUrl() string {
	// TODO: Determinate what base_url use AWS S3
	return ""
}

func (r *awsS3BucketRepository) UploadFile(filePath, mimeType string, file multipart.File) error {
	logger.Logger.Infof("Uploading file to AWS S3: bucket=%s, filePath=%s", r.bucketName, filePath)

	input := &aws_s3.PutObjectInput{
		Bucket:      &r.bucketName,
		Key:         &filePath,
		Body:        file,
		ContentType: aws.String(mimeType),
	}

	logger.Logger.Debugf("PutObjectInput: Bucket=%s, Key=%s, ContentType=%s", r.bucketName, filePath, mimeType)

	_, err := r.awsS3Client.PutObject(context.Background(), input)
	if err != nil {
		logger.Logger.Errorf("Error uploading file to AWS S3: %v", err)
		return err
	}

	return nil
}

func (r *awsS3BucketRepository) DeleteFile(filePath string) error {
	input := &aws_s3.DeleteObjectInput{
		Bucket: &r.bucketName,
		Key:    &filePath,
	}

	_, err := r.awsS3Client.DeleteObject(context.Background(), input)
	if err != nil {
		return err
	}

	return nil
}
