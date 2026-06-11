package models_test

import (
	"bytes"
	"mime/multipart"
	"os"
	"testing"

	"github.com/FeedTheRealm-org/core-service/internal/assets-service/dtos"
	"github.com/FeedTheRealm-org/core-service/internal/assets-service/models"
	modelservice "github.com/FeedTheRealm-org/core-service/internal/assets-service/services/models"
	"github.com/FeedTheRealm-org/core-service/internal/utils/logger"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	logger.InitLogger(false)
	os.Exit(m.Run())
}

type fakeModelsRepo struct {
	uploadFn func(models.Model) (*models.Model, error)
	listFn   func(uuid.UUID) ([]models.Model, error)
}

func (f *fakeModelsRepo) UploadModel(model models.Model) (*models.Model, error) {
	if f.uploadFn == nil {
		return &model, nil
	}
	return f.uploadFn(model)
}

func (f *fakeModelsRepo) GetModelsByWorld(worldId uuid.UUID) ([]models.Model, error) {
	if f.listFn == nil {
		return nil, nil
	}
	return f.listFn(worldId)
}

type fakeModelsBucketRepo struct {
	uploadFn func(fileName, mimeType string, file multipart.File) error
	deleteFn func(fileName string) error
}

func (f *fakeModelsBucketRepo) GetBaseUrl() string { return "" }

func (f *fakeModelsBucketRepo) UploadFile(fileName, mimeType string, file multipart.File) error {
	if f.uploadFn != nil {
		return f.uploadFn(fileName, mimeType, file)
	}
	return nil
}

func (f *fakeModelsBucketRepo) DeleteFile(fileName string) error {
	if f.deleteFn != nil {
		return f.deleteFn(fileName)
	}
	return nil
}

func createModelFileHeader(t *testing.T, filename, contentType string, data []byte) *multipart.FileHeader {
	t.Helper()

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		t.Fatalf("create form file: %v", err)
	}
	if _, err := part.Write(data); err != nil {
		t.Fatalf("write data: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close writer: %v", err)
	}

	reader := multipart.NewReader(&buf, writer.Boundary())
	form, err := reader.ReadForm(int64(len(data)) + 1024)
	if err != nil {
		t.Fatalf("read form: %v", err)
	}
	files := form.File["file"]
	if len(files) != 1 {
		t.Fatalf("expected 1 file header, got %d", len(files))
	}
	files[0].Header.Set("Content-Type", contentType)
	return files[0]
}

func TestModelsService_UploadModel_InvalidID(t *testing.T) {
	repo := &fakeModelsRepo{}
	bucket := &fakeModelsBucketRepo{}
	service := modelservice.NewModelsService(nil, repo, bucket)

	request := dtos.ModelRequest{Id: uuid.Nil}
	model, err := service.UploadModel(request)
	assert.Error(t, err)
	assert.Nil(t, model)
}

func TestModelsService_UploadModel_InvalidExtension(t *testing.T) {
	repo := &fakeModelsRepo{}
	bucket := &fakeModelsBucketRepo{}
	service := modelservice.NewModelsService(nil, repo, bucket)

	request := dtos.ModelRequest{
		Id:        uuid.New(),
		WorldID:   uuid.New(),
		CreatedBy: uuid.New(),
		ModelFile: createModelFileHeader(t, "model.txt", "text/plain", []byte("data")),
	}
	model, err := service.UploadModel(request)
	assert.Error(t, err)
	assert.Nil(t, model)
}

func TestModelsService_UploadModel_SuccessAndSanitize(t *testing.T) {
	worldID := uuid.New()
	modelID := uuid.New()

	repo := &fakeModelsRepo{}
	var uploadedPath string
	bucket := &fakeModelsBucketRepo{
		uploadFn: func(fileName, mimeType string, file multipart.File) error {
			uploadedPath = fileName
			return nil
		},
	}
	service := modelservice.NewModelsService(nil, repo, bucket)

	request := dtos.ModelRequest{
		Id:        modelID,
		WorldID:   worldID,
		CreatedBy: uuid.New(),
		ModelFile: createModelFileHeader(t, "my model@.glb", "model/gltf-binary", []byte("data")),
	}
	model, err := service.UploadModel(request)
	assert.NoError(t, err)
	assert.NotNil(t, model)
	assert.Contains(t, uploadedPath, "my_model_.glb")
}

func TestModelsService_UploadModel_RepoErrorDeletesFile(t *testing.T) {
	worldID := uuid.New()
	modelID := uuid.New()
	deletedPath := ""

	repo := &fakeModelsRepo{
		uploadFn: func(model models.Model) (*models.Model, error) {
			return nil, assert.AnError
		},
	}
	bucket := &fakeModelsBucketRepo{
		deleteFn: func(fileName string) error {
			deletedPath = fileName
			return nil
		},
	}
	service := modelservice.NewModelsService(nil, repo, bucket)

	request := dtos.ModelRequest{
		Id:        modelID,
		WorldID:   worldID,
		CreatedBy: uuid.New(),
		ModelFile: createModelFileHeader(t, "model.glb", "model/gltf-binary", []byte("data")),
	}
	model, err := service.UploadModel(request)
	assert.Error(t, err)
	assert.Nil(t, model)
	assert.Contains(t, deletedPath, "worlds/"+worldID.String()+"/models/"+modelID.String()+"/model.glb")
}
