package sprites

import (
	"github.com/FeedTheRealm-org/core-service/config"
	"github.com/FeedTheRealm-org/core-service/internal/assets-service/services/sprites"
	"github.com/gin-gonic/gin"
)

type spritesController struct {
	conf           *config.Config
	spritesService sprites.SpritesService
}

// NewSpritesController creates a new instance of CharacterController.
func NewSpritesController(conf *config.Config, spritesService sprites.SpritesService) SpritesController {
	return &spritesController{
		conf:           conf,
		spritesService: spritesService,
	}
}

// @Summary GetCategoriesList
// @Description Retrieves the list of existing categories UUIDs.
// @Tags assets-service
// @Produce  json
// @Success 200  {object}  dtos.SpriteCategoryListResponse "Category list"
// @Failure 401  {object}  dtos.ErrorResponse "Invalid credentials or invalid JWT token"
// @Router /assets/sprites/categories [get]
func (sc *spritesController) GetCategoriesList(c *gin.Context) {
	c.JSON(200, gin.H{"message": "GetCategoriesList not implemented"})
}

// @Summary GetSpritesListByCategory
// @Description Retrieves the list of existing sprites UUIDs for a given category UUID.
// @Tags assets-service
// @Produce  json
// @Param category_id path string true "Category UUID"
// @Success 200  {object}  dtos.SpritesListResponse "Sprite list"
// @Failure 401  {object}  dtos.ErrorResponse "Invalid credentials or invalid JWT token"
// @Router /assets/sprites/categories/{category_id} [get]
func (sc *spritesController) GetSpritesListByCategory(c *gin.Context) {
	c.JSON(200, gin.H{"message": "GetSpritesListByCategory not implemented"})
}

// @Summary DownloadSpriteData
// @Description Downloads the sprite file for a given sprite UUID.
// @Tags assets-service
// @Produce  octet-stream
// @Param sprite_id path string true "Sprite UUID"
// @Success 200  {file}  byte "Sprite file"
// @Failure 401  {object}  dtos.ErrorResponse "Invalid credentials or invalid JWT token"
// @Router /assets/sprites/{sprite_id} [get]
func (sc *spritesController) DownloadSpriteData(c *gin.Context) {
	c.JSON(200, gin.H{"message": "DownloadSpriteData not implemented"})
}

// @Summary AddCategory
// @Description Adds a new sprite category.
// @Tags assets-service
// @Accept  json
// @Produce  json
// @Param category body dtos.AddSpriteCategoryRequest true "Category data"
// @Success 201  {object}  dtos.SpriteCategoryResponse "Created category"
// @Failure 401  {object}  dtos.ErrorResponse "Invalid credentials or invalid JWT token"
// @Router /assets/sprites/categories [post]
func (sc *spritesController) AddCategory(c *gin.Context) {
	c.JSON(200, gin.H{"message": "AddCategory not implemented"})
}

// @Summary UploadSpriteData
// @Description Uploads a sprite file.
// @Tags assets-service
// @Accept  multipart/form-data
// @Produce  json
// @Param sprite formData file true "Sprite file"
// @Success 201  {object}  dtos.SpriteResponse "Uploaded sprite"
// @Failure 401  {object}  dtos.ErrorResponse "Invalid credentials or invalid JWT token"
// @Router /assets/sprites [put]
func (sc *spritesController) UploadSpriteData(c *gin.Context) {
	c.JSON(200, gin.H{"message": "UploadSpriteData not implemented"})
}
