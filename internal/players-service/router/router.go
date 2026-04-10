package router

import (
	"github.com/FeedTheRealm-org/core-service/config"
	character_controller "github.com/FeedTheRealm-org/core-service/internal/players-service/controllers/character"
	world_access_controller "github.com/FeedTheRealm-org/core-service/internal/players-service/controllers/world_access"
	character_repo "github.com/FeedTheRealm-org/core-service/internal/players-service/repositories/character"
	world_access_repo "github.com/FeedTheRealm-org/core-service/internal/players-service/repositories/world_access"
	character_service "github.com/FeedTheRealm-org/core-service/internal/players-service/services/character"
	world_access_service "github.com/FeedTheRealm-org/core-service/internal/players-service/services/world_access"
	"github.com/gin-gonic/gin"
)

func SetupPlayerServiceRouter(r *gin.Engine, conf *config.Config, db *config.DB) error {
	g := r.Group("/player")

	characterRepo := character_repo.NewCharacterRepository(conf, db)
	characterService := character_service.NewCharacterService(conf, characterRepo)
	characterController := character_controller.NewCharacterController(conf, characterService)
	worldAccessRepo := world_access_repo.NewWorldAccessRepository(conf, db)
	worldAccessService := world_access_service.NewWorldAccessService(conf, worldAccessRepo, characterRepo)
	worldAccessController := world_access_controller.NewWorldAccessController(conf, worldAccessService)

	characterGroup := g.Group("/character")
	characterGroup.PATCH("", characterController.PatchCharacterInfo)
	characterGroup.GET("", characterController.GetCharacterInfo)
	characterGroup.GET(":id", characterController.GetCharacterInfo)

	worldAccessGroup := g.Group("/world-access")
	worldAccessGroup.POST("/token", worldAccessController.IssueWorldJoinToken)
	worldAccessGroup.POST("/token/consume", worldAccessController.ConsumeWorldJoinToken)

	return nil
}
