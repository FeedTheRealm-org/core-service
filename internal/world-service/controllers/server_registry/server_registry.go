package server_registry

import (
	"net/http"
	"strconv"

	"github.com/FeedTheRealm-org/core-service/config"
	"github.com/FeedTheRealm-org/core-service/internal/common_handlers"
	"github.com/FeedTheRealm-org/core-service/internal/errors"
	"github.com/FeedTheRealm-org/core-service/internal/world-service/dtos"
	"github.com/FeedTheRealm-org/core-service/internal/world-service/services/server_registry"
	"github.com/FeedTheRealm-org/core-service/internal/world-service/services/world"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// ServerRegistryController handles ftr-server reports and WorldId,ZoneId -> IP:port mapping
type serverRegistryController struct {
	conf                  *config.Config
	worldService          world.WorldService
	nomadJobSenderService server_registry.ServerRegistryService
}

func NewServerRegistryController(conf *config.Config, worldService world.WorldService, nomadJobSenderService server_registry.ServerRegistryService) ServerRegistryController {
	return &serverRegistryController{
		conf:                  conf,
		worldService:          worldService,
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
// @Router       /world/orchestrator/webhook/servers/update [post]
func (c *serverRegistryController) UpdateServer(ctx *gin.Context) {
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
			_ = ctx.Error(errors.NewNotFoundError(err.Error()))
			return
		}
	}

	common_handlers.HandleBodilessResponse(ctx, http.StatusOK)
}
