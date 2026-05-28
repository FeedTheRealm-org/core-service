package items_test

import (
	"bytes"
	"mime/multipart"
	"os"
	"testing"

	"github.com/FeedTheRealm-org/core-service/config"
	"github.com/FeedTheRealm-org/core-service/internal/assets-service/models"
	itemservice "github.com/FeedTheRealm-org/core-service/internal/assets-service/services/items"
	"github.com/FeedTheRealm-org/core-service/internal/utils/logger"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	logger.InitLogger(false)
	os.Exit(m.Run())
}

type fakeItemRepo struct {
	upsertErr error
	getErr    error
	deleteErr error
	items     map[uuid.UUID]*models.Item
}

func (f *fakeItemRepo) UpsertItem(item *models.Item) error {
	if f.upsertErr != nil {
		return f.upsertErr
	}
	if f.items == nil {
		f.items = map[uuid.UUID]*models.Item{}
	}
	f.items[item.Id] = item
	return nil
}

func (f *fakeItemRepo) GetItemById(id uuid.UUID) (*models.Item, error) {
	if f.getErr != nil {
		return nil, f.getErr
	}
	item, ok := f.items[id]
	if !ok {
		return nil, assert.AnError
	}
	return item, nil
}

func (f *fakeItemRepo) GetAllItems() ([]*models.Item, error) {
	var items []*models.Item
	for _, item := range f.items {
		items = append(items, item)
	}
	return items, nil
}

func (f *fakeItemRepo) GetItemsListByWorld(worldId uuid.UUID) ([]*models.Item, error) {
	var items []*models.Item
	for _, item := range f.items {
		if item.WorldID == worldId {
			items = append(items, item)
		}
	}
	return items, nil
}

func (f *fakeItemRepo) DeleteSprite(id uuid.UUID) error {
	if f.deleteErr != nil {
		return f.deleteErr
	}
	delete(f.items, id)
	return nil
}

type fakeBucketRepo struct {
	uploadFn func(fileName, mimeType string, file multipart.File) error
	deleteFn func(fileName string) error
}

func (f *fakeBucketRepo) GetBaseUrl() string { return "" }

func (f *fakeBucketRepo) UploadFile(fileName, mimeType string, file multipart.File) error {
	if f.uploadFn != nil {
		return f.uploadFn(fileName, mimeType, file)
	}
	return nil
}

func (f *fakeBucketRepo) DeleteFile(fileName string) error {
	if f.deleteFn != nil {
		return f.deleteFn(fileName)
	}
	return nil
}

func createFileHeader(t *testing.T, filename, contentType string, data []byte) *multipart.FileHeader {
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

func TestItemService_UploadSprite_Success(t *testing.T) {
	worldID := uuid.New()
	itemID := uuid.New()
	userID := uuid.New()

	repo := &fakeItemRepo{items: map[uuid.UUID]*models.Item{}}
	var uploadedPath string
	bucket := &fakeBucketRepo{
		uploadFn: func(fileName, mimeType string, file multipart.File) error {
			uploadedPath = fileName
			return nil
		},
	}
	conf := &config.Config{Assets: &config.AssetsConfig{MaxUploadSizeBytes: 1024}}
	service := itemservice.NewItemService(conf, repo, bucket)

	fileHeader := createFileHeader(t, "sprite.png", "image/png", []byte("data"))

	item, err := service.UploadSprite(worldID, itemID, fileHeader, userID)
	assert.NoError(t, err)
	assert.NotNil(t, item)
	assert.Equal(t, "/worlds/"+worldID.String()+"/items/"+itemID.String()+".png", item.Url)
	assert.Equal(t, "worlds/"+worldID.String()+"/items/"+itemID.String()+".png", uploadedPath)
}

func TestItemService_UploadSprite_InvalidContentType(t *testing.T) {
	worldID := uuid.New()
	itemID := uuid.New()
	userID := uuid.New()

	repo := &fakeItemRepo{items: map[uuid.UUID]*models.Item{}}
	bucket := &fakeBucketRepo{}
	service := itemservice.NewItemService(nil, repo, bucket)

	fileHeader := createFileHeader(t, "sprite.gif", "image/gif", []byte("data"))

	item, err := service.UploadSprite(worldID, itemID, fileHeader, userID)
	assert.Error(t, err)
	assert.Nil(t, item)
}

func TestItemService_UploadSprite_TooLarge(t *testing.T) {
	worldID := uuid.New()
	itemID := uuid.New()
	userID := uuid.New()

	repo := &fakeItemRepo{items: map[uuid.UUID]*models.Item{}}
	bucket := &fakeBucketRepo{}
	conf := &config.Config{Assets: &config.AssetsConfig{MaxUploadSizeBytes: 1}}
	service := itemservice.NewItemService(conf, repo, bucket)

	fileHeader := createFileHeader(t, "sprite.png", "image/png", []byte("toolarge"))

	item, err := service.UploadSprite(worldID, itemID, fileHeader, userID)
	assert.Error(t, err)
	assert.Nil(t, item)
}

func TestItemService_DeleteItem_BucketDeleteError(t *testing.T) {
	itemID := uuid.New()
	worldID := uuid.New()

	repo := &fakeItemRepo{items: map[uuid.UUID]*models.Item{}}
	repo.items[itemID] = &models.Item{Id: itemID, WorldID: worldID, Url: "/worlds/x/items/y.png"}
	bucket := &fakeBucketRepo{
		deleteFn: func(fileName string) error {
			return assert.AnError
		},
	}
	service := itemservice.NewItemService(nil, repo, bucket)

	err := service.DeleteItem(itemID)
	assert.NoError(t, err)
	_, err = repo.GetItemById(itemID)
	assert.Error(t, err)
}

func TestItemService_UploadSprite_BucketUploadError(t *testing.T) {
	worldID := uuid.New()
	itemID := uuid.New()
	userID := uuid.New()

	repo := &fakeItemRepo{items: map[uuid.UUID]*models.Item{}}
	bucket := &fakeBucketRepo{
		uploadFn: func(fileName, mimeType string, file multipart.File) error {
			return assert.AnError
		},
	}
	service := itemservice.NewItemService(nil, repo, bucket)

	fileHeader := createFileHeader(t, "sprite.png", "image/png", []byte("data"))
	item, err := service.UploadSprite(worldID, itemID, fileHeader, userID)
	assert.Error(t, err)
	assert.Nil(t, item)
}

func TestItemService_UploadSprite_UpsertError(t *testing.T) {
	worldID := uuid.New()
	itemID := uuid.New()
	userID := uuid.New()

	repo := &fakeItemRepo{items: map[uuid.UUID]*models.Item{}, upsertErr: assert.AnError}
	bucket := &fakeBucketRepo{}
	service := itemservice.NewItemService(nil, repo, bucket)

	fileHeader := createFileHeader(t, "sprite.png", "image/png", []byte("data"))
	item, err := service.UploadSprite(worldID, itemID, fileHeader, userID)
	assert.Error(t, err)
	assert.Nil(t, item)
}

func TestItemService_DeleteItem_RepoDeleteError(t *testing.T) {
	itemID := uuid.New()
	worldID := uuid.New()

	repo := &fakeItemRepo{items: map[uuid.UUID]*models.Item{}, deleteErr: assert.AnError}
	repo.items[itemID] = &models.Item{Id: itemID, WorldID: worldID, Url: "/worlds/x/items/y.png"}
	bucket := &fakeBucketRepo{}
	service := itemservice.NewItemService(nil, repo, bucket)

	err := service.DeleteItem(itemID)
	assert.Error(t, err)
}
