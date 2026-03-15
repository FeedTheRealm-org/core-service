package packs

import (
	"github.com/FeedTheRealm-org/core-service/config"
	"github.com/FeedTheRealm-org/core-service/internal/common_handlers"
	"github.com/FeedTheRealm-org/core-service/internal/errors"
	"github.com/FeedTheRealm-org/core-service/internal/payment-service/dtos"
	"github.com/FeedTheRealm-org/core-service/internal/payment-service/services/packs"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type packsController struct {
	conf        *config.Config
	packService packs.PacksService
}

func NewPacksController(conf *config.Config, packService packs.PacksService) PacksController {
	return &packsController{
		conf:        conf,
		packService: packService,
	}
}

func (pc *packsController) GetAllPacks(c *gin.Context) {
	_, err := common_handlers.GetUserIDFromSession(c)
	if err != nil {
		_ = c.Error(errors.NewUnauthorizedError(err.Error()))
		return
	}

	packs, err := pc.packService.GetAllPacks()
	if err != nil {
		_ = c.Error(err)
		return
	}

	common_handlers.HandleSuccessResponse(c, 200, packs)
}

func (pc *packsController) GetPackById(c *gin.Context) {
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

	pack, err := pc.packService.GetPackById(packId)
	if err != nil {
		_ = c.Error(err)
		return
	}

	common_handlers.HandleSuccessResponse(c, 200, pack)
}

func (pc *packsController) CreatePack(c *gin.Context) {
	_, err := common_handlers.GetUserIDFromSession(c)
	if err != nil {
		_ = c.Error(errors.NewUnauthorizedError(err.Error()))
		return
	}

	if err := common_handlers.IsAdminSession(c); err != nil {
		_ = c.Error(errors.NewUnauthorizedError(err.Error()))
		return
	}

	pack := &dtos.CreatePackRequest{}
	if err := c.ShouldBindJSON(pack); err != nil {
		_ = c.Error(errors.NewBadRequestError("invalid request body: " + err.Error()))
		return
	}

	createdPack, err := pc.packService.CreatePack(pack.Name, pack.Gems, pack.Price)
	if err != nil {
		_ = c.Error(err)
		return
	}

	res := &dtos.PackResponse{
		Id:        createdPack.Id,
		Name:      createdPack.Name,
		Gems:      createdPack.Gems,
		Price:     createdPack.Price,
		CreatedAt: createdPack.CreatedAt,
		UpdatedAt: createdPack.UpdatedAt,
	}

	common_handlers.HandleSuccessResponse(c, 201, res)
}

func (pc *packsController) UpdatePack(c *gin.Context) {
	_, err := common_handlers.GetUserIDFromSession(c)
	if err != nil {
		_ = c.Error(errors.NewUnauthorizedError(err.Error()))
		return
	}

	if err := common_handlers.IsAdminSession(c); err != nil {
		_ = c.Error(errors.NewUnauthorizedError(err.Error()))
		return
	}

	req := &dtos.UpdatePackRequest{}
	if err := c.ShouldBindJSON(req); err != nil {
		_ = c.Error(errors.NewBadRequestError("invalid request body: " + err.Error()))
		return
	}

	packId, err := uuid.Parse(c.Param("id"))
	if err != nil {
		_ = c.Error(errors.NewBadRequestError("invalid pack_id: " + err.Error()))
		return
	}

	updatedPack, err := pc.packService.UpdatePack(packId, req.Name, req.Gems, req.Price)
	if err != nil {
		_ = c.Error(err)
		return
	}

	res := &dtos.PackResponse{
		Id:        updatedPack.Id,
		Name:      updatedPack.Name,
		Gems:      updatedPack.Gems,
		Price:     updatedPack.Price,
		CreatedAt: updatedPack.CreatedAt,
		UpdatedAt: updatedPack.UpdatedAt,
	}

	common_handlers.HandleSuccessResponse(c, 200, res)
}

func (pc *packsController) DeletePack(c *gin.Context) {
	_, err := common_handlers.GetUserIDFromSession(c)
	if err != nil {
		_ = c.Error(errors.NewUnauthorizedError(err.Error()))
		return
	}

	if err := common_handlers.IsAdminSession(c); err != nil {
		_ = c.Error(errors.NewUnauthorizedError(err.Error()))
		return
	}

	packId, err := uuid.Parse(c.Param("id"))
	if err != nil {
		_ = c.Error(errors.NewBadRequestError("invalid pack_id: " + err.Error()))
		return
	}

	if err := pc.packService.DeletePack(packId); err != nil {
		_ = c.Error(err)
		return
	}

	res := &dtos.PackDeletedResponse{
		Id: packId,
	}

	common_handlers.HandleSuccessResponse(c, 200, res)
}
