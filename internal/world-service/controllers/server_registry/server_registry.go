package server_registry

import (
	"net/http"
	"strconv"

	"github.com/FeedTheRealm-org/core-service/config"
	"github.com/FeedTheRealm-org/core-service/internal/common_handlers"
	"github.com/FeedTheRealm-org/core-service/internal/errors"
	"github.com/FeedTheRealm-org/core-service/internal/utils/logger"
	"github.com/FeedTheRealm-org/core-service/internal/world-service/dtos"
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
		_ = ctx.Error(errors.NewNotFoundError("Could not stop job."))
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
		_ = ctx.Error(errors.NewNotFoundError("Failed to get server address."))
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

	if common_handlers.IsGithubOIDCTokenValid(ctx) != nil {
		_ = ctx.Error(errors.NewUnauthorizedError("Invalid GitHub OIDC token"))
		return
	}

	activeZones, err := c.worldService.GetActiveWorldZones()
	if err != nil {
		_ = ctx.Error(errors.NewNotFoundError("Failed to get active world zones."))
		return
	}

	for _, zone := range activeZones {
		err := c.nomadJobSenderService.StartNewJob(zone.WorldID, zone.ID, false)
		if err != nil {
			_ = ctx.Error(errors.NewInternalServerError("Failed to start new job."))
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
		_ = ctx.Error(errors.NewInternalServerError("Failed to update zone status."))
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
// @Success      200  {string}  string "OK"
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

	if err := c.zoneService.UpdateZonePlayerCount(worldId, zoneId, req.ActivePlayers, req.AveragePlayerTime); err != nil {
		_ = ctx.Error(errors.NewInternalServerError("Failed to update zone player count."))
		return
	}

	common_handlers.HandleBodilessResponse(ctx, http.StatusOK)
}

// GetWorldPlayerCounts godoc
// @Summary      Get world player counts
// @Description  Returns player counts per zone for a world.
// @Tags         world-service
// @Security     BearerAuth
// @Produce      json
// @Param        id path string true "World UUID"
// @Success      200  {object}  dtos.PlayerCountsResponse
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

	activePlayers, averagePlayerTime, err := c.zoneService.GetWorldZonePlayerCounts(worldId)
	if err != nil {
		_ = ctx.Error(errors.NewInternalServerError("Failed to get world player count."))
		return
	}

	response := &dtos.PlayerCountsResponse{
		ActivePlayers:     activePlayers,
		AveragePlayerTime: averagePlayerTime,
	}

	common_handlers.HandleSuccessResponse(ctx, http.StatusOK, response)
}

// GetAllWorldPlayerCounts godoc
// @Summary      Get all world player counts
// @Description  Returns player counts for all worlds.
// @Tags         world-service
// @Security     BearerAuth
// @Produce      json
// @Success      200  {array}  dtos.PlayerCountsResponse
// @Failure      500  {object} dtos.ErrorResponse
// @Router       /world/orchestrator/players [get]
func (c *serverRegistryController) GetAllWorldPlayerCounts(ctx *gin.Context) {
	activePlayers, averagePlayerTime, err := c.zoneService.GetAllWorldZonePlayerCounts()
	if err != nil {
		_ = ctx.Error(errors.NewInternalServerError("Failed to get all world player counts."))
		return
	}

	responses := &dtos.PlayerCountsResponse{
		ActivePlayers:     activePlayers,
		AveragePlayerTime: averagePlayerTime,
	}

	common_handlers.HandleSuccessResponse(ctx, http.StatusOK, responses)
}
