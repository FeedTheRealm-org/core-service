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

// @Summary PublishWorld
// @Description Publishes a new world with the provided information
// @Tags world-service
// @Accept   json
// @Produce  json
// @Param   request body dtos.WorldRequest true "World Data"
// @Success 200  {object}  dtos.WorldResponse "Published correctly"
// @Failure 400  {object}  dtos.ErrorResponse "Bad request body"
// @Failure 401  {object}  dtos.ErrorResponse "Invalid credentials or invalid JWT token"
// @Router /world [post]
func (c *worldController) PublishWorld(ctx *gin.Context) {
	// userId, err := common_handlers.GetUserIDFromSession(ctx)
	// if err != nil {
	// 	_ = ctx.Error(errors.NewUnauthorizedError(err.Error()))
	// 	return
	// }

	var req dtos.WorldRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		_ = ctx.Error(errors.NewBadRequestError("invalid JSON payload: " + err.Error()))
		return
	}

	if len(req.FileName) < 3 || len(req.FileName) > 24 {
		_ = ctx.Error(errors.NewBadRequestError("world name must be between 3 and 24 characters"))
		return
	}

	if input_validation.ValidateInvalidCharacters(req.FileName) || input_validation.HasSpaces(req.FileName) {
		_ = ctx.Error(errors.NewBadRequestError("world name contains invalid special characters"))
		return
	}

	bytes, _ := json.Marshal(req.Data)

	worldData := &models.WorldData{
		UserId: uuid.New(), // this is a temp for testing purposes
		Name:   req.FileName,
		Data:   datatypes.JSON(bytes),
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
		UserId:    createdWorld.UserId.String(),
		Name:      createdWorld.Name,
		Data:      string(createdWorld.Data),
		CreatedAt: createdWorld.CreatedAt,
		UpdatedAt: createdWorld.UpdatedAt,
	}

	common_handlers.HandleSuccessResponse(ctx, http.StatusCreated, response)
}

// @Summary GetWorld
// @Description Retrieves the name and data of the session player world
// @Tags world-service
// @Accept   json
// @Produce  json
// @Success 200  {object}  dtos.WorldResponse "World info retrieved correctly"
// @Failure 401  {object}  dtos.ErrorResponse "Invalid credentials or invalid JWT token"
// @Router /world/:id [get]
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
		UserId:    worldInfo.UserId.String(),
		Name:      worldInfo.Name,
		Data:      string(worldInfo.Data),
		CreatedAt: worldInfo.CreatedAt,
		UpdatedAt: worldInfo.UpdatedAt,
	}

	common_handlers.HandleSuccessResponse(ctx, http.StatusOK, res)
}

// @Summary GetWorldsList
// @Description Retrieves a paginated list of worlds
// @Tags world-service
// @Accept   json
// @Produce  json
// @Success 200  {object}  dtos.WorldResponse "Worlds list retrieved correctly"
// @Failure 401  {object}  dtos.ErrorResponse "Invalid credentials or invalid JWT token"
func (c *worldController) GetWorldsList(ctx *gin.Context) {
	_, err := common_handlers.GetUserIDFromSession(ctx)
	if err != nil {
		_ = ctx.Error(errors.NewUnauthorizedError(err.Error()))
		return
	}

	offsetStr := ctx.Query("offset")
	limitStr := ctx.Query("limit")
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

	worldsList, err := c.worldService.GetWorldsList(offset, limit)
	if err != nil {
		if _, ok := err.(*world_errors.WorldInfoNotFound); ok {
			_ = ctx.Error(errors.NewNotFoundError("world info not found"))
			return
		}
		_ = ctx.Error(err)
		return
	}

	resList := make([]dtos.WorldResponse, 0, len(worldsList))
	for _, worldInfo := range worldsList {
		resList = append(resList, dtos.WorldResponse{
			UserId:    worldInfo.UserId.String(),
			Name:      worldInfo.Name,
			Data:      string(worldInfo.Data),
			CreatedAt: worldInfo.CreatedAt,
			UpdatedAt: worldInfo.UpdatedAt,
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
