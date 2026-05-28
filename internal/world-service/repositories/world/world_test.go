package world

import (
	"errors"
	"os"
	"testing"

	"github.com/FeedTheRealm-org/core-service/config"
	"github.com/FeedTheRealm-org/core-service/internal/utils/logger"
	world_errors "github.com/FeedTheRealm-org/core-service/internal/world-service/errors"
	"github.com/FeedTheRealm-org/core-service/internal/world-service/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

var worldConf *config.Config
var worldDB *config.DB
var worldRepo WorldRepository

func TestMain(m *testing.M) {
	logger.InitLogger(false)
	worldConf = config.CreateConfig()
	var err error
	worldDB, err = config.NewDB(worldConf)
	if err != nil {
		panic(err)
	}
	worldRepo = NewWorldRepository(worldConf, worldDB)

	code := m.Run()
	os.Exit(code)
}

func newWorldData(userID uuid.UUID, name string) *models.WorldData {
	return &models.WorldData{
		ID:             uuid.New(),
		UserId:         userID,
		Name:           name,
		Description:    "desc",
		Data:           datatypes.JSON([]byte(`{"k":1}`)),
		CreateableData: datatypes.JSON([]byte(`{"c":2}`)),
	}
}

func cleanupWorld(t *testing.T, worldID uuid.UUID) {
	_ = worldDB.Conn.Exec("DELETE FROM world_zones WHERE world_id = ?", worldID).Error
	_ = worldDB.Conn.Exec("DELETE FROM world_data WHERE id = ?", worldID).Error
}

func createWorld(t *testing.T, userID uuid.UUID) *models.WorldData {
	worldName := "world-" + uuid.NewString()
	created, err := worldRepo.StoreWorldData(newWorldData(userID, worldName))
	require.NoError(t, err)
	t.Cleanup(func() {
		cleanupWorld(t, created.ID)
	})
	return created
}

func TestWorldRepository_StoreAndGet(t *testing.T) {
	userID := uuid.New()
	created := createWorld(t, userID)

	fetched, err := worldRepo.GetWorldData(created.ID)
	assert.NoError(t, err)
	assert.Equal(t, created.ID, fetched.ID)
	assert.Equal(t, created.Name, fetched.Name)
}

func TestWorldRepository_GetWorldData_NotFound(t *testing.T) {
	_, err := worldRepo.GetWorldData(uuid.New())
	assert.Error(t, err)
	var notFound *world_errors.WorldInfoNotFound
	assert.True(t, errors.As(err, &notFound))
}

func TestWorldRepository_UpdateWorldData_Unauthorized(t *testing.T) {
	ownerID := uuid.New()
	created := createWorld(t, ownerID)

	_, err := worldRepo.UpdateWorldData(created.ID, uuid.New(), []byte(`{"a":1}`), "new")
	assert.Error(t, err)
}

func TestWorldRepository_UpdateWorldData_Success(t *testing.T) {
	userID := uuid.New()
	created := createWorld(t, userID)

	updated, err := worldRepo.UpdateWorldData(created.ID, userID, []byte(`{"a":2}`), "updated")
	assert.NoError(t, err)
	assert.Contains(t, string(updated.Data), "\"a\"")
	assert.Equal(t, "updated", updated.Description)
}

func TestWorldRepository_UpdateCreateableData_Success(t *testing.T) {
	userID := uuid.New()
	created := createWorld(t, userID)

	updated, err := worldRepo.UpdateCreateableData(created.ID, userID, []byte(`{"x":3}`))
	assert.NoError(t, err)
	assert.Contains(t, string(updated.CreateableData), "\"x\"")
}

func TestWorldRepository_UpdateCreateableData_Unauthorized(t *testing.T) {
	userID := uuid.New()
	created := createWorld(t, userID)

	_, err := worldRepo.UpdateCreateableData(created.ID, uuid.New(), []byte(`{"x":9}`))
	assert.Error(t, err)
}

func TestWorldRepository_UpsertWorldZone_CreateAndUpdate(t *testing.T) {
	userID := uuid.New()
	created := createWorld(t, userID)

	zone, err := worldRepo.UpsertWorldZone(created.ID, 1, []byte(`{"z":1}`))
	assert.NoError(t, err)
	assert.Equal(t, 1, zone.ID)
	assert.Contains(t, string(zone.ZoneData), "\"z\"")

	zone, err = worldRepo.UpsertWorldZone(created.ID, 1, []byte(`{"z":2}`))
	assert.NoError(t, err)
	assert.Contains(t, string(zone.ZoneData), "\"z\"")
}

func TestWorldRepository_ZoneState(t *testing.T) {
	userID := uuid.New()
	created := createWorld(t, userID)

	_, err := worldRepo.UpsertWorldZone(created.ID, 2, []byte(`{"z":1}`))
	require.NoError(t, err)

	assert.NoError(t, worldRepo.SetWorldZoneActiveState(created.ID, 2, true))
	isActive, err := worldRepo.GetWorldZoneActiveState(created.ID, 2)
	assert.NoError(t, err)
	assert.True(t, isActive)

	assert.NoError(t, worldRepo.SetWorldZoneOnlineState(created.ID, 2, true))
}

func TestWorldRepository_SetWorldZoneActiveState_NotFound(t *testing.T) {
	err := worldRepo.SetWorldZoneActiveState(uuid.New(), 1, true)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, gorm.ErrRecordNotFound))
}

func TestWorldRepository_GetWorldsList_FilterAndUser(t *testing.T) {
	userID := uuid.New()
	created := createWorld(t, userID)
	_ = createWorld(t, uuid.New())

	list, err := worldRepo.GetWorldsList(0, 10, created.Name[:5], userID)
	assert.NoError(t, err)
	assert.Len(t, list, 1)
	assert.Equal(t, created.ID, list[0].ID)
}

func TestWorldRepository_GetWorldZonesAndGetWorldZone(t *testing.T) {
	userID := uuid.New()
	created := createWorld(t, userID)

	_, err := worldRepo.UpsertWorldZone(created.ID, 1, []byte(`{"z":1}`))
	require.NoError(t, err)
	_, err = worldRepo.UpsertWorldZone(created.ID, 2, []byte(`{"z":2}`))
	require.NoError(t, err)

	zones, err := worldRepo.GetWorldZones(created.ID)
	assert.NoError(t, err)
	assert.Len(t, zones, 2)

	zone, err := worldRepo.GetWorldZone(created.ID, 2)
	assert.NoError(t, err)
	assert.Equal(t, 2, zone.ID)
}

func TestWorldRepository_ActiveZonesAndCounts(t *testing.T) {
	userID := uuid.New()
	created := createWorld(t, userID)

	_, err := worldRepo.UpsertWorldZone(created.ID, 1, []byte(`{"z":1}`))
	require.NoError(t, err)
	_, err = worldRepo.UpsertWorldZone(created.ID, 2, []byte(`{"z":2}`))
	require.NoError(t, err)

	assert.NoError(t, worldRepo.SetWorldZoneActiveState(created.ID, 1, true))
	assert.NoError(t, worldRepo.SetWorldZoneOnlineState(created.ID, 1, true))
	assert.NoError(t, worldRepo.UpdateWorldZonePlayerCount(created.ID, 1, 5, 10))
	assert.NoError(t, worldRepo.SetWorldZoneOnlineState(created.ID, 2, true))
	assert.NoError(t, worldRepo.UpdateWorldZonePlayerCount(created.ID, 2, 3, 20))

	active, err := worldRepo.GetActiveWorldZones()
	assert.NoError(t, err)
	assert.NotEmpty(t, active)

	totalPlayers, avgTime, err := worldRepo.GetWorldZonePlayerCounts(created.ID)
	assert.NoError(t, err)
	assert.Equal(t, 8, totalPlayers)
	assert.Equal(t, 15, avgTime)

	totalAll, avgAll, err := worldRepo.GetAllWorldZonePlayerCounts()
	assert.NoError(t, err)
	assert.True(t, totalAll >= 8)
	assert.True(t, avgAll >= 10)
}

func TestWorldRepository_UserAndZonesSummary(t *testing.T) {
	userID := uuid.New()
	created := createWorld(t, userID)
	_, err := worldRepo.UpsertWorldZone(created.ID, 1, []byte(`{"z":1}`))
	require.NoError(t, err)

	ownerID, err := worldRepo.GetUserIdByWorldId(created.ID)
	assert.NoError(t, err)
	assert.Equal(t, userID, ownerID)

	count, err := worldRepo.GetTotalZonesCountByUserId(userID)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), count)

	worldIDs, err := worldRepo.GetWorldIdsByUserId(userID)
	assert.NoError(t, err)
	assert.Contains(t, worldIDs, created.ID)
}

func TestWorldRepository_DeleteWorldData_NotFound(t *testing.T) {
	err := worldRepo.DeleteWorldData(uuid.New())
	assert.Error(t, err)
	var notFound *world_errors.WorldInfoNotFound
	assert.True(t, errors.As(err, &notFound))
}

func TestWorldRepository_DeleteWorldData_Success(t *testing.T) {
	userID := uuid.New()
	created := createWorld(t, userID)

	err := worldRepo.DeleteWorldData(created.ID)
	assert.NoError(t, err)
}
