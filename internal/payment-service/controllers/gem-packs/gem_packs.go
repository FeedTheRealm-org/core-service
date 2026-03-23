package packs

import (
	"github.com/FeedTheRealm-org/core-service/config"
	"github.com/FeedTheRealm-org/core-service/internal/common_handlers"
	"github.com/FeedTheRealm-org/core-service/internal/errors"
	"github.com/FeedTheRealm-org/core-service/internal/payment-service/dtos"
	gem_packs "github.com/FeedTheRealm-org/core-service/internal/payment-service/services/gem-packs"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type gemGemPacksController struct {
	conf               *config.Config
	gemGemPacksService gem_packs.GemPacksService
}

func NewGemPacksController(conf *config.Config, gemGemPacksService gem_packs.GemPacksService) GemPacksController {
	return &gemGemPacksController{
		conf:               conf,
		gemGemPacksService: gemGemPacksService,
	}
}

func (pc *gemGemPacksController) GetAllGemPacks(c *gin.Context) {
	_, err := common_handlers.GetUserIDFromSession(c)
	if err != nil {
		_ = c.Error(errors.NewUnauthorizedError(err.Error()))
		return
	}

	packs, err := pc.gemGemPacksService.GetAllGemPacks()
	if err != nil {
		_ = c.Error(err)
		return
	}

	common_handlers.HandleSuccessResponse(c, 200, packs)
}

func (pc *gemGemPacksController) GetGemPackById(c *gin.Context) {
	_, err := common_handlers.GetUserIDFromSession(c)
	if err != nil {
		_ = c.Error(errors.NewUnauthorizedError(err.Error()))
		return
	}

	packId, err := uuid.Parse(c.Param("id"))
	if err != nil {
		_ = c.Error(errors.NewBadRequestError("invalid pack_id: " + err.Error()))
		return
	}

	pack, err := pc.gemGemPacksService.GetGemPackById(packId)
	if err != nil {
		_ = c.Error(err)
		return
	}

	common_handlers.HandleSuccessResponse(c, 200, pack)
}

func (pc *gemGemPacksController) CreateGemPack(c *gin.Context) {
	pack := &dtos.CreateGemPackRequest{}
	if err := c.ShouldBindJSON(pack); err != nil {
		_ = c.Error(errors.NewBadRequestError("invalid request body: " + err.Error()))
		return
	}

	createdPack, err := pc.gemGemPacksService.CreateGemPack(pack.Name, pack.Gems, pack.Price)
	if err != nil {
		_ = c.Error(err)
		return
	}

	res := &dtos.GemPackResponse{
		Id:        createdPack.Id,
		Name:      createdPack.Name,
		Gems:      createdPack.Gems,
		Price:     createdPack.Price,
		CreatedAt: createdPack.CreatedAt,
		UpdatedAt: createdPack.UpdatedAt,
	}

	common_handlers.HandleSuccessResponse(c, 201, res)
}

func (pc *gemGemPacksController) UpdateGemPack(c *gin.Context) {
	req := &dtos.UpdateGemPackRequest{}
	if err := c.ShouldBindJSON(req); err != nil {
		_ = c.Error(errors.NewBadRequestError("invalid request body: " + err.Error()))
		return
	}

	packId, err := uuid.Parse(c.Param("id"))
	if err != nil {
		_ = c.Error(errors.NewBadRequestError("invalid pack_id: " + err.Error()))
		return
	}

	updatedPack, err := pc.gemGemPacksService.UpdateGemPack(packId, req.Name, req.Gems, req.Price)
	if err != nil {
		_ = c.Error(err)
		return
	}

	res := &dtos.GemPackResponse{
		Id:        updatedPack.Id,
		Name:      updatedPack.Name,
		Gems:      updatedPack.Gems,
		Price:     updatedPack.Price,
		CreatedAt: updatedPack.CreatedAt,
		UpdatedAt: updatedPack.UpdatedAt,
	}

	common_handlers.HandleSuccessResponse(c, 200, res)
}

func (pc *gemGemPacksController) DeleteGemPack(c *gin.Context) {
	packId, err := uuid.Parse(c.Param("id"))
	if err != nil {
		_ = c.Error(errors.NewBadRequestError("invalid pack_id: " + err.Error()))
		return
	}

	if err := pc.gemGemPacksService.DeleteGemPack(packId); err != nil {
		_ = c.Error(err)
		return
	}

	res := &dtos.GemPackDeletedResponse{
		Id: packId,
	}

	common_handlers.HandleSuccessResponse(c, 200, res)
}
