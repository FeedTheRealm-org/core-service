package exports_test

import (
	"errors"
	"fmt"
	"mime/multipart"
	"os"
	"testing"

	"github.com/FeedTheRealm-org/core-service/config"
	"github.com/FeedTheRealm-org/core-service/internal/exports-service/models"
	"github.com/FeedTheRealm-org/core-service/internal/exports-service/services/exports"
	"github.com/FeedTheRealm-org/core-service/internal/utils/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockExportRepository struct {
	mock.Mock
}

func (m *MockExportRepository) CreateExportVersion(exportZip *models.ExportZip) error {
	args := m.Called(exportZip)
	return args.Error(0)
}

func (m *MockExportRepository) GetExportVersion(appName, version, osName string) (*models.ExportZip, error) {
	args := m.Called(appName, version, osName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.ExportZip), args.Error(1)
}

func (m *MockExportRepository) GetLatestExportVersion(appName, osName string) (*models.ExportZip, error) {
	args := m.Called(appName, osName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.ExportZip), args.Error(1)
}

func (m *MockExportRepository) ListExportVersions(appName, osName string) ([]*models.ExportZip, error) {
	args := m.Called(appName, osName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.ExportZip), args.Error(1)
}

func (m *MockExportRepository) DeleteExportVersion(appName, version, osName string) error {
	args := m.Called(appName, version, osName)
	return args.Error(0)
}

func (m *MockExportRepository) SetLatestExportVersion(appName, version, osName string) (*models.ExportZip, error) {
	args := m.Called(appName, version, osName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.ExportZip), args.Error(1)
}

type MockBucketRepository struct {
	mock.Mock
}

func (m *MockBucketRepository) UploadFile(filePath, contentType string, file multipart.File) error {
	args := m.Called(filePath, contentType, file)
	return args.Error(0)
}

func (m *MockBucketRepository) DeleteFile(filePath string) error {
	args := m.Called(filePath)
	return args.Error(0)
}

func (m *MockBucketRepository) GetBaseUrl() string {
	args := m.Called()
	return args.String(0)
}

func TestMain(m *testing.M) {
	logger.InitLogger(false)
	code := m.Run()
	os.Exit(code)
}

func TestExportsService_UploadZip_Success(t *testing.T) {
	mockRepo := new(MockExportRepository)
	mockBucket := new(MockBucketRepository)
	conf := &config.Config{}
	service := exports.NewExportsService(conf, mockRepo, mockBucket)

	appName, version, osName := "realm-striker", "v1.2.0", "android"
	expectedPath := fmt.Sprintf("exports/%s/%s/%s-%s.zip", appName, osName, appName, version)

	// Configurar expectativas de los Mocks
	mockBucket.On("UploadFile", expectedPath, "application/zip", mock.Anything).Return(nil)
	mockRepo.On("CreateExportVersion", mock.AnythingOfType("*models.ExportZip")).Return(nil)

	res, err := service.UploadZip(appName, version, osName, "Fix bugs", nil)

	assert.Nil(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, fmt.Sprintf("/%s", expectedPath), res.Path)
	mockBucket.AssertExpectations(t)
	mockRepo.AssertExpectations(t)
}

func TestExportsService_UploadZip_BucketError(t *testing.T) {
	mockRepo := new(MockExportRepository)
	mockBucket := new(MockBucketRepository)
	service := exports.NewExportsService(&config.Config{}, mockRepo, mockBucket)

	mockBucket.On("UploadFile", mock.Anything, mock.Anything, mock.Anything).
		Return(errors.New("s3 connection timeout"))

	res, err := service.UploadZip("app", "v1", "linux", "note", nil)

	assert.NotNil(t, err)
	assert.Nil(t, res)
	assert.Equal(t, "s3 connection timeout", err.Error())
	mockRepo.AssertNotCalled(t, "CreateExportVersion", mock.Anything) // DB no debería llamarse si falló el bucket
}

func TestExportsService_GetZipPath(t *testing.T) {
	t.Run("Should call GetExportVersion when version is provided", func(t *testing.T) {
		mockRepo := new(MockExportRepository)
		mockBucket := new(MockBucketRepository)
		service := exports.NewExportsService(&config.Config{}, mockRepo, mockBucket)

		expectedZip := &models.ExportZip{Path: "/exports/app/ios/app-v1.zip"}
		mockRepo.On("GetExportVersion", "app", "v1", "ios").Return(expectedZip, nil)

		path, err := service.GetZipPath("app", "v1", "ios")

		assert.Nil(t, err)
		assert.Equal(t, "/exports/app/ios/app-v1.zip", path)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Should call GetLatestExportVersion when version is empty", func(t *testing.T) {
		mockRepo := new(MockExportRepository)
		mockBucket := new(MockBucketRepository)
		service := exports.NewExportsService(&config.Config{}, mockRepo, mockBucket)

		expectedZip := &models.ExportZip{Path: "/exports/app/ios/app-latest.zip"}
		mockRepo.On("GetLatestExportVersion", "app", "ios").Return(expectedZip, nil)

		path, err := service.GetZipPath("app", "", "ios")

		assert.Nil(t, err)
		assert.Equal(t, "/exports/app/ios/app-latest.zip", path)
		mockRepo.AssertExpectations(t)
	})
}

func TestExportsService_DeleteZipVersion_Success(t *testing.T) {
	mockRepo := new(MockExportRepository)
	mockBucket := new(MockBucketRepository)
	service := exports.NewExportsService(&config.Config{}, mockRepo, mockBucket)

	appName, version, osName := "game-core", "v2.0", "windows"
	expectedPath := "exports/game-core/windows/game-core-v2.0.zip"

	mockRepo.On("DeleteExportVersion", appName, version, osName).Return(nil)
	mockBucket.On("DeleteFile", expectedPath).Return(nil)

	err := service.DeleteZipVersion(appName, version, osName)

	assert.Nil(t, err)
	mockRepo.AssertExpectations(t)
	mockBucket.AssertExpectations(t)
}

func TestExportsService_DeleteZipVersion_RepoError(t *testing.T) {
	mockRepo := new(MockExportRepository)
	mockBucket := new(MockBucketRepository)
	service := exports.NewExportsService(&config.Config{}, mockRepo, mockBucket)

	mockRepo.On("DeleteExportVersion", mock.Anything, mock.Anything, mock.Anything).
		Return(errors.New("row locked"))

	err := service.DeleteZipVersion("game-core", "v2.0", "windows")

	assert.NotNil(t, err)
	assert.Equal(t, "row locked", err.Error())
	mockBucket.AssertNotCalled(t, "DeleteFile", mock.Anything) // Si falla la DB, no borramos el archivo físico
}

func TestExportsService_ListZipVersions(t *testing.T) {
	mockRepo := new(MockExportRepository)
	service := exports.NewExportsService(&config.Config{}, mockRepo, nil)

	expectedList := []*models.ExportZip{
		{AppName: "app", Version: "1.0"},
		{AppName: "app", Version: "2.0"},
	}
	mockRepo.On("ListExportVersions", "app", "linux").Return(expectedList, nil)

	list, err := service.ListZipVersions("app", "linux")

	assert.Nil(t, err)
	assert.Len(t, list, 2)
	assert.Equal(t, "1.0", list[0].Version)
}
