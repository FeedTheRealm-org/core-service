package router

import (
	"os"

	"github.com/FeedTheRealm-org/core-service/config"
	world_controller "github.com/FeedTheRealm-org/core-service/internal/world-service/controllers/world"
	world_repo "github.com/FeedTheRealm-org/core-service/internal/world-service/repositories/world"
	nomad_job_sender "github.com/FeedTheRealm-org/core-service/internal/world-service/services/nomad_job_sender"
	world_service "github.com/FeedTheRealm-org/core-service/internal/world-service/services/world"
	"github.com/gin-gonic/gin"
)

func SetupWorldServiceRouter(r *gin.Engine, conf *config.Config, db *config.DB) error {
	worldGroup := r.Group("/world")

	worldRepo := world_repo.NewWorldRepository(conf, db)

	var nomadService nomad_job_sender.NomadJobSenderService
	if conf.Server.Environment == config.Production {
		nomadService = nomad_job_sender.NewNomadJobSenderService(conf) // Real nomad service
	} else {
		nomadService = nomad_job_sender.NewStubNomadJobSenderService() // Stub
	}
	worldService := world_service.NewWorldService(conf, worldRepo, nomadService)
	worldController := world_controller.NewWorldController(conf, worldService)

	worldGroup.POST("", worldController.PublishWorld)
	worldGroup.GET("", worldController.GetWorldsList)
	worldGroup.PUT("/:id", worldController.UpdateWorld)
	worldGroup.GET("/:id", worldController.GetWorld)
	if os.Getenv("ALLOW_DB_RESET") == "true" {
		worldGroup.DELETE("/reset-database", worldController.ResetDatabase)
	}

	return nil
}
