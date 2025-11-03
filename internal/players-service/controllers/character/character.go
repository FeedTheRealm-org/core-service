package character

import (
	"github.com/FeedTheRealm-org/core-service/internal/players-service/services/character"
	"github.com/gin-gonic/gin"
)

type characterController struct {
	characterService character.CharacterService
}

// NewCharacterController creates a new instance of CharacterController.
func NewCharacterController(characterService character.CharacterService) CharacterController {
	return &characterController{
		characterService: characterService,
	}
}

func (c *characterController) UpdateCharacterInfo(ctx *gin.Context) {}

func (c *characterController) GetCharacterInfo(ctx *gin.Context) {}
