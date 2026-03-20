package nomad_job_sender

import (
	"net/http"
	"strconv"

	"github.com/FeedTheRealm-org/core-service/config"
	"github.com/FeedTheRealm-org/core-service/internal/common_handlers"
	"github.com/FeedTheRealm-org/core-service/internal/errors"
	"github.com/FeedTheRealm-org/core-service/internal/world-service/services/nomad_job_sender"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type nomadJobSenderController struct {
	conf                  *config.Config
	nomadJobSenderService nomad_job_sender.NomadJobSenderService
}

func NewNomadJobSenderController(conf *config.Config, nomadJobSenderService nomad_job_sender.NomadJobSenderService) NomadJobSenderController {
	return &nomadJobSenderController{
		conf:                  conf,
		nomadJobSenderService: nomadJobSenderService,
	}
}

func (c *nomadJobSenderController) StartNewJob(ctx *gin.Context) {
	worldIdStr := ctx.Param("world_id")
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

	common_handlers.HandleSuccessResponse(ctx, http.StatusOK, "Job started successfully")
}
