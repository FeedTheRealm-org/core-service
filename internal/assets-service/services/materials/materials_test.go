package materials_test

import (
	"bytes"
	"mime/multipart"
	"os"
	"testing"

	"github.com/FeedTheRealm-org/core-service/config"
	"github.com/FeedTheRealm-org/core-service/internal/assets-service/models"
	materialservice "github.com/FeedTheRealm-org/core-service/internal/assets-service/services/materials"
	"github.com/FeedTheRealm-org/core-service/internal/utils/logger"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	logger.InitLogger(false)
	os.Exit(m.Run())
}

type fakeMaterialsRepo struct {
	upsertErr error
	getErr    error
	deleteErr error
	items     map[uuid.UUID]*models.Material
}

func (f *fakeMaterialsRepo) UpsertMaterial(material *models.Material) error {
	if f.upsertErr != nil {
		return f.upsertErr
	}
	if f.items == nil {
		f.items = map[uuid.UUID]*models.Material{}
	}
	f.items[material.ID] = material
	return nil
}

func (f *fakeMaterialsRepo) GetMaterialByID(id uuid.UUID) (*models.Material, error) {
	if f.getErr != nil {
		return nil, f.getErr
	}
	item, ok := f.items[id]
	if !ok {
		return nil, assert.AnError
	}
	return item, nil
}

func (f *fakeMaterialsRepo) GetMaterialsListByWorld(worldID uuid.UUID, offset int, limit int) ([]*models.Material, error) {
	var materials []*models.Material
	for _, item := range f.items {
		if item.WorldID == worldID || item.WorldID == uuid.Nil {
			materials = append(materials, item)
		}
	}
	return materials, nil
}

func (f *fakeMaterialsRepo) GetMaterialsListByWorldAndType(worldID uuid.UUID, offset int, limit int) ([]*models.Material, error) {
	return f.GetMaterialsListByWorld(worldID, offset, limit)
}

func (f *fakeMaterialsRepo) DeleteMaterial(material *models.Material) error {
	if f.deleteErr != nil {
		return f.deleteErr
	}
	delete(f.items, material.ID)
	return nil
}

type fakeMaterialsBucketRepo struct {
	uploadFn func(fileName, mimeType string, file multipart.File) error
	deleteFn func(fileName string) error
}

func (f *fakeMaterialsBucketRepo) GetBaseUrl() string { return "" }

func (f *fakeMaterialsBucketRepo) UploadFile(fileName, mimeType string, file multipart.File) error {
	if f.uploadFn != nil {
		return f.uploadFn(fileName, mimeType, file)
	}
	return nil
}

func (f *fakeMaterialsBucketRepo) DeleteFile(fileName string) error {
	if f.deleteFn != nil {
		return f.deleteFn(fileName)
	}
	return nil
}

func createMaterialFileHeader(t *testing.T, filename, contentType string, data []byte) *multipart.FileHeader {
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

func TestMaterialsService_UploadMaterial_Success(t *testing.T) {
	worldID := uuid.New()
	materialID := uuid.New()
	userID := uuid.New()

	repo := &fakeMaterialsRepo{items: map[uuid.UUID]*models.Material{}}
	var uploadedPath string
	bucket := &fakeMaterialsBucketRepo{
		uploadFn: func(fileName, mimeType string, file multipart.File) error {
			uploadedPath = fileName
			return nil
		},
	}
	conf := &config.Config{Assets: &config.AssetsConfig{MaxUploadSizeBytes: 1024}}
	service := materialservice.NewMaterialsService(conf, repo, bucket)

	fileHeader := createMaterialFileHeader(t, "material.png", "image/png", []byte("data"))

	material, err := service.UploadMaterial(worldID, materialID, "name", fileHeader, userID)
	assert.NoError(t, err)
	assert.NotNil(t, material)
	assert.Equal(t, "/worlds/"+worldID.String()+"/materials/"+materialID.String()+".png", material.URL)
	assert.Equal(t, "worlds/"+worldID.String()+"/materials/"+materialID.String()+".png", uploadedPath)
}

func TestMaterialsService_UploadMaterial_InvalidContentType(t *testing.T) {
	worldID := uuid.New()
	materialID := uuid.New()
	userID := uuid.New()

	repo := &fakeMaterialsRepo{items: map[uuid.UUID]*models.Material{}}
	bucket := &fakeMaterialsBucketRepo{}
	service := materialservice.NewMaterialsService(nil, repo, bucket)

	fileHeader := createMaterialFileHeader(t, "material.gif", "image/gif", []byte("data"))

	material, err := service.UploadMaterial(worldID, materialID, "name", fileHeader, userID)
	assert.Error(t, err)
	assert.Nil(t, material)
}

func TestMaterialsService_UploadMaterial_TooLarge(t *testing.T) {
	worldID := uuid.New()
	materialID := uuid.New()
	userID := uuid.New()

	repo := &fakeMaterialsRepo{items: map[uuid.UUID]*models.Material{}}
	bucket := &fakeMaterialsBucketRepo{}
	conf := &config.Config{Assets: &config.AssetsConfig{MaxUploadSizeBytes: 1}}
	service := materialservice.NewMaterialsService(conf, repo, bucket)

	fileHeader := createMaterialFileHeader(t, "material.png", "image/png", []byte("toolarge"))

	material, err := service.UploadMaterial(worldID, materialID, "name", fileHeader, userID)
	assert.Error(t, err)
	assert.Nil(t, material)
}

func TestMaterialsService_DeleteMaterial_BucketDeleteError(t *testing.T) {
	materialID := uuid.New()
	worldID := uuid.New()

	repo := &fakeMaterialsRepo{items: map[uuid.UUID]*models.Material{}}
	repo.items[materialID] = &models.Material{ID: materialID, WorldID: worldID, URL: "/worlds/x/materials/y.png"}
	bucket := &fakeMaterialsBucketRepo{
		deleteFn: func(fileName string) error {
			return assert.AnError
		},
	}
	service := materialservice.NewMaterialsService(nil, repo, bucket)

	err := service.DeleteMaterial(materialID)
	assert.NoError(t, err)
	_, err = repo.GetMaterialByID(materialID)
	assert.Error(t, err)
}

func TestMaterialsService_UploadMaterial_BucketUploadError(t *testing.T) {
	worldID := uuid.New()
	materialID := uuid.New()
	userID := uuid.New()

	repo := &fakeMaterialsRepo{items: map[uuid.UUID]*models.Material{}}
	bucket := &fakeMaterialsBucketRepo{
		uploadFn: func(fileName, mimeType string, file multipart.File) error {
			return assert.AnError
		},
	}
	service := materialservice.NewMaterialsService(nil, repo, bucket)

	fileHeader := createMaterialFileHeader(t, "material.png", "image/png", []byte("data"))
	material, err := service.UploadMaterial(worldID, materialID, "name", fileHeader, userID)
	assert.Error(t, err)
	assert.Nil(t, material)
}

func TestMaterialsService_UploadMaterial_UpsertError(t *testing.T) {
	worldID := uuid.New()
	materialID := uuid.New()
	userID := uuid.New()

	repo := &fakeMaterialsRepo{items: map[uuid.UUID]*models.Material{}, upsertErr: assert.AnError}
	bucket := &fakeMaterialsBucketRepo{}
	service := materialservice.NewMaterialsService(nil, repo, bucket)

	fileHeader := createMaterialFileHeader(t, "material.png", "image/png", []byte("data"))
	material, err := service.UploadMaterial(worldID, materialID, "name", fileHeader, userID)
	assert.Error(t, err)
	assert.Nil(t, material)
}

func TestMaterialsService_DeleteMaterial_RepoDeleteError(t *testing.T) {
	materialID := uuid.New()
	worldID := uuid.New()

	repo := &fakeMaterialsRepo{items: map[uuid.UUID]*models.Material{}, deleteErr: assert.AnError}
	repo.items[materialID] = &models.Material{ID: materialID, WorldID: worldID, URL: "/worlds/x/materials/y.png"}
	bucket := &fakeMaterialsBucketRepo{}
	service := materialservice.NewMaterialsService(nil, repo, bucket)

	err := service.DeleteMaterial(materialID)
	assert.Error(t, err)
}
