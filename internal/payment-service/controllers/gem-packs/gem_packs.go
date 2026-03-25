package gem_packs

import (
	"github.com/FeedTheRealm-org/core-service/config"
	"github.com/FeedTheRealm-org/core-service/internal/common_handlers"
	"github.com/FeedTheRealm-org/core-service/internal/errors"
	"github.com/FeedTheRealm-org/core-service/internal/payment-service/dtos"
	gem_packs "github.com/FeedTheRealm-org/core-service/internal/payment-service/services/gem-packs"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type gemPacksController struct {
	conf            *config.Config
	gemPacksService gem_packs.GemPacksService
}

func NewGemPacksController(conf *config.Config, gemPacksService gem_packs.GemPacksService) GemPacksController {
	return &gemPacksController{
		conf:            conf,
		gemPacksService: gemPacksService,
	}
}

func (pc *gemPacksController) GetAllGemPacks(c *gin.Context) {
	_, err := common_handlers.GetUserIDFromSession(c)
	if err != nil {
		_ = c.Error(errors.NewUnauthorizedError(err.Error()))
		return
	}

	packs, err := pc.gemPacksService.GetAllGemPacks()
	if err != nil {
		_ = c.Error(err)
		return
	}

	common_handlers.HandleSuccessResponse(c, 200, packs)
}

func (pc *gemPacksController) GetGemPackById(c *gin.Context) {
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

	pack, err := pc.gemPacksService.GetGemPackById(packId)
	if err != nil {
		_ = c.Error(err)
		return
	}

	common_handlers.HandleSuccessResponse(c, 200, pack)
}

func (pc *gemPacksController) CreateGemPack(c *gin.Context) {
	pack := &dtos.CreateGemPackRequest{}
	if err := c.ShouldBindJSON(pack); err != nil {
		_ = c.Error(errors.NewBadRequestError("invalid request body: " + err.Error()))
		return
	}

	createdPack, err := pc.gemPacksService.CreateGemPack(pack.Name, pack.Gems, pack.Price)
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

func (pc *gemPacksController) UpdateGemPack(c *gin.Context) {
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

	updatedPack, err := pc.gemPacksService.UpdateGemPack(packId, req.Name, req.Gems, req.Price)
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

func (pc *gemPacksController) DeleteGemPack(c *gin.Context) {
	packId, err := uuid.Parse(c.Param("id"))
	if err != nil {
		_ = c.Error(errors.NewBadRequestError("invalid pack_id: " + err.Error()))
		return
	}

	if err := pc.gemPacksService.DeleteGemPack(packId); err != nil {
		_ = c.Error(err)
		return
	}

	res := &dtos.GemPackDeletedResponse{
		Id: packId,
	}

	common_handlers.HandleSuccessResponse(c, 200, res)
}
