package router

import (
	"github.com/FeedTheRealm-org/core-service/config"
	"github.com/FeedTheRealm-org/core-service/internal/middleware"
	server_registry_controller "github.com/FeedTheRealm-org/core-service/internal/world-service/controllers/server_registry"
	world_controller "github.com/FeedTheRealm-org/core-service/internal/world-service/controllers/world"
	world_repo "github.com/FeedTheRealm-org/core-service/internal/world-service/repositories/world"
	server_registry_service "github.com/FeedTheRealm-org/core-service/internal/world-service/services/server_registry"
	world_service "github.com/FeedTheRealm-org/core-service/internal/world-service/services/world"
	"github.com/gin-gonic/gin"
)

func SetupWorldServiceRouter(r *gin.Engine, conf *config.Config, db *config.DB) error {
	worldGroup := r.Group("/world")

	worldRepo := world_repo.NewWorldRepository(conf, db)

	var nomadService server_registry_service.ServerRegistryService
	if conf.Server.Environment == config.Production {
		var err error
		nomadService, err = server_registry_service.NewServerRegistryService(conf) // Real nomad service
		if err != nil {
			return err
		}
	} else {
		nomadService = server_registry_service.NewStubServerRegistryService() // Stub
	}
	worldService := world_service.NewWorldService(conf, worldRepo, nomadService)
	worldController := world_controller.NewWorldController(conf, worldService)
	serverRegistryController := server_registry_controller.NewServerRegistryController(conf, nomadService)

	worldGroup.POST("", worldController.PublishWorld)
	worldGroup.PUT("/:id/zones/:zone_id", worldController.PublishZone)
	worldGroup.GET("", worldController.GetWorldsList)
	worldGroup.PUT("/:id", worldController.UpdateWorld)
	worldGroup.PUT("/:id/createable-data", worldController.UpdateCreateableData)
	worldGroup.GET("/:id/zones", worldController.GetWorldZones)
	worldGroup.GET("/:id/zones/:zone_id", worldController.GetWorldZoneData)
	worldGroup.GET("/:id", worldController.GetWorld)
	worldGroup.DELETE("/:id", worldController.DeleteWorld)
	worldGroup.DELETE("/reset-database", middleware.AdminCheckMiddleware(), worldController.ResetDatabase)

	orchestratorGroup := worldGroup.Group("/orchestrator")
	orchestratorGroup.GET("/:id/zones/:zone_id/start-job", middleware.AdminCheckMiddleware(), serverRegistryController.StartNewJob)
	orchestratorGroup.GET("/:id/zones/:zone_id/stop-job", middleware.AdminCheckMiddleware(), serverRegistryController.StopJob)
	orchestratorGroup.GET("/:id/zones/:zone_id/address", serverRegistryController.GetServerAddress)

	return nil
}
