package router

import (
	"github.com/FeedTheRealm-org/core-service/config"
	exports_controller "github.com/FeedTheRealm-org/core-service/internal/exports-service/controllers/exports"
	"github.com/FeedTheRealm-org/core-service/internal/exports-service/repositories/bucket"
	exports_repo "github.com/FeedTheRealm-org/core-service/internal/exports-service/repositories/exports"
	exports_service "github.com/FeedTheRealm-org/core-service/internal/exports-service/services/exports"
	"github.com/FeedTheRealm-org/core-service/internal/middleware"
	"github.com/gin-gonic/gin"
)

func getNewBucketRepository(name string, conf *config.Config) (bucket.BucketRepository, error) {
	if conf.Server.Environment == config.Development || conf.Server.Environment == config.Testing {
		return bucket.NewOnDiskBucketRepository(name, conf)
	}
	return bucket.NewAwsS3BucketRepository(name, conf)
}

func SetupExportsServiceRouter(r *gin.Engine, conf *config.Config, db *config.DB) error {
	g := r.Group("/exports")

	worldsBucketRepo, err := getNewBucketRepository(conf.Assets.WorldsBucketName, conf)
	if err != nil {
		return err
	}

	exportsRepo := exports_repo.NewExportRepository(conf, db)
	exportsService := exports_service.NewExportsService(conf, exportsRepo, worldsBucketRepo)
	exportsController := exports_controller.NewExportsController(conf, exportsService)

	g.PUT("/zip", middleware.AdminCheckMiddleware(), exportsController.UploadZip)
	g.GET("/zip", exportsController.GetZipPath)
	g.GET("/zip/versions", middleware.AdminCheckMiddleware(), exportsController.ListZipVersions)
	g.DELETE("/zip", middleware.AdminCheckMiddleware(), exportsController.DeleteZipVersion)
	g.PATCH("/zip/latest", middleware.AdminCheckMiddleware(), exportsController.SetLatestZipVersion)

	return nil
}
