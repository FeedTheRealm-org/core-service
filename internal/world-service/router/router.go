package router

import (
	"os"

	"github.com/FeedTheRealm-org/core-service/config"
	world_controller "github.com/FeedTheRealm-org/core-service/internal/world-service/controllers/world"
	world_repo "github.com/FeedTheRealm-org/core-service/internal/world-service/repositories/world"
	world_service "github.com/FeedTheRealm-org/core-service/internal/world-service/services/world"
	"github.com/gin-gonic/gin"
)

func SetupWorldServiceRouter(r *gin.Engine, conf *config.Config, db *config.DB) {
	worldGroup := r.Group("/world")

	worldRepo := world_repo.NewWorldRepository(conf, db)
	worldService := world_service.NewWorldService(conf, worldRepo)
	worldController := world_controller.NewWorldController(conf, worldService)

	worldGroup.POST("", worldController.PublishWorld)
	worldGroup.GET("", worldController.GetWorldsList)
	worldGroup.PUT(":id", worldController.UpdateWorld)
	worldGroup.GET(":id", worldController.GetWorld)
	if os.Getenv("ALLOW_DB_RESET") == "true" {
		worldGroup.DELETE("/reset-database", worldController.ResetDatabase)
	}
}
