package exports_test

import (
	"os"
	"testing"
	"time"

	"github.com/FeedTheRealm-org/core-service/config"
	"github.com/FeedTheRealm-org/core-service/internal/exports-service/models"
	"github.com/FeedTheRealm-org/core-service/internal/exports-service/repositories/exports"
	"github.com/FeedTheRealm-org/core-service/internal/utils/logger"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	logger.InitLogger(false)
	code := m.Run()
	os.Exit(code)
}

func repoTestAppName(prefix string) string {
	return prefix + "-" + uuid.NewString() + "-repo.local"
}

func cleanupRepoExports(db *config.DB) {
	_ = db.Conn.Exec("DELETE FROM export_zips WHERE app_name LIKE '%-repo.local';")
}

func TestExportRepository_CreateExportVersion(t *testing.T) {
	conf := config.CreateConfig()
	db, _ := config.NewDB(conf)
	cleanupRepoExports(db)
	repo := exports.NewExportRepository(conf, db)

	appName := repoTestAppName("realm-app")
	exportZip := &models.ExportZip{
		AppName: appName,
		Version: "v1.0.0",
		OS:      "linux",
	}

	err := repo.CreateExportVersion(exportZip)
	assert.Nil(t, err, "failed to create export version")

	result, err := repo.GetExportVersion(appName, "v1.0.0", "linux")
	assert.Nil(t, err, "failed to get export version")
	assert.NotNil(t, result, "expected export zip, got nil")
	assert.Equal(t, appName, result.AppName)
}

func TestExportRepository_GetExportVersion_NotFound(t *testing.T) {
	conf := config.CreateConfig()
	db, _ := config.NewDB(conf)
	cleanupRepoExports(db)
	repo := exports.NewExportRepository(conf, db)

	appName := repoTestAppName("notfound")
	result, err := repo.GetExportVersion(appName, "v1.0.0", "windows")
	assert.NotNil(t, err, "expected error on getting non-existing export version")
	assert.Nil(t, result, "expected no export zip to be found")
}

func TestExportRepository_GetLatestExportVersion(t *testing.T) {
	conf := config.CreateConfig()
	db, _ := config.NewDB(conf)
	cleanupRepoExports(db)
	repo := exports.NewExportRepository(conf, db)

	appName := repoTestAppName("latest-test")

	v1 := &models.ExportZip{AppName: appName, Version: "1.0.0", OS: "mac", IsLatest: false, CreatedAt: time.Now().Add(-time.Hour)}
	v2 := &models.ExportZip{AppName: appName, Version: "2.0.0", OS: "mac", IsLatest: true, CreatedAt: time.Now()}

	assert.NoError(t, repo.CreateExportVersion(v1))
	assert.NoError(t, repo.CreateExportVersion(v2))

	latest, err := repo.GetLatestExportVersion(appName, "mac")
	assert.Nil(t, err, "failed to get latest export version")
	assert.NotNil(t, latest)
	assert.Equal(t, "2.0.0", latest.Version)
}

func TestExportRepository_ListExportVersions(t *testing.T) {
	conf := config.CreateConfig()
	db, _ := config.NewDB(conf)
	cleanupRepoExports(db)
	repo := exports.NewExportRepository(conf, db)

	appNameA := repoTestAppName("list-a")
	appNameB := repoTestAppName("list-b")

	assert.NoError(t, repo.CreateExportVersion(&models.ExportZip{AppName: appNameA, Version: "1.0.0", OS: "linux"}))
	assert.NoError(t, repo.CreateExportVersion(&models.ExportZip{AppName: appNameA, Version: "2.0.0", OS: "windows"}))
	assert.NoError(t, repo.CreateExportVersion(&models.ExportZip{AppName: appNameB, Version: "1.0.0", OS: "linux"}))

	// Filtrar por App y OS
	list, err := repo.ListExportVersions(appNameA, "linux")
	assert.NoError(t, err)
	assert.Len(t, list, 1)
	assert.Equal(t, "1.0.0", list[0].Version)

	// Filtrar solo por App
	listAllA, err := repo.ListExportVersions(appNameA, "")
	assert.NoError(t, err)
	assert.Len(t, listAllA, 2)
}

func TestExportRepository_DeleteExportVersion_WithTransactionPromotion(t *testing.T) {
	conf := config.CreateConfig()
	db, _ := config.NewDB(conf)
	cleanupRepoExports(db)
	repo := exports.NewExportRepository(conf, db)

	appName := repoTestAppName("tx-delete")
	now := time.Now().UTC()

	v1 := &models.ExportZip{AppName: appName, Version: "1.0.0", OS: "linux", IsLatest: false, CreatedAt: now.Add(-2 * time.Hour)}
	v2 := &models.ExportZip{AppName: appName, Version: "2.0.0", OS: "linux", IsLatest: false, CreatedAt: now.Add(-time.Hour)}
	v3 := &models.ExportZip{AppName: appName, Version: "3.0.0", OS: "linux", IsLatest: true, CreatedAt: now}

	assert.NoError(t, repo.CreateExportVersion(v1))
	assert.NoError(t, repo.CreateExportVersion(v2))
	assert.NoError(t, repo.CreateExportVersion(v3))

	// Borrar la versión que era 'IsLatest' (v3) debe gatillar la transacción y promover a v2
	err := repo.DeleteExportVersion(appName, "3.0.0", "linux")
	assert.Nil(t, err, "failed to delete export version")

	// Verificar que v3 ya no existe
	_, err = repo.GetExportVersion(appName, "3.0.0", "linux")
	assert.NotNil(t, err, "expected error looking for deleted version")

	// Verificar que v2 fue automáticamente promovida a IsLatest = true
	updatedV2, err := repo.GetExportVersion(appName, "2.0.0", "linux")
	assert.NoError(t, err)
	assert.True(t, updatedV2.IsLatest, "expected previous version to be promoted to latest")
}

func TestExportRepository_SetLatestExportVersion(t *testing.T) {
	conf := config.CreateConfig()
	db, _ := config.NewDB(conf)
	cleanupRepoExports(db)
	repo := exports.NewExportRepository(conf, db)

	appName := repoTestAppName("switch-latest")

	v1 := &models.ExportZip{AppName: appName, Version: "1.0.0", OS: "linux", IsLatest: true}
	v2 := &models.ExportZip{AppName: appName, Version: "2.0.0", OS: "linux", IsLatest: false}

	assert.NoError(t, repo.CreateExportVersion(v1))
	assert.NoError(t, repo.CreateExportVersion(v2))

	// Cambiar el puntero de última versión de v1 hacia v2
	updated, err := repo.SetLatestExportVersion(appName, "2.0.0", "linux")
	assert.Nil(t, err)
	assert.True(t, updated.IsLatest)

	// Verificar que v1 pasó a ser IsLatest = false de forma segura
	oldLatest, err := repo.GetExportVersion(appName, "1.0.0", "linux")
	assert.NoError(t, err)
	assert.False(t, oldLatest.IsLatest, "expected old version to lose latest status")
}

func newExportsRepo(t *testing.T) exports.ExportRepository {
	t.Helper()
	conf := config.CreateConfig()
	db, err := config.NewDB(conf)
	if err != nil {
		t.Fatalf("failed to connect DB: %v", err)
	}
	cleanupRepoExports(db)
	return exports.NewExportRepository(conf, db)
}

func TestExportRepository_CreateExportVersion_Duplicate(t *testing.T) {
	repo := newExportsRepo(t)
	appName := repoTestAppName("dup")

	v1 := &models.ExportZip{AppName: appName, Version: "1.0.0", OS: "linux"}
	assert.NoError(t, repo.CreateExportVersion(v1))

	// Intentar crear la misma combinación (app+version+os) debe retornar conflict
	v1dup := &models.ExportZip{AppName: appName, Version: "1.0.0", OS: "linux"}
	err := repo.CreateExportVersion(v1dup)
	assert.Error(t, err)
}

func TestExportRepository_DeleteExportVersion_NotLatest(t *testing.T) {
	repo := newExportsRepo(t)
	appName := repoTestAppName("del-notlatest")
	now := time.Now().UTC()

	v1 := &models.ExportZip{AppName: appName, Version: "1.0.0", OS: "linux", IsLatest: false, CreatedAt: now.Add(-time.Hour)}
	v2 := &models.ExportZip{AppName: appName, Version: "2.0.0", OS: "linux", IsLatest: true, CreatedAt: now}
	assert.NoError(t, repo.CreateExportVersion(v1))
	assert.NoError(t, repo.CreateExportVersion(v2))

	// Borrar v1 (no era IsLatest) → v2 debe quedar sin cambios
	err := repo.DeleteExportVersion(appName, "1.0.0", "linux")
	assert.NoError(t, err)

	_, err = repo.GetExportVersion(appName, "1.0.0", "linux")
	assert.Error(t, err)

	v2stored, err := repo.GetExportVersion(appName, "2.0.0", "linux")
	assert.NoError(t, err)
	assert.True(t, v2stored.IsLatest)
}

func TestExportRepository_DeleteExportVersion_NotFound(t *testing.T) {
	repo := newExportsRepo(t)
	appName := repoTestAppName("del-notfound")

	err := repo.DeleteExportVersion(appName, "9.9.9", "linux")
	assert.Error(t, err)
}

func TestExportRepository_DeleteExportVersion_LastVersion(t *testing.T) {
	repo := newExportsRepo(t)
	appName := repoTestAppName("del-last")

	v1 := &models.ExportZip{AppName: appName, Version: "1.0.0", OS: "linux", IsLatest: true}
	assert.NoError(t, repo.CreateExportVersion(v1))

	// Borrar la única versión (era IsLatest) → no hay reemplazo disponible
	err := repo.DeleteExportVersion(appName, "1.0.0", "linux")
	assert.NoError(t, err)

	_, err = repo.GetExportVersion(appName, "1.0.0", "linux")
	assert.Error(t, err)
}

func TestExportRepository_GetLatestExportVersion_FallbackToNewest(t *testing.T) {
	repo := newExportsRepo(t)
	appName := repoTestAppName("fallback-latest")
	now := time.Now().UTC()

	// Ninguna versión tiene IsLatest=true → debe retornar la más reciente por created_at
	v1 := &models.ExportZip{AppName: appName, Version: "1.0.0", OS: "mac", IsLatest: false, CreatedAt: now.Add(-2 * time.Hour)}
	v2 := &models.ExportZip{AppName: appName, Version: "2.0.0", OS: "mac", IsLatest: false, CreatedAt: now.Add(-time.Hour)}
	assert.NoError(t, repo.CreateExportVersion(v1))
	assert.NoError(t, repo.CreateExportVersion(v2))

	latest, err := repo.GetLatestExportVersion(appName, "mac")
	assert.NoError(t, err)
	assert.NotNil(t, latest)
	assert.Equal(t, "2.0.0", latest.Version)
}

func TestExportRepository_GetLatestExportVersion_NotFound(t *testing.T) {
	repo := newExportsRepo(t)
	appName := repoTestAppName("notfound-latest")

	latest, err := repo.GetLatestExportVersion(appName, "linux")
	assert.Error(t, err)
	assert.Nil(t, latest)
}

func TestExportRepository_SetLatestExportVersion_NotFound(t *testing.T) {
	repo := newExportsRepo(t)
	appName := repoTestAppName("set-notfound")

	_, err := repo.SetLatestExportVersion(appName, "9.9.9", "linux")
	assert.Error(t, err)
}

func TestExportRepository_ListExportVersions_NoFilter(t *testing.T) {
	repo := newExportsRepo(t)
	appName := repoTestAppName("list-nofilter")

	assert.NoError(t, repo.CreateExportVersion(&models.ExportZip{AppName: appName, Version: "1.0.0", OS: "linux"}))
	assert.NoError(t, repo.CreateExportVersion(&models.ExportZip{AppName: appName, Version: "2.0.0", OS: "mac"}))

	all, err := repo.ListExportVersions("", "")
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(all), 2)
}
