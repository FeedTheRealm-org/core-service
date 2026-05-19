package server_registry

import (
	"io"
	"net/http"
	"strconv"

	"github.com/FeedTheRealm-org/core-service/config"
	"github.com/FeedTheRealm-org/core-service/internal/common_handlers"
	"github.com/FeedTheRealm-org/core-service/internal/errors"
	"github.com/FeedTheRealm-org/core-service/internal/utils/logger"
	"github.com/FeedTheRealm-org/core-service/internal/world-service/dtos"
	"github.com/FeedTheRealm-org/core-service/internal/world-service/models"
	"github.com/FeedTheRealm-org/core-service/internal/world-service/services/server_registry"
	"github.com/FeedTheRealm-org/core-service/internal/world-service/services/world"
	"github.com/FeedTheRealm-org/core-service/internal/world-service/services/zones"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// ServerRegistryController handles ftr-server reports and WorldId,ZoneId -> IP:port mapping
type serverRegistryController struct {
	conf                  *config.Config
	worldService          world.WorldService
	zoneService           zones.ZonesService
	nomadJobSenderService server_registry.ServerRegistryService
}

func NewServerRegistryController(conf *config.Config, worldService world.WorldService, zoneService zones.ZonesService, nomadJobSenderService server_registry.ServerRegistryService) ServerRegistryController {
	return &serverRegistryController{
		conf:                  conf,
		worldService:          worldService,
		zoneService:           zoneService,
		nomadJobSenderService: nomadJobSenderService,
	}
}

// StartNewJob godoc
// @Summary      Start Nomad allocation Job
// @Description  Init an orchestrator game server container to map back to this World and Zone chunk.
// @Tags         world-service
// @Security     BearerAuth
// @Produce      json
// @Param        id path string true "World UUID"
// @Param        zone_id path int true "World Zone Number"
// @Param        test query bool false "Whether to start the job in test mode, which uses a different image and resource allocation"
// @Success      200  {string}  string "Acknowledge Boot"
// @Failure      400  {object} dtos.ErrorResponse
// @Failure      500  {object} dtos.ErrorResponse
// @Router       /world/orchestrator/{id}/zones/{zone_id}/start-job [get]
func (c *serverRegistryController) StartNewJob(ctx *gin.Context) {
	worldIdStr := ctx.Param("id")
	zoneIdStr := ctx.Param("zone_id")
	isTest := ctx.Query("test") == "true"

	worldId, err := uuid.Parse(worldIdStr)
	if err != nil {
		_ = ctx.Error(errors.NewBadRequestError("invalid world ID: " + worldIdStr))
		return
	}

	zoneId, err := strconv.Atoi(zoneIdStr)
	if err != nil {
		_ = ctx.Error(errors.NewBadRequestError("invalid zone ID: " + zoneIdStr))
		return
	}

	err = c.nomadJobSenderService.StartNewJob(worldId, zoneId, isTest)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	common_handlers.HandleBodilessResponse(ctx, http.StatusOK)
}

// StopJob godoc
// @Summary      Shutdown job execution
// @Description  Stop local processing chunk container linked to orchestrator mapping.
// @Tags         world-service
// @Security     BearerAuth
// @Produce      json
// @Param        id path string true "World UUID"
// @Param        zone_id path int true "World Zone Number"
// @Success      200  {string}  string "Acknowledge Stop"
// @Failure      400  {object} dtos.ErrorResponse
// @Failure      404  {object} dtos.ErrorResponse
// @Router       /world/orchestrator/{id}/zones/{zone_id}/stop-job [get]
func (c *serverRegistryController) StopJob(ctx *gin.Context) {
	worldIdStr := ctx.Param("id")
	zoneIdStr := ctx.Param("zone_id")

	worldId, err := uuid.Parse(worldIdStr)
	if err != nil {
		_ = ctx.Error(errors.NewBadRequestError("invalid world ID: " + worldIdStr))
		return
	}

	zoneId, err := strconv.Atoi(zoneIdStr)
	if err != nil {
		_ = ctx.Error(errors.NewBadRequestError("invalid zone ID: " + zoneIdStr))
		return
	}

	err = c.nomadJobSenderService.StopJob(worldId, zoneId)
	if err != nil {
		_ = ctx.Error(errors.NewNotFoundError(err.Error()))
		return
	}

	common_handlers.HandleBodilessResponse(ctx, http.StatusOK)
}

// GetServerAddress godoc
// @Summary      Fetch running container address
// @Description  Returns resolved dynamically routed TCP/UDP IP and Port mappings for client connections.
// @Tags         world-service
// @Security     BearerAuth
// @Produce      json
// @Param        id path string true "World UUID"
// @Param        zone_id path int true "World Zone Number"
// @Success      200  {object}  dtos.WorldAddressResponse
// @Failure      400  {object} dtos.ErrorResponse
// @Failure      404  {object} dtos.ErrorResponse
// @Router       /world/orchestrator/{id}/zones/{zone_id}/address [get]
func (c *serverRegistryController) GetServerAddress(ctx *gin.Context) {
	worldIdStr := ctx.Param("id")
	zoneIdStr := ctx.Param("zone_id")

	worldId, err := uuid.Parse(worldIdStr)
	if err != nil {
		_ = ctx.Error(errors.NewBadRequestError("invalid world ID: " + worldIdStr))
		return
	}

	zoneId, err := strconv.Atoi(zoneIdStr)
	if err != nil {
		_ = ctx.Error(errors.NewBadRequestError("invalid zone ID: " + zoneIdStr))
		return
	}

	addr, port, err := c.nomadJobSenderService.GetServerAddress(worldId, zoneId)
	if err != nil {
		_ = ctx.Error(errors.NewNotFoundError(err.Error()))
		return
	}

	common_handlers.HandleSuccessResponse(ctx, http.StatusOK, &dtos.WorldAddressResponse{
		IP:   addr,
		Port: port,
	})
}

// UpdateServer godoc
// @Summary      Update running servers
// @Description  Webhook to update running world servers
// @Tags         world-service
// @Produce      json
// @Success      200  {string}  string "OK"
// @Failure      401  {object} dtos.ErrorResponse
// @Failure      404  {object} dtos.ErrorResponse
// @Failure      500  {object} dtos.ErrorResponse
// @Router       /world/orchestrator/webhook/servers/update [post]
func (c *serverRegistryController) UpdateServer(ctx *gin.Context) {
	logger.Logger.Info("Received server update webhook call")

	// To be deleted
	rawBody, _ := io.ReadAll(ctx.Request.Body)
	logger.Logger.Infow("request log",
		"method", ctx.Request.Method,
		"path", ctx.Request.URL.Path,
		"query", ctx.Request.URL.RawQuery,
		"status", ctx.Writer.Status(),
		"request_body", string(rawBody),
	)

	if common_handlers.IsGithubOIDCTokenValid(ctx) != nil {
		_ = ctx.Error(errors.NewUnauthorizedError("Invalid GitHub OIDC token"))
		return
	}

	activeZones, err := c.worldService.GetActiveWorldZones()
	if err != nil {
		_ = ctx.Error(errors.NewNotFoundError(err.Error()))
		return
	}

	for _, zone := range activeZones {
		err := c.nomadJobSenderService.StartNewJob(zone.WorldID, zone.ID, false)
		if err != nil {
			_ = ctx.Error(errors.NewInternalServerError(err.Error()))
			return
		}
	}

	common_handlers.HandleBodilessResponse(ctx, http.StatusOK)
}

// UpdateStatus godoc
// @Summary      Update zone status
// @Description  Update the online status of a specific zone in a world.
// @Tags         world-service
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id path string true "World UUID"
// @Param        zone_id path int true "World Zone Number"
// @Param        request body dtos.UpdateStatusRequest true "Status Update Request"
// @Success      200  {string}  string "OK"
// @Failure      400  {object} dtos.ErrorResponse
// @Failure      500  {object} dtos.ErrorResponse
// @Router       /world/orchestrator/{id}/zones/{zone_id}/status [post]
func (c *serverRegistryController) UpdateStatus(ctx *gin.Context) {
	worldIdStr := ctx.Param("id")
	zoneIdStr := ctx.Param("zone_id")

	worldId, err := uuid.Parse(worldIdStr)
	if err != nil {
		_ = ctx.Error(errors.NewBadRequestError("invalid world ID: " + worldIdStr))
		return
	}

	zoneId, err := strconv.Atoi(zoneIdStr)
	if err != nil {
		_ = ctx.Error(errors.NewBadRequestError("invalid zone ID: " + zoneIdStr))
		return
	}

	var req dtos.UpdateStatusRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		_ = ctx.Error(errors.NewBadRequestError("invalid request body: " + err.Error()))
		return
	}

	err = c.zoneService.UpdateZoneStatus(worldId, zoneId, req.IsOnline)
	if err != nil {
		_ = ctx.Error(errors.NewInternalServerError(err.Error()))
		return
	}

	common_handlers.HandleBodilessResponse(ctx, http.StatusOK)
}

// UpdatePlayerCount godoc
// @Summary      Update zone player count
// @Description  Servers report active players and average player time every 2 minutes.
// @Tags         world-service
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id path string true "World UUID"
// @Param        zone_id path int true "World Zone Number"
// @Param        request body dtos.UpdatePlayerCountRequest true "Player count payload"
// @Success      200  {object}  dtos.WorldPlayerCountsResponse
// @Failure      400  {object} dtos.ErrorResponse
// @Failure      500  {object} dtos.ErrorResponse
// @Router       /world/orchestrator/{id}/zones/{zone_id}/players [post]
func (c *serverRegistryController) UpdatePlayerCount(ctx *gin.Context) {
	worldIdStr := ctx.Param("id")
	zoneIdStr := ctx.Param("zone_id")

	worldId, err := uuid.Parse(worldIdStr)
	if err != nil {
		_ = ctx.Error(errors.NewBadRequestError("invalid world ID: " + worldIdStr))
		return
	}

	zoneId, err := strconv.Atoi(zoneIdStr)
	if err != nil {
		_ = ctx.Error(errors.NewBadRequestError("invalid zone ID: " + zoneIdStr))
		return
	}

	var req dtos.UpdatePlayerCountRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		_ = ctx.Error(errors.NewBadRequestError("invalid request body: " + err.Error()))
		return
	}

	zone, err := c.zoneService.GetWorldZone(worldId, zoneId)
	if err != nil {
		_ = ctx.Error(errors.NewNotFoundError("zone not found"))
		return
	}

	activePlayers := req.ActivePlayers
	averagePlayerTime := req.AveragePlayerTime
	if !zone.IsOnline || !zone.IsActive {
		activePlayers = 0
		averagePlayerTime = 0
	}

	if err := c.zoneService.UpdateZonePlayerCount(worldId, zoneId, activePlayers, averagePlayerTime); err != nil {
		_ = ctx.Error(errors.NewInternalServerError(err.Error()))
		return
	}

	response := c.buildWorldPlayerCounts(worldId)
	common_handlers.HandleSuccessResponse(ctx, http.StatusOK, response)
}

// GetWorldPlayerCounts godoc
// @Summary      Get world player counts
// @Description  Returns player counts per zone for a world.
// @Tags         world-service
// @Security     BearerAuth
// @Produce      json
// @Param        id path string true "World UUID"
// @Success      200  {object}  dtos.WorldPlayerCountsResponse
// @Failure      400  {object} dtos.ErrorResponse
// @Failure      500  {object} dtos.ErrorResponse
// @Router       /world/orchestrator/{id}/players [get]
func (c *serverRegistryController) GetWorldPlayerCounts(ctx *gin.Context) {
	worldIdStr := ctx.Param("id")
	worldId, err := uuid.Parse(worldIdStr)
	if err != nil {
		_ = ctx.Error(errors.NewBadRequestError("invalid world ID: " + worldIdStr))
		return
	}

	response := c.buildWorldPlayerCounts(worldId)
	common_handlers.HandleSuccessResponse(ctx, http.StatusOK, response)
}

// GetAllWorldPlayerCounts godoc
// @Summary      Get all world player counts
// @Description  Returns player counts for all worlds.
// @Tags         world-service
// @Security     BearerAuth
// @Produce      json
// @Success      200  {array}  dtos.WorldPlayerCountsResponse
// @Failure      500  {object} dtos.ErrorResponse
// @Router       /world/orchestrator/players [get]
func (c *serverRegistryController) GetAllWorldPlayerCounts(ctx *gin.Context) {
	zones, err := c.zoneService.GetAllWorldZonePlayerCounts()
	if err != nil {
		_ = ctx.Error(errors.NewInternalServerError(err.Error()))
		return
	}

	responses := groupZonesByWorld(zones)
	common_handlers.HandleSuccessResponse(ctx, http.StatusOK, responses)
}

func (c *serverRegistryController) buildWorldPlayerCounts(worldId uuid.UUID) dtos.WorldPlayerCountsResponse {
	zones, err := c.zoneService.GetWorldZonePlayerCounts(worldId)
	if err != nil {
		return dtos.WorldPlayerCountsResponse{
			WorldID:           worldId.String(),
			TotalPlayers:      0,
			AveragePlayerTime: 0,
			Zones:             []dtos.ZonePlayerCountResponse{},
		}
	}

	responses := groupZonesByWorld(zones)
	for _, res := range responses {
		if res.WorldID == worldId.String() {
			return res
		}
	}

	return dtos.WorldPlayerCountsResponse{
		WorldID:           worldId.String(),
		TotalPlayers:      0,
		AveragePlayerTime: 0,
		Zones:             []dtos.ZonePlayerCountResponse{},
	}
}

func groupZonesByWorld(zones []*models.WorldZone) []dtos.WorldPlayerCountsResponse {
	worldMap := make(map[string]*dtos.WorldPlayerCountsResponse)
	for _, zone := range zones {
		worldId := zone.WorldID.String()
		entry, ok := worldMap[worldId]
		if !ok {
			entry = &dtos.WorldPlayerCountsResponse{
				WorldID:           worldId,
				TotalPlayers:      0,
				AveragePlayerTime: 0,
				Zones:             []dtos.ZonePlayerCountResponse{},
			}
			worldMap[worldId] = entry
		}
		entry.TotalPlayers += zone.ActivePlayers
		entry.Zones = append(entry.Zones, dtos.ZonePlayerCountResponse{
			ZoneID:            zone.ID,
			ActivePlayers:     zone.ActivePlayers,
			AveragePlayerTime: zone.AveragePlayerTime,
			UpdatedAt:         zone.PlayerCountUpdatedAt,
		})
	}

	responses := make([]dtos.WorldPlayerCountsResponse, 0, len(worldMap))
	for _, response := range worldMap {
		response.AveragePlayerTime = averageZoneTime(response.Zones)
		responses = append(responses, *response)
	}
	return responses
}

func averageZoneTime(zones []dtos.ZonePlayerCountResponse) int {
	if len(zones) == 0 {
		return 0
	}

	total := 0
	for _, zone := range zones {
		total += zone.AveragePlayerTime
	}

	return int(float64(total)/float64(len(zones)) + 0.5)
}
