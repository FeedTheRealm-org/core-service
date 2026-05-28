package zones

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/FeedTheRealm-org/core-service/config"
	"github.com/FeedTheRealm-org/core-service/internal/world-service/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/datatypes"
)

type fakeZonesRepo struct {
	worlds         map[uuid.UUID]*models.WorldData
	zones          map[uuid.UUID][]*models.WorldZone
	active         map[string]bool
	online         map[string]bool
	userByWorld    map[uuid.UUID]uuid.UUID
	setActiveErr   error
	setOnlineErr   error
	updateCountErr error
	playerCounts   map[uuid.UUID][2]int
	allCounts      [2]int
}

func newFakeZonesRepo() *fakeZonesRepo {
	return &fakeZonesRepo{
		worlds:       make(map[uuid.UUID]*models.WorldData),
		zones:        make(map[uuid.UUID][]*models.WorldZone),
		active:       make(map[string]bool),
		online:       make(map[string]bool),
		userByWorld:  make(map[uuid.UUID]uuid.UUID),
		playerCounts: make(map[uuid.UUID][2]int),
	}
}

func zoneKey(worldID uuid.UUID, zoneID int) string {
	return worldID.String() + ":" + strconv.Itoa(zoneID)
}

func (f *fakeZonesRepo) StoreWorldData(newWorldData *models.WorldData) (*models.WorldData, error) {
	f.worlds[newWorldData.ID] = newWorldData
	return newWorldData, nil
}

func (f *fakeZonesRepo) GetWorldData(worldID uuid.UUID) (*models.WorldData, error) {
	world, ok := f.worlds[worldID]
	if !ok {
		return nil, errors.New("not found")
	}
	return world, nil
}

func (f *fakeZonesRepo) UpdateWorldData(worldID uuid.UUID, userId uuid.UUID, data []byte, description string) (*models.WorldData, error) {
	return nil, nil
}

func (f *fakeZonesRepo) UpdateCreateableData(worldID uuid.UUID, userId uuid.UUID, createableData []byte) (*models.WorldData, error) {
	return nil, nil
}

func (f *fakeZonesRepo) UpsertWorldZone(worldID uuid.UUID, zoneID int, zoneData []byte) (*models.WorldZone, error) {
	zone := &models.WorldZone{ID: zoneID, WorldID: worldID, ZoneData: datatypes.JSON(zoneData)}
	f.zones[worldID] = append(f.zones[worldID], zone)
	return zone, nil
}

func (f *fakeZonesRepo) SetWorldZoneActiveState(worldID uuid.UUID, zoneID int, isActive bool) error {
	if f.setActiveErr != nil {
		return f.setActiveErr
	}
	f.active[zoneKey(worldID, zoneID)] = isActive
	zones := f.zones[worldID]
	for _, zone := range zones {
		if zone.ID == zoneID {
			zone.IsActive = isActive
		}
	}
	return nil
}

func (f *fakeZonesRepo) SetWorldZoneOnlineState(worldID uuid.UUID, zoneID int, isOnline bool) error {
	if f.setOnlineErr != nil {
		return f.setOnlineErr
	}
	f.online[zoneKey(worldID, zoneID)] = isOnline
	zones := f.zones[worldID]
	for _, zone := range zones {
		if zone.ID == zoneID {
			zone.IsOnline = isOnline
		}
	}
	return nil
}

func (f *fakeZonesRepo) GetWorldZoneActiveState(worldID uuid.UUID, zoneID int) (bool, error) {
	return f.active[zoneKey(worldID, zoneID)], nil
}

func (f *fakeZonesRepo) DeleteWorldData(worldID uuid.UUID) error {
	return nil
}

func (f *fakeZonesRepo) GetWorldsList(offset int, limit int, filter string, userId uuid.UUID) ([]*models.WorldData, error) {
	return nil, nil
}

func (f *fakeZonesRepo) GetWorldZones(worldID uuid.UUID) ([]*models.WorldZone, error) {
	return f.zones[worldID], nil
}

func (f *fakeZonesRepo) GetWorldZone(worldID uuid.UUID, zoneID int) (*models.WorldZone, error) {
	for _, zone := range f.zones[worldID] {
		if zone.ID == zoneID {
			return zone, nil
		}
	}
	return nil, errors.New("not found")
}

func (f *fakeZonesRepo) GetActiveWorldZones() ([]*models.WorldZone, error) {
	return nil, nil
}

func (f *fakeZonesRepo) UpdateWorldZonePlayerCount(worldID uuid.UUID, zoneID int, activePlayers int, averagePlayerTime int) error {
	if f.updateCountErr != nil {
		return f.updateCountErr
	}
	zones := f.zones[worldID]
	for _, zone := range zones {
		if zone.ID == zoneID {
			zone.ActivePlayers = activePlayers
			zone.AveragePlayerTime = averagePlayerTime
		}
	}
	return nil
}

func (f *fakeZonesRepo) GetWorldZonePlayerCounts(worldID uuid.UUID) (int, int, error) {
	if counts, ok := f.playerCounts[worldID]; ok {
		return counts[0], counts[1], nil
	}
	total := 0
	avgSum := 0
	onlineCount := 0
	for _, zone := range f.zones[worldID] {
		if zone.IsOnline {
			total += zone.ActivePlayers
			avgSum += zone.AveragePlayerTime
			onlineCount++
		}
	}
	avg := 0
	if onlineCount > 0 {
		avg = avgSum / onlineCount
	}
	return total, avg, nil
}

func (f *fakeZonesRepo) GetAllWorldZonePlayerCounts() (int, int, error) {
	if f.allCounts[0] != 0 || f.allCounts[1] != 0 {
		return f.allCounts[0], f.allCounts[1], nil
	}
	total := 0
	avgSum := 0
	onlineCount := 0
	for _, zones := range f.zones {
		for _, zone := range zones {
			if zone.IsOnline {
				total += zone.ActivePlayers
				avgSum += zone.AveragePlayerTime
				onlineCount++
			}
		}
	}
	avg := 0
	if onlineCount > 0 {
		avg = avgSum / onlineCount
	}
	return total, avg, nil
}

func (f *fakeZonesRepo) GetUserIdByWorldId(worldID uuid.UUID) (uuid.UUID, error) {
	userID, ok := f.userByWorld[worldID]
	if !ok {
		return uuid.Nil, errors.New("not found")
	}
	return userID, nil
}

func (f *fakeZonesRepo) GetTotalZonesCountByUserId(userId uuid.UUID) (int64, error) {
	return 0, nil
}

func (f *fakeZonesRepo) GetWorldIdsByUserId(userId uuid.UUID) ([]uuid.UUID, error) {
	worldIDs := []uuid.UUID{}
	for worldID, ownerID := range f.userByWorld {
		if ownerID == userId {
			worldIDs = append(worldIDs, worldID)
		}
	}
	return worldIDs, nil
}

func (f *fakeZonesRepo) ClearDatabase() error {
	return nil
}

type fakeZonesRegistry struct {
	startCalls []string
	stopCalls  []string
	startErr   error
	stopErr    error
}

func (f *fakeZonesRegistry) StartNewJob(worldId uuid.UUID, zoneId int, isTest bool) error {
	if f.startErr != nil {
		return f.startErr
	}
	f.startCalls = append(f.startCalls, worldId.String()+":"+strconv.Itoa(zoneId))
	return nil
}

func (f *fakeZonesRegistry) StopJob(worldId uuid.UUID, zoneId int) error {
	if f.stopErr != nil {
		return f.stopErr
	}
	f.stopCalls = append(f.stopCalls, worldId.String()+":"+strconv.Itoa(zoneId))
	return nil
}

func (f *fakeZonesRegistry) GetServerAddress(worldId uuid.UUID, zoneId int) (string, int, error) {
	return "", 0, nil
}

func TestZonesService_ActivateZone_AlreadyActive(t *testing.T) {
	repo := newFakeZonesRepo()
	worldID := uuid.New()
	repo.active[zoneKey(worldID, 1)] = true

	registry := &fakeZonesRegistry{}
	conf := config.CreateConfig()
	conf.Server.SubscriptionOn = false
	svc := NewZonesService(conf, repo, registry)

	err := svc.ActivateZone(worldID, 1)
	assert.NoError(t, err)
	assert.Empty(t, registry.startCalls)
}

func TestZonesService_ActivateZone_StartsAndSetsActive(t *testing.T) {
	repo := newFakeZonesRepo()
	worldID := uuid.New()
	repo.active[zoneKey(worldID, 1)] = false

	registry := &fakeZonesRegistry{}
	conf := config.CreateConfig()
	conf.Server.SubscriptionOn = false
	svc := NewZonesService(conf, repo, registry)

	err := svc.ActivateZone(worldID, 1)
	assert.NoError(t, err)
	assert.Len(t, registry.startCalls, 1)
	assert.True(t, repo.active[zoneKey(worldID, 1)])
}

func TestZonesService_ActivateZone_SetActiveFails(t *testing.T) {
	repo := newFakeZonesRepo()
	worldID := uuid.New()
	repo.setActiveErr = errors.New("fail")

	registry := &fakeZonesRegistry{}
	conf := config.CreateConfig()
	conf.Server.SubscriptionOn = false
	svc := NewZonesService(conf, repo, registry)

	err := svc.ActivateZone(worldID, 1)
	assert.Error(t, err)
	assert.Len(t, registry.stopCalls, 1)
}

func TestZonesService_DeactivateZone_HappyPath(t *testing.T) {
	repo := newFakeZonesRepo()
	worldID := uuid.New()
	repo.active[zoneKey(worldID, 1)] = true

	registry := &fakeZonesRegistry{}
	conf := config.CreateConfig()
	conf.Server.SubscriptionOn = false
	svc := NewZonesService(conf, repo, registry)

	err := svc.DeactivateZone(worldID, 1)
	assert.NoError(t, err)
	assert.Len(t, registry.stopCalls, 1)
	assert.False(t, repo.active[zoneKey(worldID, 1)])
}

func TestZonesService_DeactivateZone_NotActive(t *testing.T) {
	repo := newFakeZonesRepo()
	worldID := uuid.New()

	registry := &fakeZonesRegistry{}
	conf := config.CreateConfig()
	conf.Server.SubscriptionOn = false
	svc := NewZonesService(conf, repo, registry)

	err := svc.DeactivateZone(worldID, 1)
	assert.NoError(t, err)
	assert.Empty(t, registry.stopCalls)
}

func TestZonesService_StopAllZonesForUser(t *testing.T) {
	repo := newFakeZonesRepo()
	userID := uuid.New()
	worldID := uuid.New()
	repo.userByWorld[worldID] = userID
	repo.zones[worldID] = []*models.WorldZone{{ID: 1, WorldID: worldID, IsActive: true}}
	repo.active[zoneKey(worldID, 1)] = true

	registry := &fakeZonesRegistry{}
	conf := config.CreateConfig()
	conf.Server.SubscriptionOn = false
	svc := NewZonesService(conf, repo, registry)

	err := svc.StopAllZonesForUser(userID)
	assert.NoError(t, err)
	assert.Len(t, registry.stopCalls, 1)
}

func TestZonesService_PlayerCounts(t *testing.T) {
	repo := newFakeZonesRepo()
	worldID := uuid.New()
	repo.zones[worldID] = []*models.WorldZone{{ID: 1, WorldID: worldID, IsOnline: true, ActivePlayers: 3, AveragePlayerTime: 10}}

	registry := &fakeZonesRegistry{}
	conf := config.CreateConfig()
	svc := NewZonesService(conf, repo, registry)

	total, avg, err := svc.GetWorldZonePlayerCounts(worldID)
	assert.NoError(t, err)
	assert.Equal(t, 3, total)
	assert.Equal(t, 10, avg)

	totalAll, avgAll, err := svc.GetAllWorldZonePlayerCounts()
	assert.NoError(t, err)
	assert.Equal(t, 3, totalAll)
	assert.Equal(t, 10, avgAll)
}

func TestZonesService_CheckAvailableZones(t *testing.T) {
	userID := uuid.New()
	worldID := uuid.New()
	repo := newFakeZonesRepo()
	repo.userByWorld[worldID] = userID

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/status") {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"data":{"allowed":true,"free_slots":1}}`))
			return
		}
		if r.Method == http.MethodPut {
			w.WriteHeader(http.StatusOK)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	conf := config.CreateConfig()
	portStr := strings.TrimPrefix(server.URL, "http://127.0.0.1:")
	port, _ := strconv.Atoi(portStr)
	conf.Server.Port = port
	conf.Server.SubscriptionOn = true

	svc := NewZonesService(conf, repo, &fakeZonesRegistry{}).(*zonesService)
	assert.NoError(t, svc.checkAvailableZonesForActivation(worldID))
	assert.NoError(t, svc.updateUsedSlots(userID, 1, true))
}

func TestZonesService_CheckAvailableZones_Denied(t *testing.T) {
	userID := uuid.New()
	worldID := uuid.New()
	repo := newFakeZonesRepo()
	repo.userByWorld[worldID] = userID

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"allowed":false,"free_slots":0}}`))
	}))
	defer server.Close()

	conf := config.CreateConfig()
	portStr := strings.TrimPrefix(server.URL, "http://127.0.0.1:")
	port, _ := strconv.Atoi(portStr)
	conf.Server.Port = port
	conf.Server.SubscriptionOn = true

	svc := NewZonesService(conf, repo, &fakeZonesRegistry{}).(*zonesService)
	err := svc.checkAvailableZonesForActivation(worldID)
	assert.Error(t, err)
}

func TestZonesService_UpdateZoneStatusAndPlayerCount(t *testing.T) {
	repo := newFakeZonesRepo()
	worldID := uuid.New()
	repo.zones[worldID] = []*models.WorldZone{{ID: 1, WorldID: worldID}}

	registry := &fakeZonesRegistry{}
	conf := config.CreateConfig()
	svc := NewZonesService(conf, repo, registry)

	assert.NoError(t, svc.UpdateZoneStatus(worldID, 1, true))
	assert.True(t, repo.online[zoneKey(worldID, 1)])

	assert.NoError(t, svc.UpdateZonePlayerCount(worldID, 1, 7, 12))
	zone, err := repo.GetWorldZone(worldID, 1)
	require.NoError(t, err)
	assert.Equal(t, 7, zone.ActivePlayers)
	assert.Equal(t, 12, zone.AveragePlayerTime)
}

func TestZonesService_GetWorldAndPublishZone(t *testing.T) {
	repo := newFakeZonesRepo()
	worldID := uuid.New()
	repo.worlds[worldID] = &models.WorldData{ID: worldID, UserId: uuid.New(), Name: "world"}

	registry := &fakeZonesRegistry{}
	conf := config.CreateConfig()
	svc := NewZonesService(conf, repo, registry)

	world, err := svc.GetWorld(worldID)
	assert.NoError(t, err)
	assert.Equal(t, worldID, world.ID)

	zone, err := svc.PublishZone(worldID, 1, []byte(`{"z":1}`))
	assert.NoError(t, err)
	assert.Equal(t, 1, zone.ID)
}

func TestZonesService_GetWorldZone(t *testing.T) {
	repo := newFakeZonesRepo()
	worldID := uuid.New()
	repo.zones[worldID] = []*models.WorldZone{{ID: 1, WorldID: worldID}}

	svc := NewZonesService(config.CreateConfig(), repo, &fakeZonesRegistry{})
	zone, err := svc.GetWorldZone(worldID, 1)
	assert.NoError(t, err)
	assert.Equal(t, 1, zone.ID)
}

func TestZonesService_GetWorldZones(t *testing.T) {
	repo := newFakeZonesRepo()
	worldID := uuid.New()
	repo.zones[worldID] = []*models.WorldZone{{ID: 1, WorldID: worldID}, {ID: 2, WorldID: worldID}}

	svc := NewZonesService(config.CreateConfig(), repo, &fakeZonesRegistry{})
	zones, err := svc.GetWorldZones(worldID)
	assert.NoError(t, err)
	assert.Len(t, zones, 2)
}
