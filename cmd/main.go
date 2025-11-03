package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/FeedTheRealm-org/core-service/config"
	"github.com/FeedTheRealm-org/core-service/internal/server"
	"github.com/FeedTheRealm-org/core-service/internal/utils/logger"
)

func main() {
	conf := config.CreateConfig()
	log := logger.InitLogger(conf.Server.Environment == config.Production)

	s, err := server.NewServer(conf)
	if err != nil {
		log.Fatalf("Error %v", err)
	}

	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel, syscall.SIGTERM)

	go func() {
		<-signalChannel
		close(signalChannel)
		s.Shutdown()
		os.Exit(0)
	}()

	err = s.Start()
	if err != nil {
		log.Fatalf("Error %v", err)
	}
}
