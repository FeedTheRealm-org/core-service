package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/FeedTheRealm-org/core-service/config"
	"github.com/FeedTheRealm-org/core-service/internal/server"
)

func shutdown(s *server.Server) {
	s.Shutdown()
}

func main() {
	conf := config.CreateConfig()
	s := server.NewServer(conf)

	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel, syscall.SIGTERM)

	go func() {
		<-signalChannel
		close(signalChannel)
		shutdown(s)
		os.Exit(0)
	}()

	err := s.Start()
	if err != nil {
		log.Fatalf("Error %v", err)
	}
}
