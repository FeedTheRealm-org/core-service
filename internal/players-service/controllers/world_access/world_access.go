package world_access

import (
	"net/http"
	"strings"

	"github.com/FeedTheRealm-org/core-service/config"
	"github.com/FeedTheRealm-org/core-service/internal/common_handlers"
	"github.com/FeedTheRealm-org/core-service/internal/errors"
	"github.com/FeedTheRealm-org/core-service/internal/players-service/dtos"
	player_errors "github.com/FeedTheRealm-org/core-service/internal/players-service/errors"
	"github.com/FeedTheRealm-org/core-service/internal/players-service/services/world_access"
	"github.com/gin-gonic/gin"
)

type worldAccessController struct {
	conf               *config.Config
	worldAccessService world_access.WorldAccessService
}

func NewWorldAccessController(conf *config.Config, worldAccessService world_access.WorldAccessService) WorldAccessController {
	return &worldAccessController{conf: conf, worldAccessService: worldAccessService}
}

// IssueWorldJoinToken godoc
// @Summary      Issue one-time world join token
// @Description  Creates a short-lived token tied to the current authenticated user and target world.
// @Tags         players-service
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        request body dtos.IssueWorldJoinTokenRequest true "Issue world join token DTO"
// @Success      200  {object}  dtos.IssueWorldJoinTokenResponse
// @Failure      400  {object} dtos.ErrorResponse
// @Failure      401  {object} dtos.ErrorResponse
// @Router       /player/world-access/token [post]
func (c *worldAccessController) IssueWorldJoinToken(ctx *gin.Context) {
	userId, err := common_handlers.GetUserIDFromSession(ctx)
	if err != nil {
		_ = ctx.Error(errors.NewUnauthorizedError(err.Error()))
		return
	}

	req := &dtos.IssueWorldJoinTokenRequest{}
	if err := ctx.ShouldBindJSON(req); err != nil {
		_ = ctx.Error(errors.NewBadRequestError(err.Error()))
		return
	}

	req.WorldId = strings.TrimSpace(req.WorldId)
	if req.WorldId == "" {
		_ = ctx.Error(errors.NewBadRequestError("world_id is required"))
		return
	}

	token, err := c.worldAccessService.IssueWorldJoinToken(userId, req.WorldId)
	if err != nil {
		if _, ok := err.(*player_errors.WorldJoinTokenInvalid); ok {
			_ = ctx.Error(errors.NewBadRequestError(err.Error()))
			return
		}
		_ = ctx.Error(err)
		return
	}

	res := &dtos.IssueWorldJoinTokenResponse{
		TokenId:   token.TokenId,
		ExpiresAt: token.ExpiresAt,
	}
	common_handlers.HandleSuccessResponse(ctx, http.StatusOK, res)
}

// ConsumeWorldJoinToken godoc
// @Summary      Consume one-time world join token
// @Description  Validates and burns a one-time token, returning the associated user ID.
// @Tags         players-service
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        request body dtos.ConsumeWorldJoinTokenRequest true "Consume world join token DTO"
// @Success      200  {object}  dtos.ConsumeWorldJoinTokenResponse
// @Failure      400  {object} dtos.ErrorResponse
// @Failure      401  {object} dtos.ErrorResponse
// @Failure      404  {object} dtos.ErrorResponse
// @Router       /player/world-access/token/consume [post]
func (c *worldAccessController) ConsumeWorldJoinToken(ctx *gin.Context) {
	if err := common_handlers.IsSessionValid(ctx); err != nil {
		_ = ctx.Error(errors.NewUnauthorizedError(err.Error()))
		return
	}

	req := &dtos.ConsumeWorldJoinTokenRequest{}
	if err := ctx.ShouldBindJSON(req); err != nil {
		_ = ctx.Error(errors.NewBadRequestError(err.Error()))
		return
	}

	token, err := c.worldAccessService.ConsumeWorldJoinToken(req.TokenId)
	if err != nil {
		switch err.(type) {
		case *player_errors.WorldJoinTokenInvalid:
			_ = ctx.Error(errors.NewBadRequestError(err.Error()))
			return
		case *player_errors.WorldJoinTokenExpired:
			_ = ctx.Error(errors.NewBadRequestError(err.Error()))
			return
		case *player_errors.WorldJoinTokenConsumed:
			_ = ctx.Error(errors.NewBadRequestError(err.Error()))
			return
		case *player_errors.WorldJoinTokenNotFound:
			_ = ctx.Error(errors.NewNotFoundError(err.Error()))
			return
		default:
			_ = ctx.Error(err)
			return
		}
	}

	res := &dtos.ConsumeWorldJoinTokenResponse{
		UserId:  token.UserId,
		WorldId: token.WorldId,
	}
	common_handlers.HandleSuccessResponse(ctx, http.StatusOK, res)
}
