package server

import (
	"context"
	"net/http"
	"strconv"

	"github.com/FeedTheRealm-org/core-service/config"
	"github.com/FeedTheRealm-org/core-service/internal/router"
	"github.com/FeedTheRealm-org/core-service/internal/utils/logger"
	"github.com/FeedTheRealm-org/core-service/internal/utils/seed_database"
	"github.com/gin-gonic/gin"
)

type Server struct {
	conf *config.Config
	db   *config.DB
	srv  *http.Server
}

func NewServer(conf *config.Config) (*Server, error) {
	db, err := config.NewDB(conf)
	if err != nil {
		return nil, err
	}
	return &Server{
		conf: conf,
		db:   db,
	}, nil
}

func (s *Server) Start() error {
	switch s.conf.Server.Environment {
	case config.Development:
		gin.SetMode(gin.DebugMode)
	case config.Testing:
		gin.SetMode(gin.TestMode)
		err := seed_database.SeedDatabase(s.db)
		if err != nil {
			logger.GetLogger().Errorf("Failed to seed database: %v", err)
			return err
		}
	case config.Production:
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.Default()
	router.SetupRouter(r, s.conf, s.db)

	s.srv = &http.Server{
		Addr:    "0.0.0.0:" + strconv.Itoa(s.conf.Server.Port),
		Handler: r,
	}

	logger.GetLogger().Info("Starting server on port " + strconv.Itoa(s.conf.Server.Port))
	if err := s.srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}

	return nil
}

func (s *Server) Shutdown() {
	logger.GetLogger().Info("Shutting down server")
	ctx, cancel := context.WithTimeout(context.Background(), s.conf.Server.ShutdownTimeout)
	defer cancel()

	if err := s.srv.Shutdown(ctx); err != nil {
		logger.GetLogger().Errorf("Server forced to shutdown: %v", err)
	} else {
		logger.GetLogger().Info("Server exited properly")
	}
}
