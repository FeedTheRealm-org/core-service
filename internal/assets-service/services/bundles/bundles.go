package models

import (
	"fmt"
	"io"
	"mime/multipart"
	"os"

	"github.com/FeedTheRealm-org/core-service/config"
	"github.com/FeedTheRealm-org/core-service/internal/assets-service/models"
	repo "github.com/FeedTheRealm-org/core-service/internal/assets-service/repositories/bundles"
	"github.com/google/uuid"
)

type bundlesService struct {
	conf              *config.Config
	bundlesRepository repo.BundlesRepository
}

// NewBundlesService creates a new instance of BundleService.
func NewBundlesService(conf *config.Config, bundlesRepository repo.BundlesRepository) BundleService {
	return &bundlesService{
		conf:              conf,
		bundlesRepository: bundlesRepository,
	}
}

func (bs *bundlesService) PublishWorldBundle(bundle models.Bundle) (models.Bundle, error) {
	if bundle.BundleFile == nil {
		return models.Bundle{}, fmt.Errorf("bundle file is required")
	}
	// TODO: this should be replaced with proper object storage
	bundleDir := fmt.Sprintf("bucket/bundles/%s", bundle.WorldID.String())
	if err := os.MkdirAll(bundleDir, os.ModePerm); err != nil {
		return models.Bundle{}, fmt.Errorf("failed to create bundle directory: %w", err)
	}

	bundlePath := fmt.Sprintf("%s/%s", bundleDir, bundle.BundleFile.Filename)
	if err := uploadedFile(bundle.BundleFile, bundlePath); err != nil {
		return models.Bundle{}, err
	}
	bundle.BundleURL = bundlePath
	bundle.BundleFile = nil

	publishedBundle, err := bs.bundlesRepository.PublishWorldBundle(bundle)
	if err != nil {
		return models.Bundle{}, err
	}

	return publishedBundle, nil
}

func (bs *bundlesService) GetWorldBundle(worldId uuid.UUID) (models.Bundle, error) {
	bundle, err := bs.bundlesRepository.GetWorldBundle(worldId)
	if err != nil {
		return models.Bundle{}, err
	}
	return bundle, nil
}

func uploadedFile(fileHeader *multipart.FileHeader, path string) error {
	in, err := fileHeader.Open()
	if err != nil {
		return err
	}
	defer func() {
		_ = in.Close()
	}()

	out, err := os.Create(path)
	if err != nil {
		return err
	}
	defer func() {
		_ = out.Close()
	}()

	_, err = io.Copy(out, in)
	return err
}
