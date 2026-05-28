package cosmetics_test

import (
	"bytes"
	"mime/multipart"
	"os"
	"testing"

	assets_errors "github.com/FeedTheRealm-org/core-service/internal/assets-service/errors"
	"github.com/FeedTheRealm-org/core-service/internal/assets-service/models"
	cosmeticservice "github.com/FeedTheRealm-org/core-service/internal/assets-service/services/cosmetics"
	"github.com/FeedTheRealm-org/core-service/internal/utils/logger"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	logger.InitLogger(false)
	os.Exit(m.Run())
}

type fakeCosmeticsRepo struct {
	getCategoryByIdFn               func(uuid.UUID) (*models.CosmeticCategory, error)
	getCosmeticsListByCategoryFn    func(uuid.UUID, *uuid.UUID, *uuid.UUID, int, int) ([]*models.Cosmetic, int64, error)
	getEconomySummaryFn             func() (*models.CosmeticsEconomySummary, error)
	getCosmeticByIdFn               func(uuid.UUID) (*models.Cosmetic, error)
	getCosmeticsListByWorldFn       func(uuid.UUID, int, int) ([]*models.Cosmetic, int64, error)
	addCategoryFn                   func(string) (*models.CosmeticCategory, error)
	addPurchaseForUserIdFn          func(uuid.UUID, uuid.UUID) error
	createCosmeticFn                func(uuid.UUID, uuid.UUID, int64, *models.Cosmetic, uuid.UUID) error
	getCosmeticByUrlCategoryWorldFn func(string, uuid.UUID, uuid.UUID) (*models.Cosmetic, error)
	updateCosmeticFn                func(uuid.UUID, int64, string) error
	deleteCosmeticFn                func(uuid.UUID) error
}

func (f *fakeCosmeticsRepo) GetCategoriesList() ([]*models.CosmeticCategory, error) {
	return nil, nil
}

func (f *fakeCosmeticsRepo) GetCosmeticsListByCategory(category uuid.UUID, worldId *uuid.UUID, playerId *uuid.UUID, offset int, limit int) ([]*models.Cosmetic, int64, error) {
	if f.getCosmeticsListByCategoryFn == nil {
		panic("GetCosmeticsListByCategory not set")
	}
	return f.getCosmeticsListByCategoryFn(category, worldId, playerId, offset, limit)
}

func (f *fakeCosmeticsRepo) GetEconomySummary() (*models.CosmeticsEconomySummary, error) {
	if f.getEconomySummaryFn == nil {
		return nil, nil
	}
	return f.getEconomySummaryFn()
}

func (f *fakeCosmeticsRepo) GetCosmeticById(cosmeticId uuid.UUID) (*models.Cosmetic, error) {
	if f.getCosmeticByIdFn == nil {
		panic("GetCosmeticById not set")
	}
	return f.getCosmeticByIdFn(cosmeticId)
}

func (f *fakeCosmeticsRepo) GetCosmeticsListByWorld(worldId uuid.UUID, offset int, limit int) ([]*models.Cosmetic, int64, error) {
	if f.getCosmeticsListByWorldFn == nil {
		panic("GetCosmeticsListByWorld not set")
	}
	return f.getCosmeticsListByWorldFn(worldId, offset, limit)
}

func (f *fakeCosmeticsRepo) AddCategory(category string) (*models.CosmeticCategory, error) {
	if f.addCategoryFn == nil {
		panic("AddCategory not set")
	}
	return f.addCategoryFn(category)
}

func (f *fakeCosmeticsRepo) AddPurchaseForUserId(cosmeticId uuid.UUID, userId uuid.UUID) error {
	if f.addPurchaseForUserIdFn == nil {
		panic("AddPurchaseForUserId not set")
	}
	return f.addPurchaseForUserIdFn(cosmeticId, userId)
}

func (f *fakeCosmeticsRepo) GetCategoryById(categoryId uuid.UUID) (*models.CosmeticCategory, error) {
	if f.getCategoryByIdFn == nil {
		panic("GetCategoryById not set")
	}
	return f.getCategoryByIdFn(categoryId)
}

func (f *fakeCosmeticsRepo) CreateCosmetic(categoryId uuid.UUID, worldId uuid.UUID, price int64, cosmetic *models.Cosmetic, userId uuid.UUID) error {
	if f.createCosmeticFn == nil {
		panic("CreateCosmetic not set")
	}
	return f.createCosmeticFn(categoryId, worldId, price, cosmetic, userId)
}

func (f *fakeCosmeticsRepo) GetCosmeticByUrlCategoryAndWorld(url string, categoryId uuid.UUID, worldId uuid.UUID) (*models.Cosmetic, error) {
	if f.getCosmeticByUrlCategoryWorldFn == nil {
		panic("GetCosmeticByUrlCategoryAndWorld not set")
	}
	return f.getCosmeticByUrlCategoryWorldFn(url, categoryId, worldId)
}

func (f *fakeCosmeticsRepo) UpdateCosmetic(cosmeticId uuid.UUID, price int64, url string) error {
	if f.updateCosmeticFn == nil {
		panic("UpdateCosmetic not set")
	}
	return f.updateCosmeticFn(cosmeticId, price, url)
}

func (f *fakeCosmeticsRepo) DeleteCosmetic(cosmeticId uuid.UUID) error {
	if f.deleteCosmeticFn == nil {
		panic("DeleteCosmetic not set")
	}
	return f.deleteCosmeticFn(cosmeticId)
}

type fakeCosmeticsBucketRepo struct {
	uploadFn func(fileName, mimeType string, file multipart.File) error
	deleteFn func(fileName string) error
}

func (f *fakeCosmeticsBucketRepo) GetBaseUrl() string { return "" }

func (f *fakeCosmeticsBucketRepo) UploadFile(fileName, mimeType string, file multipart.File) error {
	if f.uploadFn != nil {
		return f.uploadFn(fileName, mimeType, file)
	}
	return nil
}

func (f *fakeCosmeticsBucketRepo) DeleteFile(fileName string) error {
	if f.deleteFn != nil {
		return f.deleteFn(fileName)
	}
	return nil
}

func createCosmeticFileHeader(t *testing.T, filename string, data []byte) multipart.File {
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

	file, err := files[0].Open()
	if err != nil {
		t.Fatalf("open file: %v", err)
	}
	return file
}

func TestCosmeticsService_GetCosmeticsListByCategory_CategoryNotFound(t *testing.T) {
	repo := &fakeCosmeticsRepo{
		getCategoryByIdFn: func(id uuid.UUID) (*models.CosmeticCategory, error) {
			return nil, assets_errors.NewCategoryNotFound("category not found")
		},
	}
	bucket := &fakeCosmeticsBucketRepo{}
	service := cosmeticservice.NewCosmeticsService(nil, repo, bucket)

	cosmetics, total, err := service.GetCosmeticsListByCategory(uuid.New(), nil, nil, 0, 10)
	assert.Error(t, err)
	assert.Nil(t, cosmetics)
	assert.Equal(t, int64(0), total)
}

func TestCosmeticsService_UploadCosmeticData_InvalidPrice(t *testing.T) {
	repo := &fakeCosmeticsRepo{}
	bucket := &fakeCosmeticsBucketRepo{}
	service := cosmeticservice.NewCosmeticsService(nil, repo, bucket)

	cosmetic, err := service.UploadCosmeticData(uuid.New(), uuid.New(), 0, nil, ".png", uuid.New())
	assert.Error(t, err)
	assert.Nil(t, cosmetic)
}

func TestCosmeticsService_UploadCosmeticData_Success(t *testing.T) {
	categoryId := uuid.New()
	worldId := uuid.New()
	userId := uuid.New()
	price := int64(10)

	repo := &fakeCosmeticsRepo{
		getCategoryByIdFn: func(id uuid.UUID) (*models.CosmeticCategory, error) {
			return &models.CosmeticCategory{Id: categoryId, Name: "hats"}, nil
		},
		createCosmeticFn: func(category uuid.UUID, world uuid.UUID, p int64, cosmetic *models.Cosmetic, createdBy uuid.UUID) error {
			return nil
		},
	}
	bucket := &fakeCosmeticsBucketRepo{}
	service := cosmeticservice.NewCosmeticsService(nil, repo, bucket)

	file := createCosmeticFileHeader(t, "cosmetic.png", []byte("data"))
	defer func() { _ = file.Close() }()
	cosmetic, err := service.UploadCosmeticData(categoryId, worldId, price, file, ".png", userId)
	assert.NoError(t, err)
	assert.NotNil(t, cosmetic)
	assert.Contains(t, cosmetic.Url, "/hats/")
}

func TestCosmeticsService_UploadCosmeticByID_UpdateExisting(t *testing.T) {
	categoryId := uuid.New()
	worldId := uuid.New()
	userId := uuid.New()
	price := int64(15)
	spriteId := uuid.New()

	updated := false
	repo := &fakeCosmeticsRepo{
		getCategoryByIdFn: func(id uuid.UUID) (*models.CosmeticCategory, error) {
			return &models.CosmeticCategory{Id: categoryId}, nil
		},
		getCosmeticByIdFn: func(id uuid.UUID) (*models.Cosmetic, error) {
			return &models.Cosmetic{Id: spriteId, Url: "/hats/one.png"}, nil
		},
		getCosmeticByUrlCategoryWorldFn: func(url string, category uuid.UUID, world uuid.UUID) (*models.Cosmetic, error) {
			return &models.Cosmetic{Id: uuid.New(), Url: url}, nil
		},
		updateCosmeticFn: func(cosmeticId uuid.UUID, p int64, url string) error {
			updated = true
			return nil
		},
	}
	bucket := &fakeCosmeticsBucketRepo{}
	service := cosmeticservice.NewCosmeticsService(nil, repo, bucket)

	cosmetic, err := service.UploadCosmeticByID(categoryId, worldId, price, spriteId, userId, nil, ".png")
	assert.NoError(t, err)
	assert.NotNil(t, cosmetic)
	assert.True(t, updated)
}

func TestCosmeticsService_DeleteCosmetic_BucketError(t *testing.T) {
	cosmeticId := uuid.New()

	repo := &fakeCosmeticsRepo{
		getCosmeticByIdFn: func(id uuid.UUID) (*models.Cosmetic, error) {
			return &models.Cosmetic{Id: cosmeticId, Url: "/path/file.png"}, nil
		},
		deleteCosmeticFn: func(id uuid.UUID) error {
			return nil
		},
	}
	bucket := &fakeCosmeticsBucketRepo{
		deleteFn: func(fileName string) error {
			return assert.AnError
		},
	}
	service := cosmeticservice.NewCosmeticsService(nil, repo, bucket)

	err := service.DeleteCosmetic(cosmeticId)
	assert.Error(t, err)
}

func TestCosmeticsService_PurchaseCosmeticForUserInternal_Error(t *testing.T) {
	repo := &fakeCosmeticsRepo{
		addPurchaseForUserIdFn: func(cosmeticId uuid.UUID, userId uuid.UUID) error {
			return assert.AnError
		},
	}
	bucket := &fakeCosmeticsBucketRepo{}
	service := cosmeticservice.NewCosmeticsService(nil, repo, bucket)

	err := service.PurchaseCosmeticForUserInternal(uuid.New(), uuid.New())
	assert.Error(t, err)
}

func TestCosmeticsService_GetCategoriesAndCosmeticById(t *testing.T) {
	categoryID := uuid.New()
	cosmeticID := uuid.New()

	repo := &fakeCosmeticsRepo{
		getCategoryByIdFn: func(id uuid.UUID) (*models.CosmeticCategory, error) {
			return &models.CosmeticCategory{Id: categoryID, Name: "hats"}, nil
		},
		getCosmeticByIdFn: func(id uuid.UUID) (*models.Cosmetic, error) {
			return &models.Cosmetic{Id: cosmeticID, Url: "/hats/one.png"}, nil
		},
		getCosmeticsListByWorldFn: func(worldId uuid.UUID, offset int, limit int) ([]*models.Cosmetic, int64, error) {
			return []*models.Cosmetic{{Id: cosmeticID, Url: "/hats/one.png"}}, 1, nil
		},
		addCategoryFn: func(category string) (*models.CosmeticCategory, error) {
			return &models.CosmeticCategory{Id: categoryID, Name: category}, nil
		},
	}
	bucket := &fakeCosmeticsBucketRepo{}
	service := cosmeticservice.NewCosmeticsService(nil, repo, bucket)

	category, err := service.AddCategory("hats")
	assert.NoError(t, err)
	assert.Equal(t, "hats", category.Name)

	cosmetic, err := service.GetCosmeticById(cosmeticID)
	assert.NoError(t, err)
	assert.Equal(t, "/hats/one.png", cosmetic.Url)

	list, total, err := service.GetCosmeticsListByWorld(uuid.New(), 0, 10)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, list, 1)
}

func TestCosmeticsService_UploadCosmeticData_CategoryError(t *testing.T) {
	repo := &fakeCosmeticsRepo{
		getCategoryByIdFn: func(id uuid.UUID) (*models.CosmeticCategory, error) {
			return nil, assert.AnError
		},
	}
	bucket := &fakeCosmeticsBucketRepo{}
	service := cosmeticservice.NewCosmeticsService(nil, repo, bucket)

	file := createCosmeticFileHeader(t, "cosmetic.png", []byte("data"))
	defer func() { _ = file.Close() }()
	cosmetic, err := service.UploadCosmeticData(uuid.New(), uuid.New(), 10, file, ".png", uuid.New())
	assert.Error(t, err)
	assert.Nil(t, cosmetic)
}

func TestCosmeticsService_UploadCosmeticData_BucketError(t *testing.T) {
	categoryID := uuid.New()
	repo := &fakeCosmeticsRepo{
		getCategoryByIdFn: func(id uuid.UUID) (*models.CosmeticCategory, error) {
			return &models.CosmeticCategory{Id: categoryID, Name: "hats"}, nil
		},
	}
	bucket := &fakeCosmeticsBucketRepo{
		uploadFn: func(fileName, mimeType string, file multipart.File) error {
			return assert.AnError
		},
	}
	service := cosmeticservice.NewCosmeticsService(nil, repo, bucket)

	file := createCosmeticFileHeader(t, "cosmetic.png", []byte("data"))
	defer func() { _ = file.Close() }()
	cosmetic, err := service.UploadCosmeticData(categoryID, uuid.New(), 10, file, ".png", uuid.New())
	assert.Error(t, err)
	assert.Nil(t, cosmetic)
}

func TestCosmeticsService_UploadCosmeticByID_CreateNew(t *testing.T) {
	categoryID := uuid.New()
	worldID := uuid.New()
	userID := uuid.New()
	price := int64(25)
	spriteID := uuid.New()

	created := false
	repo := &fakeCosmeticsRepo{
		getCategoryByIdFn: func(id uuid.UUID) (*models.CosmeticCategory, error) {
			return &models.CosmeticCategory{Id: categoryID}, nil
		},
		getCosmeticByIdFn: func(id uuid.UUID) (*models.Cosmetic, error) {
			return &models.Cosmetic{Id: spriteID, Url: "/hats/one.png"}, nil
		},
		getCosmeticByUrlCategoryWorldFn: func(url string, category uuid.UUID, world uuid.UUID) (*models.Cosmetic, error) {
			return nil, assets_errors.NewCosmeticNotFound("missing")
		},
		createCosmeticFn: func(category uuid.UUID, world uuid.UUID, p int64, cosmetic *models.Cosmetic, createdBy uuid.UUID) error {
			created = true
			return nil
		},
	}
	bucket := &fakeCosmeticsBucketRepo{}
	service := cosmeticservice.NewCosmeticsService(nil, repo, bucket)

	cosmetic, err := service.UploadCosmeticByID(categoryID, worldID, price, spriteID, userID, nil, ".png")
	assert.NoError(t, err)
	assert.NotNil(t, cosmetic)
	assert.True(t, created)
}

func TestCosmeticsService_PurchaseCosmeticForUserInternal_Success(t *testing.T) {
	called := false
	repo := &fakeCosmeticsRepo{
		addPurchaseForUserIdFn: func(cosmeticId uuid.UUID, userId uuid.UUID) error {
			called = true
			return nil
		},
	}
	bucket := &fakeCosmeticsBucketRepo{}
	service := cosmeticservice.NewCosmeticsService(nil, repo, bucket)

	err := service.PurchaseCosmeticForUserInternal(uuid.New(), uuid.New())
	assert.NoError(t, err)
	assert.True(t, called)
}
