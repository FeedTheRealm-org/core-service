package zones

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/FeedTheRealm-org/core-service/config"
	"github.com/FeedTheRealm-org/core-service/internal/common_handlers"
	"github.com/FeedTheRealm-org/core-service/internal/errors"
	"github.com/FeedTheRealm-org/core-service/internal/world-service/dtos"
	world_errors "github.com/FeedTheRealm-org/core-service/internal/world-service/errors"
	zones_service "github.com/FeedTheRealm-org/core-service/internal/world-service/services/zones"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type zonesController struct {
	conf         *config.Config
	zonesService zones_service.ZonesService
}

func NewZonesController(conf *config.Config, zonesService zones_service.ZonesService) ZonesController {
	return &zonesController{
		conf:         conf,
		zonesService: zonesService,
	}
}

// PublishZone godoc
// @Summary      Publish zone
// @Description  Publishes or updates a zone for a world by world_id and zone_id.
// @Tags         world-service
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id path string true "World UUID"
// @Param        zone_id path int true "Zone ID"
// @Param        request body dtos.PublishZoneRequest true "Zone publish data"
// @Success      200  {object}  dtos.WorldZoneResponse
// @Failure      400  {object} dtos.ErrorResponse
// @Failure      401  {object} dtos.ErrorResponse
// @Failure      404  {object} dtos.ErrorResponse
// @Router       /world/{id}/zones/{zone_id} [put]
func (c *zonesController) PublishZone(ctx *gin.Context) {
	userId, err := common_handlers.GetUserIDFromSession(ctx)
	if err != nil {
		_ = ctx.Error(errors.NewUnauthorizedError(err.Error()))
		return
	}

	worldID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		_ = ctx.Error(errors.NewBadRequestError("invalid world_id: " + ctx.Param("id")))
		return
	}

	zoneID, err := strconv.Atoi(ctx.Param("zone_id"))
	if err != nil || zoneID <= 0 {
		_ = ctx.Error(errors.NewBadRequestError("zone_id must be a positive integer"))
		return
	}

	worldInfo, err := c.zonesService.GetWorld(worldID)
	if err != nil {
		if _, ok := err.(*world_errors.WorldInfoNotFound); ok {
			_ = ctx.Error(errors.NewNotFoundError("world info not found"))
			return
		}
		_ = ctx.Error(err)
		return
	}

	if worldInfo.UserId != userId {
		_ = ctx.Error(errors.NewUnauthorizedError("user does not own this world"))
		return
	}

	var req dtos.PublishZoneRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		_ = ctx.Error(errors.NewBadRequestError("invalid JSON payload: " + err.Error()))
		return
	}

	zoneDataBytes, err := json.Marshal(req.Data)
	if err != nil {
		_ = ctx.Error(errors.NewBadRequestError("failed to marshal zone data: " + err.Error()))
		return
	}

	zone, err := c.zonesService.PublishZone(worldID, zoneID, zoneDataBytes)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	common_handlers.HandleSuccessResponse(ctx, http.StatusOK, &dtos.WorldZoneResponse{
		WorldID:  zone.WorldID.String(),
		ZoneID:   zone.ID,
		ZoneData: string(zone.ZoneData),
		IsActive: zone.IsActive,
	})
}

// GetWorldZones godoc
// @Summary      Retrieve zones for a world
// @Description  Returns all available zone IDs for a specific world.
// @Tags         world-service
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id path string true "World UUID"
// @Success      200  {object}  dtos.WorldZonesResponse
// @Failure      400  {object} dtos.ErrorResponse
// @Failure      401  {object} dtos.ErrorResponse
// @Failure      404  {object} dtos.ErrorResponse
// @Router       /world/{id}/zones [get]
func (c *zonesController) GetWorldZones(ctx *gin.Context) {
	_, err := common_handlers.GetUserIDFromSession(ctx)
	if err != nil {
		_ = ctx.Error(errors.NewUnauthorizedError(err.Error()))
		return
	}

	worldID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		_ = ctx.Error(errors.NewBadRequestError("invalid world ID: " + ctx.Param("id")))
		return
	}

	zones, err := c.zonesService.GetWorldZones(worldID)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	zoneMetadata := make([]dtos.WorldZoneMetadata, 0, len(zones))
	for _, zone := range zones {
		zoneMetadata = append(zoneMetadata, dtos.WorldZoneMetadata{
			ZoneID:   zone.ID,
			IsActive: zone.IsActive,
		})
	}

	common_handlers.HandleSuccessResponse(ctx, http.StatusOK, &dtos.WorldZonesResponse{
		WorldID: worldID.String(),
		Zones:   zoneMetadata,
	})
}

// GetWorldZoneData godoc
// @Summary      Retrieve specific zone data
// @Description  Returns data for a specific zone in a world.
// @Tags         world-service
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id path string true "World UUID"
// @Param        zone_id path int true "Zone ID"
// @Success      200  {object}  dtos.WorldZoneResponse
// @Failure      400  {object} dtos.ErrorResponse
// @Failure      401  {object} dtos.ErrorResponse
// @Failure      404  {object} dtos.ErrorResponse
// @Router       /world/{id}/zones/{zone_id} [get]
func (c *zonesController) GetWorldZoneData(ctx *gin.Context) {
	_, err := common_handlers.GetUserIDFromSession(ctx)
	if err != nil {
		_ = ctx.Error(errors.NewUnauthorizedError(err.Error()))
		return
	}

	worldID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		_ = ctx.Error(errors.NewBadRequestError("invalid world ID: " + ctx.Param("id")))
		return
	}

	zoneID, err := strconv.Atoi(ctx.Param("zone_id"))
	if err != nil {
		_ = ctx.Error(errors.NewBadRequestError("invalid zone ID: " + ctx.Param("zone_id")))
		return
	}

	zone, err := c.zonesService.GetWorldZone(worldID, zoneID)
	if err != nil {
		_ = ctx.Error(errors.NewNotFoundError("zone not found"))
		return
	}

	common_handlers.HandleSuccessResponse(ctx, http.StatusOK, &dtos.WorldZoneResponse{
		WorldID:  worldID.String(),
		ZoneID:   zone.ID,
		ZoneData: zone.ZoneData.String(),
		IsActive: zone.IsActive,
	})
}

// ActivateZone godoc
// @Summary      Activate zone
// @Description  Starts server orchestration for a published zone and consumes one subscription slot.
// @Tags         world-service
// @Security     BearerAuth
// @Produce      json
// @Param        id path string true "World UUID"
// @Param        zone_id path int true "Zone ID"
// @Success      200  {string}  string "Acknowledge activation"
// @Failure      400  {object} dtos.ErrorResponse
// @Failure      401  {object} dtos.ErrorResponse
// @Failure      404  {object} dtos.ErrorResponse
// @Router       /world/{id}/zones/{zone_id}/activate [get]
func (c *zonesController) ActivateZone(ctx *gin.Context) {
	userId, err := common_handlers.GetUserIDFromSession(ctx)
	if err != nil {
		_ = ctx.Error(errors.NewUnauthorizedError(err.Error()))
		return
	}

	worldID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		_ = ctx.Error(errors.NewBadRequestError("invalid world_id: " + ctx.Param("id")))
		return
	}

	zoneID, err := strconv.Atoi(ctx.Param("zone_id"))
	if err != nil || zoneID <= 0 {
		_ = ctx.Error(errors.NewBadRequestError("zone_id must be a positive integer"))
		return
	}

	worldInfo, err := c.zonesService.GetWorld(worldID)
	if err != nil {
		if _, ok := err.(*world_errors.WorldInfoNotFound); ok {
			_ = ctx.Error(errors.NewNotFoundError("world info not found"))
			return
		}
		_ = ctx.Error(err)
		return
	}

	if worldInfo.UserId != userId {
		_ = ctx.Error(errors.NewUnauthorizedError("user does not own this world"))
		return
	}

	err = c.zonesService.ActivateZone(worldID, zoneID)
	if err != nil {
		_ = ctx.Error(errors.NewForbiddenError(err.Error()))
		return
	}

	common_handlers.HandleBodilessResponse(ctx, http.StatusOK)
}

// DeactivateZone godoc
// @Summary      Deactivate zone
// @Description  Stops server orchestration for an active zone and releases one subscription slot.
// @Tags         world-service
// @Security     BearerAuth
// @Produce      json
// @Param        id path string true "World UUID"
// @Param        zone_id path int true "Zone ID"
// @Success      200  {string}  string "Acknowledge deactivation"
// @Failure      400  {object} dtos.ErrorResponse
// @Failure      401  {object} dtos.ErrorResponse
// @Failure      404  {object} dtos.ErrorResponse
// @Router       /world/{id}/zones/{zone_id}/deactivate [get]
func (c *zonesController) DeactivateZone(ctx *gin.Context) {
	userId, err := common_handlers.GetUserIDFromSession(ctx)
	if err != nil {
		_ = ctx.Error(errors.NewUnauthorizedError(err.Error()))
		return
	}

	worldID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		_ = ctx.Error(errors.NewBadRequestError("invalid world_id: " + ctx.Param("id")))
		return
	}

	zoneID, err := strconv.Atoi(ctx.Param("zone_id"))
	if err != nil || zoneID <= 0 {
		_ = ctx.Error(errors.NewBadRequestError("zone_id must be a positive integer"))
		return
	}

	worldInfo, err := c.zonesService.GetWorld(worldID)
	if err != nil {
		if _, ok := err.(*world_errors.WorldInfoNotFound); ok {
			_ = ctx.Error(errors.NewNotFoundError("world info not found"))
			return
		}
		_ = ctx.Error(err)
		return
	}

	if worldInfo.UserId != userId {
		_ = ctx.Error(errors.NewUnauthorizedError("user does not own this world"))
		return
	}

	err = c.zonesService.DeactivateZone(worldID, zoneID)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	common_handlers.HandleBodilessResponse(ctx, http.StatusOK)
}
