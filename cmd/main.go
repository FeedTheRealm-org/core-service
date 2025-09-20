package main

import (
	"log"

	"github.com/FeedTheRealm-org/core-service/config"
	"github.com/FeedTheRealm-org/core-service/internal/server"
)

func main() {
	conf := &config.Config{}
	s := server.NewServer(conf)
	err := s.Start()
	if err != nil {
		log.Fatalf("Error %v", err)
	}
}
