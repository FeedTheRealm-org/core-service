package character

import (
	"net/http"

	"github.com/FeedTheRealm-org/core-service/config"
	"github.com/FeedTheRealm-org/core-service/internal/common_handlers"
	"github.com/FeedTheRealm-org/core-service/internal/errors"
	"github.com/FeedTheRealm-org/core-service/internal/players-service/dtos"
	player_errors "github.com/FeedTheRealm-org/core-service/internal/players-service/errors"
	"github.com/FeedTheRealm-org/core-service/internal/players-service/models"
	"github.com/FeedTheRealm-org/core-service/internal/players-service/services/character"
	"github.com/FeedTheRealm-org/core-service/internal/utils/input_validation"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type characterController struct {
	conf             *config.Config
	characterService character.CharacterService
}

// NewCharacterController creates a new instance of CharacterController.
func NewCharacterController(conf *config.Config, characterService character.CharacterService) CharacterController {
	return &characterController{
		conf:             conf,
		characterService: characterService,
	}
}

// @Summary PatchCharacterInfo
// @Description Updates the name and bio of the session player character
// @Tags players-service
// @Accept   json
// @Produce  json
// @Param   request body dtos.PatchCharacterInfoRequest true "Character Info data"
// @Success 200  {object}  dtos.CreateAccountResponseDTO "Updated correctly"
// @Failure 400  {object}  dtos.ErrorResponse "Bad request body"
// @Failure 401  {object}  dtos.ErrorResponse "Invalid credentials or invalid JWT token"
// @Router /players/character [put]
func (c *characterController) PatchCharacterInfo(ctx *gin.Context) {
	userId, err := common_handlers.GetUserIDFromSession(ctx)
	if err != nil {
		_ = ctx.Error(errors.NewUnauthorizedError(err.Error()))
		return
	}

	req := &dtos.PatchCharacterInfoRequest{}
	if err := ctx.ShouldBindJSON(req); err != nil {
		_ = ctx.Error(errors.NewBadRequestError(err.Error()))
		return
	}

	// Validations TODO: handle different validations per field (e.g. name cant have spaces but bio can)
	if len(req.CharacterName) < 3 || len(req.CharacterName) > 24 {
		_ = ctx.Error(errors.NewBadRequestError("character name must be between 3 and 24 characters"))
		return
	} else if len(req.CharacterBio) > 256 {
		_ = ctx.Error(errors.NewBadRequestError("character bio must be less than 256 characters"))
		return
	} else if input_validation.ValidateInvalidCharacters(req.CharacterName) || input_validation.HasSpaces(req.CharacterName) {
		_ = ctx.Error(errors.NewBadRequestError("character name contains invalid special characters"))
		return
	} else if input_validation.ValidateInvalidCharacters(req.CharacterBio) {
		_ = ctx.Error(errors.NewBadRequestError("character bio contains invalid special characters"))
		return
	} else if len(req.CategorySprites) == 0 {
		_ = ctx.Error(errors.NewBadRequestError("at least one category sprite must be provided"))
		return
	}

	characterInfo := &models.CharacterInfo{
		CharacterName: req.CharacterName,
		CharacterBio:  req.CharacterBio,
	}
	categorySprites := models.MapToCategorySprites(userId, req.CategorySprites)
	if err := c.characterService.UpdateCharacterInfo(userId, characterInfo, categorySprites); err != nil {
		if _, ok := err.(*player_errors.CharacterNameTaken); ok {
			_ = ctx.Error(errors.NewConflictError("character name is already taken"))
			return
		}
		_ = ctx.Error(err)
		return
	}

	res := &dtos.CharacterInfoResponse{
		CharacterName:   characterInfo.CharacterName,
		CharacterBio:    characterInfo.CharacterBio,
		CategorySprites: req.CategorySprites, // Same as the request, validation should be done already by assets service
		CreatedAt:       characterInfo.CreatedAt,
		UpdatedAt:       characterInfo.UpdatedAt,
	}

	common_handlers.HandleSuccessResponse(ctx, http.StatusOK, res)
}

// @Summary GetCharacterInfo
// @Description Retrieves the name and bio of the session player character
// @Tags players-service
// @Accept   json
// @Produce  json
// @Success 200  {object}  dtos.CharacterInfoResponse "Character info retrieved correctly"
// @Failure 401  {object}  dtos.ErrorResponse "Invalid credentials or invalid JWT token"
// @Router /players/character [get]
func (c *characterController) GetCharacterInfo(ctx *gin.Context) {
	sessionUserId, err := common_handlers.GetUserIDFromSession(ctx)
	if err != nil {
		_ = ctx.Error(errors.NewUnauthorizedError(err.Error()))
		return
	}

	targetUserId := sessionUserId
	userId := ctx.Param("id")
	if userId != "" {
		parsedUserId, err := uuid.Parse(userId)
		if err != nil {
			_ = ctx.Error(errors.NewBadRequestError("invalid user ID: " + userId))
			return
		}
		targetUserId = parsedUserId
	}

	characterInfo, categorySprites, err := c.characterService.GetCharacterInfo(targetUserId)
	if err != nil {
		if _, ok := err.(*player_errors.CharacterInfoNotFound); ok {
			_ = ctx.Error(errors.NewNotFoundError("character info not found"))
			return
		} else if _, ok := err.(*player_errors.CategorySpritesNotFound); ok {
			_ = ctx.Error(errors.NewNotFoundError("character info not found"))
			return
		}
		_ = ctx.Error(err)
		return
	}

	res := &dtos.CharacterInfoResponse{
		CharacterName:   characterInfo.CharacterName,
		CharacterBio:    characterInfo.CharacterBio,
		CategorySprites: models.CategorySpritesToMap(categorySprites),
		CreatedAt:       characterInfo.CreatedAt,
		UpdatedAt:       characterInfo.UpdatedAt,
	}

	common_handlers.HandleSuccessResponse(ctx, http.StatusOK, res)
}
