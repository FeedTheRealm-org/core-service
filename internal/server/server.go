package server

import (
	"github.com/FeedTheRealm-org/core-service/config"
	"github.com/FeedTheRealm-org/core-service/internal/router"
	"github.com/FeedTheRealm-org/core-service/internal/utils/logger"
	"github.com/gin-gonic/gin"
)

type Server struct {
	conf *config.Config
}

func NewServer(conf *config.Config) *Server {
	return &Server{conf: conf}
}

func (s *Server) Start() error {
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	// Routes
	router.SetupRouter(r, s.conf)

	logger.GetLogger().Info("Starting server on port 8080")
	return r.Run("0.0.0.0:8080")
}
