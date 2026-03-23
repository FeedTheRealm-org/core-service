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
