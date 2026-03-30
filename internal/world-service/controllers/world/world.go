package world

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/FeedTheRealm-org/core-service/config"
	"github.com/FeedTheRealm-org/core-service/internal/common_handlers"
	"github.com/FeedTheRealm-org/core-service/internal/errors"
	"github.com/FeedTheRealm-org/core-service/internal/utils/input_validation"
	"github.com/FeedTheRealm-org/core-service/internal/world-service/dtos"
	world_errors "github.com/FeedTheRealm-org/core-service/internal/world-service/errors"
	"github.com/FeedTheRealm-org/core-service/internal/world-service/models"
	"github.com/FeedTheRealm-org/core-service/internal/world-service/services/world"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/datatypes"
)

type worldController struct {
	conf         *config.Config
	worldService world.WorldService
}

// NewWorldController creates a new instance of CharacterController.
func NewWorldController(conf *config.Config, characterService world.WorldService) WorldController {
	return &worldController{
		conf:         conf,
		worldService: characterService,
	}
}

// PublishWorld godoc
// @Summary      Publish new world
// @Description  Create a new world instance in the registry.
// @Tags         world-service
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        request body dtos.WorldRequest true "World creation data"
// @Success      201  {object}  dtos.WorldResponse
// @Failure      400  {object} dtos.ErrorResponse
// @Failure      401  {object} dtos.ErrorResponse
// @Failure      409  {object} dtos.ErrorResponse
// @Router       /world [post]
func (c *worldController) PublishWorld(ctx *gin.Context) {
	userId, err := common_handlers.GetUserIDFromSession(ctx)
	if err != nil {
		_ = ctx.Error(errors.NewUnauthorizedError(err.Error()))
		return
	}

	var req dtos.WorldRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		_ = ctx.Error(errors.NewBadRequestError("invalid JSON payload: " + err.Error()))
		return
	}

	// TODO: Add basic validation for world data structure, currently just storing raw data
	bytes, _ := json.Marshal(req.Data)

	if len(req.FileName) < 6 || len(req.FileName) > 24 {
		_ = ctx.Error(errors.NewBadRequestError("world name must be between 6 and 24 characters"))
		return
	}

	if input_validation.ValidateInvalidCharacters(req.FileName) || input_validation.HasSpaces(req.FileName) {
		_ = ctx.Error(errors.NewBadRequestError("world name contains invalid special characters"))
		return
	}

	worldData := &models.WorldData{
		UserId:         userId,
		Name:           req.FileName,
		Description:    req.Description,
		Data:           datatypes.JSON(bytes),
		CreateableData: datatypes.JSON([]byte("{}")),
	}
	if req.CreateableData != nil {
		createableBytes, err := json.Marshal(req.CreateableData)
		if err != nil {
			_ = ctx.Error(errors.NewBadRequestError("failed to marshal createable_data: " + err.Error()))
			return
		}
		worldData.CreateableData = datatypes.JSON(createableBytes)
	}

	createdWorld, err := c.worldService.PublishWorld(worldData)

	if err != nil {
		if _, ok := err.(*world_errors.WorldNameTaken); ok {
			_ = ctx.Error(errors.NewConflictError("world name is already taken"))
			return
		}
		_ = ctx.Error(err)
		return
	}

	response := &dtos.WorldResponse{
		ID:             createdWorld.ID.String(),
		UserId:         createdWorld.UserId.String(),
		Name:           createdWorld.Name,
		Description:    createdWorld.Description,
		Data:           createdWorld.Data.String(),
		CreateableData: createdWorld.CreateableData.String(),
		CreatedAt:      createdWorld.CreatedAt,
		UpdatedAt:      createdWorld.UpdatedAt,
	}

	common_handlers.HandleSuccessResponse(ctx, http.StatusCreated, response)
}

// GetWorld godoc
// @Summary      Retrieve world detail
// @Description  Fetches full world payload and configuration data by passing the world ID.
// @Tags         world-service
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id path string true "World UUID"
// @Success      200  {object}  dtos.WorldResponse
// @Failure      400  {object} dtos.ErrorResponse
// @Failure      401  {object} dtos.ErrorResponse
// @Failure      404  {object} dtos.ErrorResponse
// @Router       /world/{id} [get]
func (c *worldController) GetWorld(ctx *gin.Context) {
	_, err := common_handlers.GetUserIDFromSession(ctx)
	if err != nil {
		_ = ctx.Error(errors.NewUnauthorizedError(err.Error()))
		return
	}

	var parsedWorldId uuid.UUID
	worldId := ctx.Param("id")
	if worldId == "" {
		_ = ctx.Error(errors.NewBadRequestError("World ID is required"))
		return
	}

	parsedWorldId, err = uuid.Parse(worldId)
	if err != nil {
		_ = ctx.Error(errors.NewBadRequestError("invalid world ID: " + worldId))
		return
	}

	worldInfo, err := c.worldService.GetWorld(parsedWorldId)
	if err != nil {
		if _, ok := err.(*world_errors.WorldInfoNotFound); ok {
			_ = ctx.Error(errors.NewNotFoundError("world info not found"))
			return
		}
		_ = ctx.Error(err)
		return
	}

	res := &dtos.WorldResponse{
		ID:             worldInfo.ID.String(),
		UserId:         worldInfo.UserId.String(),
		Name:           worldInfo.Name,
		Description:    worldInfo.Description,
		Data:           worldInfo.Data.String(),
		CreateableData: worldInfo.CreateableData.String(),
		CreatedAt:      worldInfo.CreatedAt,
		UpdatedAt:      worldInfo.UpdatedAt,
	}

	common_handlers.HandleSuccessResponse(ctx, http.StatusOK, res)
}

// GetWorldsList godoc
// @Summary      List worlds
// @Description  Generates paginated meta-list of standard player-made worlds.
// @Tags         world-service
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        offset query int false "Offset for pagination"
// @Param        limit query int false "Max hits per page"
// @Param        filter query string false "Search filters"
// @Success      200  {object}  dtos.WorldsListResponse
// @Failure      400  {object} dtos.ErrorResponse
// @Failure      401  {object} dtos.ErrorResponse
// @Failure      404  {object} dtos.ErrorResponse
// @Router       /world [get]
func (c *worldController) GetWorldsList(ctx *gin.Context) {
	_, err := common_handlers.GetUserIDFromSession(ctx)
	if err != nil {
		_ = ctx.Error(errors.NewUnauthorizedError(err.Error()))
		return
	}

	offsetStr := ctx.Query("offset")
	limitStr := ctx.Query("limit")
	filter := ctx.Query("filter")

	if offsetStr == "" || limitStr == "" {
		_ = ctx.Error(errors.NewBadRequestError("offset and limit are required"))
		return
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		_ = ctx.Error(errors.NewBadRequestError("invalid offset: " + offsetStr))
		return
	}
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 || limit > 100 {
		_ = ctx.Error(errors.NewBadRequestError("invalid limit: " + limitStr))
		return
	}

	worldsList, err := c.worldService.GetWorldsList(offset, limit, filter)
	if err != nil {
		if _, ok := err.(*world_errors.WorldInfoNotFound); ok {
			_ = ctx.Error(errors.NewNotFoundError("world info not found"))
			return
		}
		_ = ctx.Error(err)
		return
	}

	// Build a metadata-only list (do not include full world Data)
	resList := make([]dtos.WorldMetadata, 0, len(worldsList))
	for _, worldInfo := range worldsList {
		resList = append(resList, dtos.WorldMetadata{
			ID:             worldInfo.ID.String(),
			UserId:         worldInfo.UserId.String(),
			Name:           worldInfo.Name,
			Description:    worldInfo.Description,
			CreateableData: worldInfo.CreateableData.String(),
			CreatedAt:      worldInfo.CreatedAt,
			UpdatedAt:      worldInfo.UpdatedAt,
		})
	}
	res := &dtos.WorldsListResponse{
		Worlds: resList,
		Total:  len(resList),
		Limit:  limit,
		Offset: offset,
	}

	common_handlers.HandleSuccessResponse(ctx, http.StatusOK, res)
}

// UpdateWorld godoc
// @Summary      Modifies existing world data
// @Description  Changes the underlying data mappings and description values of an owned world instance.
// @Tags         world-service
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id path string true "World UUID"
// @Param        request body dtos.WorldRequest true "Updated JSON configuration block"
// @Success      200  {object}  dtos.WorldResponse
// @Failure      400  {object} dtos.ErrorResponse
// @Failure      401  {object} dtos.ErrorResponse
// @Failure      404  {object} dtos.ErrorResponse
// @Router       /world/{id} [put]
func (c *worldController) UpdateWorld(ctx *gin.Context) {
	userId, err := common_handlers.GetUserIDFromSession(ctx)
	if err != nil {
		_ = ctx.Error(errors.NewUnauthorizedError(err.Error()))
		return
	}

	worldId := ctx.Param("id")
	if worldId == "" {
		_ = ctx.Error(errors.NewBadRequestError("World ID is required"))
		return
	}

	parsedWorldId, err := uuid.Parse(worldId)
	if err != nil {
		_ = ctx.Error(errors.NewBadRequestError("invalid world ID: " + worldId))
		return
	}

	// Ownership check: fetch world and compare userId
	worldInfo, err := c.worldService.GetWorld(parsedWorldId)
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

	var req dtos.WorldRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		_ = ctx.Error(errors.NewBadRequestError("invalid JSON payload: " + err.Error()))
		return
	}

	bytes, err := json.Marshal(req.Data)
	if err != nil {
		_ = ctx.Error(errors.NewBadRequestError("failed to marshal world data: " + err.Error()))
		return
	}

	updatedWorld, err := c.worldService.UpdateWorld(parsedWorldId, userId, bytes, req.Description)
	if err != nil {
		if _, ok := err.(*world_errors.WorldInfoNotFound); ok {
			_ = ctx.Error(errors.NewNotFoundError("world info not found"))
			return
		}
		_ = ctx.Error(err)
		return
	}

	res := &dtos.WorldResponse{
		ID:             updatedWorld.ID.String(),
		UserId:         updatedWorld.UserId.String(),
		Name:           updatedWorld.Name,
		Description:    updatedWorld.Description,
		Data:           updatedWorld.Data.String(),
		CreateableData: updatedWorld.CreateableData.String(),
		CreatedAt:      updatedWorld.CreatedAt,
		UpdatedAt:      updatedWorld.UpdatedAt,
	}

	common_handlers.HandleSuccessResponse(ctx, http.StatusOK, res)
}

// DeleteWorld godoc
// @Summary      Destroy a world
// @Description  Delete a world permanently. Requires ownership.
// @Tags         world-service
// @Security     BearerAuth
// @Produce      json
// @Param        id path string true "World UUID"
// @Success      200  {string}  string "Success Message Payload"
// @Failure      400  {object} dtos.ErrorResponse
// @Failure      401  {object} dtos.ErrorResponse
// @Router       /world/{id} [delete]
func (c *worldController) DeleteWorld(ctx *gin.Context) {
	userId, err := common_handlers.GetUserIDFromSession(ctx)
	if err != nil {
		_ = ctx.Error(errors.NewUnauthorizedError(err.Error()))
		return
	}

	worldId := ctx.Param("id")
	if worldId == "" {
		_ = ctx.Error(errors.NewBadRequestError("World ID is required"))
		return
	}

	parsedWorldId, err := uuid.Parse(worldId)
	if err != nil {
		_ = ctx.Error(errors.NewBadRequestError("invalid world ID: " + worldId))
		return
	}

	if err := c.worldService.DeleteWorld(parsedWorldId, userId); err != nil {
		_ = ctx.Error(errors.NewUnauthorizedError(err.Error()))
		return
	}

	common_handlers.HandleBodilessResponse(ctx, http.StatusOK)
}

// ResetDatabase godoc
// @Summary      Development DB reset
// @Description  Purge utility used carefully during development to wipe all tables.
// @Tags         world-service
// @Produce      json
// @Success      200  {string}  string "Success reset"
// @Failure      500  {object} dtos.ErrorResponse
// @Router       /world/reset-database [delete]
func (c *worldController) ResetDatabase(ctx *gin.Context) {

	err := c.worldService.ClearDatabase()
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	common_handlers.HandleSuccessResponse(ctx, http.StatusOK, "Successfully reset database")
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
func (c *worldController) GetWorldZones(ctx *gin.Context) {
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

	zones, err := c.worldService.GetWorldZones(worldID)
	if err != nil {
		if _, ok := err.(*world_errors.WorldInfoNotFound); ok {
			_ = ctx.Error(errors.NewNotFoundError("world info not found"))
			return
		}
		_ = ctx.Error(err)
		return
	}

	zoneIds := make([]int, 0, len(zones))
	for _, zone := range zones {
		zoneIds = append(zoneIds, zone.ID)
	}

	common_handlers.HandleSuccessResponse(ctx, http.StatusOK, &dtos.WorldZonesResponse{
		WorldID: worldID.String(),
		Zones:   zoneIds,
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
func (c *worldController) GetWorldZoneData(ctx *gin.Context) {
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

	zone, err := c.worldService.GetWorldZone(worldID, zoneID)
	if err != nil {
		_ = ctx.Error(errors.NewNotFoundError("zone not found"))
		return
	}

	common_handlers.HandleSuccessResponse(ctx, http.StatusOK, &dtos.WorldZoneResponse{
		WorldID:  worldID.String(),
		ZoneID:   zone.ID,
		ZoneData: zone.ZoneData.String(),
	})
}

// UpdateCreateableData godoc
// @Summary      Update world createable data
// @Description  Updates createable_data by world ID.
// @Tags         world-service
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id path string true "World UUID"
// @Param        request body dtos.UpdateCreateableDataRequest true "Createable data"
// @Success      200  {object}  dtos.WorldResponse
// @Failure      400  {object} dtos.ErrorResponse
// @Failure      401  {object} dtos.ErrorResponse
// @Failure      404  {object} dtos.ErrorResponse
// @Router       /world/{id}/createable-data [put]
func (c *worldController) UpdateCreateableData(ctx *gin.Context) {
	userId, err := common_handlers.GetUserIDFromSession(ctx)
	if err != nil {
		_ = ctx.Error(errors.NewUnauthorizedError(err.Error()))
		return
	}

	worldID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		_ = ctx.Error(errors.NewBadRequestError("invalid world ID: " + ctx.Param("id")))
		return
	}

	var req dtos.UpdateCreateableDataRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		_ = ctx.Error(errors.NewBadRequestError("invalid JSON payload: " + err.Error()))
		return
	}

	bytes, err := json.Marshal(req.CreateableData)
	if err != nil {
		_ = ctx.Error(errors.NewBadRequestError("failed to marshal createable_data: " + err.Error()))
		return
	}

	updatedWorld, err := c.worldService.UpdateCreateableData(worldID, userId, bytes)
	if err != nil {
		if _, ok := err.(*world_errors.WorldInfoNotFound); ok {
			_ = ctx.Error(errors.NewNotFoundError("world info not found"))
			return
		}
		_ = ctx.Error(err)
		return
	}

	common_handlers.HandleSuccessResponse(ctx, http.StatusOK, &dtos.WorldResponse{
		ID:             updatedWorld.ID.String(),
		UserId:         updatedWorld.UserId.String(),
		Name:           updatedWorld.Name,
		Description:    updatedWorld.Description,
		Data:           updatedWorld.Data.String(),
		CreateableData: updatedWorld.CreateableData.String(),
		CreatedAt:      updatedWorld.CreatedAt,
		UpdatedAt:      updatedWorld.UpdatedAt,
	})
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
func (c *worldController) PublishZone(ctx *gin.Context) {
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

	var req dtos.PublishZoneRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		_ = ctx.Error(errors.NewBadRequestError("invalid JSON payload: " + err.Error()))
		return
	}

	worldInfo, err := c.worldService.GetWorld(worldID)
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

	zoneDataBytes, err := json.Marshal(req.Data)
	if err != nil {
		_ = ctx.Error(errors.NewBadRequestError("failed to marshal zone data: " + err.Error()))
		return
	}

	zone, err := c.worldService.PublishZone(worldID, zoneID, zoneDataBytes)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	common_handlers.HandleSuccessResponse(ctx, http.StatusOK, &dtos.WorldZoneResponse{
		WorldID:  zone.WorldID.String(),
		ZoneID:   zone.ID,
		ZoneData: string(zone.ZoneData),
	})
}
