package server_registry

import (
	"net/http"
	"strconv"

	"github.com/FeedTheRealm-org/core-service/config"
	"github.com/FeedTheRealm-org/core-service/internal/common_handlers"
	"github.com/FeedTheRealm-org/core-service/internal/errors"
	"github.com/FeedTheRealm-org/core-service/internal/world-service/dtos"
	"github.com/FeedTheRealm-org/core-service/internal/world-service/services/server_registry"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// ServerRegistryController handles ftr-server reports and WorldId,ZoneId -> IP:port mapping
type serverRegistryController struct {
	conf                  *config.Config
	nomadJobSenderService server_registry.ServerRegistryService
}

func NewServerRegistryController(conf *config.Config, nomadJobService server_registry.ServerRegistryService) ServerRegistryController {
	return &serverRegistryController{
		conf:                  conf,
		nomadJobSenderService: nomadJobService,
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
// @Success      200  {string}  string "Acknowledge Boot"
// @Failure      400  {object}  errors.HttpError
// @Failure      500  {object}  errors.HttpError
// @Router       /orchestrator/{id}/zone/{zone_id}/start [post]
func (c *serverRegistryController) StartNewJob(ctx *gin.Context) {
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

	err = c.nomadJobSenderService.StartNewJob(worldId, zoneId)
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
// @Failure      400  {object}  errors.HttpError
// @Failure      404  {object}  errors.HttpError
// @Router       /orchestrator/{id}/zone/{zone_id}/stop [post]
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
// @Failure      400  {object}  errors.HttpError
// @Failure      404  {object}  errors.HttpError
// @Router       /orchestrator/{id}/zone/{zone_id}/address [get]
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
