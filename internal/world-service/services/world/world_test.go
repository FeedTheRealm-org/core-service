package world

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/FeedTheRealm-org/core-service/config"
	"github.com/FeedTheRealm-org/core-service/internal/world-service/models"
	"github.com/FeedTheRealm-org/core-service/internal/world-service/services/server_registry"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/datatypes"
)

type fakeWorldRepo struct {
	storeArg              *models.WorldData
	storeErr              error
	getWorld              *models.WorldData
	getErr                error
	updateWorld           *models.WorldData
	updateWorldErr        error
	updateCreateable      *models.WorldData
	updateCreateableErr   error
	zones                 []*models.WorldZone
	zonesErr              error
	deleteErr             error
	deleteCalled          bool
	getWorldDataCallCount int
}

func (f *fakeWorldRepo) StoreWorldData(newWorldData *models.WorldData) (*models.WorldData, error) {
	f.storeArg = newWorldData
	if f.storeErr != nil {
		return nil, f.storeErr
	}
	return newWorldData, nil
}

func (f *fakeWorldRepo) GetWorldData(worldID uuid.UUID) (*models.WorldData, error) {
	f.getWorldDataCallCount++
	if f.getErr != nil {
		return nil, f.getErr
	}
	return f.getWorld, nil
}

func (f *fakeWorldRepo) UpdateWorldData(worldID uuid.UUID, userId uuid.UUID, data []byte, description string) (*models.WorldData, error) {
	if f.updateWorldErr != nil {
		return nil, f.updateWorldErr
	}
	return f.updateWorld, nil
}

func (f *fakeWorldRepo) UpdateCreateableData(worldID uuid.UUID, userId uuid.UUID, createableData []byte) (*models.WorldData, error) {
	if f.updateCreateableErr != nil {
		return nil, f.updateCreateableErr
	}
	return f.updateCreateable, nil
}

func (f *fakeWorldRepo) UpsertWorldZone(worldID uuid.UUID, zoneID int, zoneData []byte) (*models.WorldZone, error) {
	return &models.WorldZone{ID: zoneID, WorldID: worldID}, nil
}

func (f *fakeWorldRepo) SetWorldZoneActiveState(worldID uuid.UUID, zoneID int, isActive bool) error {
	return nil
}

func (f *fakeWorldRepo) SetWorldZoneOnlineState(worldID uuid.UUID, zoneID int, isOnline bool) error {
	return nil
}

func (f *fakeWorldRepo) GetWorldZoneActiveState(worldID uuid.UUID, zoneID int) (bool, error) {
	return false, nil
}

func (f *fakeWorldRepo) DeleteWorldData(worldID uuid.UUID) error {
	f.deleteCalled = true
	return f.deleteErr
}

func (f *fakeWorldRepo) GetWorldsList(offset int, limit int, filter string, userId uuid.UUID) ([]*models.WorldData, error) {
	return nil, nil
}

func (f *fakeWorldRepo) GetWorldZones(worldID uuid.UUID) ([]*models.WorldZone, error) {
	if f.zonesErr != nil {
		return nil, f.zonesErr
	}
	return f.zones, nil
}

func (f *fakeWorldRepo) GetWorldZone(worldID uuid.UUID, zoneID int) (*models.WorldZone, error) {
	return nil, nil
}

func (f *fakeWorldRepo) GetActiveWorldZones() ([]*models.WorldZone, error) {
	return nil, nil
}

func (f *fakeWorldRepo) UpdateWorldZonePlayerCount(worldID uuid.UUID, zoneID int, activePlayers int, averagePlayerTime int) error {
	return nil
}

func (f *fakeWorldRepo) GetWorldZonePlayerCounts(worldID uuid.UUID) (int, int, error) {
	return 0, 0, nil
}

func (f *fakeWorldRepo) GetAllWorldZonePlayerCounts() (int, int, error) {
	return 0, 0, nil
}

func (f *fakeWorldRepo) GetUserIdByWorldId(worldID uuid.UUID) (uuid.UUID, error) {
	return uuid.Nil, nil
}

func (f *fakeWorldRepo) GetTotalZonesCountByUserId(userId uuid.UUID) (int64, error) {
	return 0, nil
}

func (f *fakeWorldRepo) GetWorldIdsByUserId(userId uuid.UUID) ([]uuid.UUID, error) {
	return nil, nil
}

func (f *fakeWorldRepo) ClearDatabase() error {
	return nil
}

type fakeServerRegistry struct {
	stopCalls []string
	stopErr   error
}

func (f *fakeServerRegistry) StartNewJob(worldId uuid.UUID, zoneId int, isTest bool) error {
	return nil
}

func (f *fakeServerRegistry) StopJob(worldId uuid.UUID, zoneId int) error {
	if f.stopErr != nil {
		return f.stopErr
	}
	f.stopCalls = append(f.stopCalls, worldId.String()+":"+strconv.Itoa(zoneId))
	return nil
}

func (f *fakeServerRegistry) GetServerAddress(worldId uuid.UUID, zoneId int) (string, int, error) {
	return "", 0, errors.New("not implemented")
}

func TestWorldService_PublishWorld_DefaultCreateableData(t *testing.T) {
	repo := &fakeWorldRepo{}
	registry := &fakeServerRegistry{}
	conf := config.CreateConfig()
	svc := NewWorldService(conf, repo, registry)

	data := &models.WorldData{
		ID:          uuid.New(),
		UserId:      uuid.New(),
		Name:        "world-" + uuid.NewString(),
		Data:        datatypes.JSON([]byte(`{"k":1}`)),
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
		Description: "desc",
	}
	created, err := svc.PublishWorld(data)
	require.NoError(t, err)
	assert.NotNil(t, created)
	assert.True(t, len(repo.storeArg.CreateableData) > 0)
	assert.Equal(t, "{}", strings.TrimSpace(string(repo.storeArg.CreateableData)))
}

func TestWorldService_DeleteWorld_NotOwner(t *testing.T) {
	ownerID := uuid.New()
	repo := &fakeWorldRepo{
		getWorld: &models.WorldData{ID: uuid.New(), UserId: ownerID},
	}
	registry := &fakeServerRegistry{}
	conf := config.CreateConfig()
	svc := NewWorldService(conf, repo, registry)

	err := svc.DeleteWorld(repo.getWorld.ID, uuid.New())
	assert.Error(t, err)
	assert.False(t, repo.deleteCalled)
}

func TestWorldService_DeleteWorld_StopsActiveZones(t *testing.T) {
	ownerID := uuid.New()
	worldID := uuid.New()
	repo := &fakeWorldRepo{
		getWorld: &models.WorldData{ID: worldID, UserId: ownerID},
		zones:    []*models.WorldZone{{ID: 1, WorldID: worldID, IsActive: true}, {ID: 2, WorldID: worldID, IsActive: false}},
	}
	registry := &fakeServerRegistry{}
	conf := config.CreateConfig()
	conf.Server.SubscriptionOn = false
	svc := NewWorldService(conf, repo, registry)

	err := svc.DeleteWorld(worldID, ownerID)
	assert.NoError(t, err)
	assert.True(t, repo.deleteCalled)
	assert.Len(t, registry.stopCalls, 1)
}

func TestWorldService_UpdateUsedSlots_HTTP(t *testing.T) {
	userID := uuid.New()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		if !strings.Contains(r.URL.Path, "/subscriptions/internal/users/") {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	conf := config.CreateConfig()
	portStr := strings.TrimPrefix(server.URL, "http://127.0.0.1:")
	port, _ := strconv.Atoi(portStr)
	conf.Server.Port = port

	svc := NewWorldService(conf, &fakeWorldRepo{}, server_registry.NewStubServerRegistryService()).(*worldService)
	assert.NoError(t, svc.UpdateUsedSlots(userID, 2, true))
}

func TestWorldService_UpdateUsedSlots_HTTPError(t *testing.T) {
	userID := uuid.New()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	conf := config.CreateConfig()
	portStr := strings.TrimPrefix(server.URL, "http://127.0.0.1:")
	port, _ := strconv.Atoi(portStr)
	conf.Server.Port = port

	svc := NewWorldService(conf, &fakeWorldRepo{}, server_registry.NewStubServerRegistryService()).(*worldService)
	err := svc.UpdateUsedSlots(userID, 1, false)
	assert.Error(t, err)
}
