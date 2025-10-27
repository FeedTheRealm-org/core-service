package server

import (
	"context"
	"net/http"
	"strconv"

	"github.com/FeedTheRealm-org/core-service/config"
	"github.com/FeedTheRealm-org/core-service/internal/router"
	"github.com/FeedTheRealm-org/core-service/internal/utils/logger"
	"github.com/gin-gonic/gin"
)

type Server struct {
	conf *config.Config
	srv  *http.Server
}

func NewServer(conf *config.Config) *Server {
	return &Server{conf: conf}
}

func (s *Server) Start() error {
	gin.SetMode(gin.ReleaseMode)

	r := gin.Default()
	router.SetupRouter(r, s.conf)

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
