package router

import (
	"github.com/FeedTheRealm-org/core-service/config"
	character_controller "github.com/FeedTheRealm-org/core-service/internal/players-service/controllers/character"
	character_repo "github.com/FeedTheRealm-org/core-service/internal/players-service/repositories/character"
	character_service "github.com/FeedTheRealm-org/core-service/internal/players-service/services/character"
	"github.com/gin-gonic/gin"
)

func SetupPlayerServiceRouter(r *gin.Engine, conf *config.Config, db *config.DB) {
	g := r.Group("/player")

	characterRepo := character_repo.NewCharacterRepository(conf, db)
	characterService := character_service.NewCharacterService(conf, characterRepo)
	characterController := character_controller.NewCharacterController(conf, characterService)

	characterGroup := g.Group("/character")
	characterGroup.PUT("", characterController.UpdateCharacterInfo)
	characterGroup.GET("", characterController.GetCharacterInfo)
	characterGroup.GET(":id", characterController.GetCharacterInfo)
}
