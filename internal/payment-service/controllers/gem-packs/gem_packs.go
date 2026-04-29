package gem_packs

import (
	"github.com/FeedTheRealm-org/core-service/config"
	"github.com/FeedTheRealm-org/core-service/internal/common_handlers"
	"github.com/FeedTheRealm-org/core-service/internal/errors"
	"github.com/FeedTheRealm-org/core-service/internal/payment-service/dtos"
	gem_packs "github.com/FeedTheRealm-org/core-service/internal/payment-service/services/gem-packs"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
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

// GetAllGemPacks godoc
// @Summary      List gem packs
// @Description  Returns all available gem packs.
// @Tags         payment-service
// @Security     BearerAuth
// @Produce      json
// @Success      200  {object}  dtos.GemPackResponse
// @Failure      401  {object}  dtos.ErrorResponse
// @Failure      500  {object}  dtos.ErrorResponse
// @Router       /payments/gems/packs [get]
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

	responses := make([]*dtos.GemPackResponse, 0, len(packs))

	for _, pack := range packs {
		price, _ := pack.Price.Float64()

		responses = append(responses, &dtos.GemPackResponse{
			Id:        pack.Id,
			Name:      pack.Name,
			Gems:      pack.Gems,
			Price:     price,
			CreatedAt: pack.CreatedAt,
			UpdatedAt: pack.UpdatedAt,
		})
	}

	common_handlers.HandleSuccessResponse(c, 200, responses)
}

// GetGemPackById godoc
// @Summary      Get gem pack by ID
// @Description  Returns details for a specific gem pack.
// @Tags         payment-service
// @Security     BearerAuth
// @Produce      json
// @Param        id    path      string               true  "Pack ID"
// @Success      200   {object}  dtos.GemPackResponse
// @Failure      400   {object}  dtos.ErrorResponse
// @Failure      401   {object}  dtos.ErrorResponse
// @Failure      404   {object}  dtos.ErrorResponse
// @Failure      500   {object}  dtos.ErrorResponse
// @Router       /payments/gems/packs/{id} [get]
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

	price, _ := pack.Price.Float64()

	res := &dtos.GemPackResponse{
		Id:        pack.Id,
		Name:      pack.Name,
		Gems:      pack.Gems,
		Price:     price,
		CreatedAt: pack.CreatedAt,
		UpdatedAt: pack.UpdatedAt,
	}

	common_handlers.HandleSuccessResponse(c, 200, res)
}

// CreateGemPack godoc
// @Summary      Create gem pack
// @Description  Creates a new gem pack. Admin only.
// @Tags         payment-service
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        request  body      dtos.CreateGemPackRequest  true  "Create gem pack payload"
// @Success      201      {object}  dtos.GemPackResponse
// @Failure      400      {object}  dtos.ErrorResponse
// @Failure      401      {object}  dtos.ErrorResponse
// @Failure      500      {object}  dtos.ErrorResponse
// @Router       /payments/gems/packs [post]
func (pc *gemPacksController) CreateGemPack(c *gin.Context) {
	pack := &dtos.CreateGemPackRequest{}
	if err := c.ShouldBindJSON(pack); err != nil {
		_ = c.Error(errors.NewBadRequestError("invalid request body: " + err.Error()))
		return
	}

	createdPack, err := pc.gemPacksService.CreateGemPack(pack.Name, pack.Gems, decimal.NewFromFloat(pack.Price))
	if err != nil {
		_ = c.Error(err)
		return
	}

	price, _ := createdPack.Price.Float64()

	res := &dtos.GemPackResponse{
		Id:        createdPack.Id,
		Name:      createdPack.Name,
		Gems:      createdPack.Gems,
		Price:     price,
		CreatedAt: createdPack.CreatedAt,
		UpdatedAt: createdPack.UpdatedAt,
	}

	common_handlers.HandleSuccessResponse(c, 201, res)
}

// UpdateGemPack godoc
// @Summary      Update gem pack
// @Description  Updates an existing gem pack. Admin only.
// @Tags         payment-service
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id       path      string                   true  "Pack ID"
// @Param        request  body      dtos.UpdateGemPackRequest true  "Update gem pack payload"
// @Success      200      {object}  dtos.GemPackResponse
// @Failure      400      {object}  dtos.ErrorResponse
// @Failure      401      {object}  dtos.ErrorResponse
// @Failure      404      {object}  dtos.ErrorResponse
// @Failure      500      {object}  dtos.ErrorResponse
// @Router       /payments/gems/packs/{id} [put]
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

	updatedPack, err := pc.gemPacksService.UpdateGemPack(packId, req.Name, req.Gems, decimal.NewFromFloat(req.Price))
	if err != nil {
		_ = c.Error(err)
		return
	}

	price, _ := updatedPack.Price.Float64()

	res := &dtos.GemPackResponse{
		Id:        updatedPack.Id,
		Name:      updatedPack.Name,
		Gems:      updatedPack.Gems,
		Price:     price,
		CreatedAt: updatedPack.CreatedAt,
		UpdatedAt: updatedPack.UpdatedAt,
	}

	common_handlers.HandleSuccessResponse(c, 200, res)
}

// DeleteGemPack godoc
// @Summary      Delete gem pack
// @Description  Deletes a gem pack by ID. Admin only.
// @Tags         payment-service
// @Security     BearerAuth
// @Produce      json
// @Param        id    path      string                     true  "Pack ID"
// @Success      200   {object}  dtos.GemPackDeletedResponse
// @Failure      400   {object}  dtos.ErrorResponse
// @Failure      401   {object}  dtos.ErrorResponse
// @Failure      404   {object}  dtos.ErrorResponse
// @Failure      500   {object}  dtos.ErrorResponse
// @Router       /payments/gems/packs/{id} [delete]
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
