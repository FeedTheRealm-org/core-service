package router

import (
	"github.com/FeedTheRealm-org/core-service/config"
	"github.com/FeedTheRealm-org/core-service/internal/middleware"
	server_registry_controller "github.com/FeedTheRealm-org/core-service/internal/world-service/controllers/server_registry"
	world_controller "github.com/FeedTheRealm-org/core-service/internal/world-service/controllers/world"
	zones_controller "github.com/FeedTheRealm-org/core-service/internal/world-service/controllers/zones"
	world_repo "github.com/FeedTheRealm-org/core-service/internal/world-service/repositories/world"
	server_registry_service "github.com/FeedTheRealm-org/core-service/internal/world-service/services/server_registry"
	world_service "github.com/FeedTheRealm-org/core-service/internal/world-service/services/world"
	zones_service "github.com/FeedTheRealm-org/core-service/internal/world-service/services/zones"
	"github.com/gin-gonic/gin"
)

func SetupEndpointsForWorldService(worldGroup *gin.RouterGroup, db *config.DB, conf *config.Config, nomadService server_registry_service.ServerRegistryService) {
	worldRepo := world_repo.NewWorldRepository(conf, db)
	worldService := world_service.NewWorldService(conf, worldRepo, nomadService)
	worldController := world_controller.NewWorldController(conf, worldService)

	worldGroup.POST("", worldController.PublishWorld)
	worldGroup.GET("", worldController.GetWorldsList)
	worldGroup.GET("/:id", worldController.GetWorld)
	worldGroup.PUT("/:id", worldController.UpdateWorld)
	worldGroup.DELETE("/:id", worldController.DeleteWorld)

	worldGroup.PUT("/:id/createable-data", worldController.UpdateCreateableData)

	worldGroup.DELETE("/reset-database", middleware.AdminCheckMiddleware(), worldController.ResetDatabase)
}

func SetupEndpointsForZonesService(worldGroup *gin.RouterGroup, db *config.DB, conf *config.Config, nomadService server_registry_service.ServerRegistryService) {
	worldRepo := world_repo.NewWorldRepository(conf, db)
	zonesService := zones_service.NewZonesService(conf, worldRepo, nomadService)
	zonesController := zones_controller.NewZonesController(conf, zonesService)

	worldGroup.PUT("/:id/zones/:zone_id", zonesController.PublishZone)
	worldGroup.GET("/:id/zones", zonesController.GetWorldZones)
	worldGroup.GET("/:id/zones/:zone_id", zonesController.GetWorldZoneData)
	worldGroup.GET("/:id/zones/:zone_id/activate", zonesController.ActivateZone)
	worldGroup.GET("/:id/zones/:zone_id/deactivate", zonesController.DeactivateZone)
}

func SetupEndpointsForServiceRegistry(orchestratorGroup *gin.RouterGroup, db *config.DB, conf *config.Config, nomadService server_registry_service.ServerRegistryService) {
	worldRepo := world_repo.NewWorldRepository(conf, db)
	worldService := world_service.NewWorldService(conf, worldRepo, nomadService)
	serverRegistryController := server_registry_controller.NewServerRegistryController(conf, worldService, nomadService)

	orchestratorGroup.GET("/:id/zones/:zone_id/start-job", middleware.AdminCheckMiddleware(), serverRegistryController.StartNewJob)
	orchestratorGroup.GET("/:id/zones/:zone_id/stop-job", middleware.AdminCheckMiddleware(), serverRegistryController.StopJob)
	orchestratorGroup.GET("/:id/zones/:zone_id/address", serverRegistryController.GetServerAddress)
}

func CreateNomadService(conf *config.Config) (server_registry_service.ServerRegistryService, error) {
	if conf.Server.Environment == config.Production {
		return server_registry_service.NewServerRegistryService(conf)
	} else {
		return server_registry_service.NewStubServerRegistryService(), nil
	}
}

func SetupWorldServiceRouter(r *gin.Engine, conf *config.Config, db *config.DB) error {
	worldGroup := r.Group("/world")
	orchestratorGroup := worldGroup.Group("/orchestrator")

	nomadService, err := CreateNomadService(conf)
	if err != nil {
		return err
	}

	SetupEndpointsForWorldService(worldGroup, db, conf, nomadService)
	SetupEndpointsForZonesService(worldGroup, db, conf, nomadService)
	SetupEndpointsForServiceRegistry(orchestratorGroup, db, conf, nomadService)

	return nil
}
