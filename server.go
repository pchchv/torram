package main

import (
	"log"

	"github.com/anacrolix/torrent"
)

func server() {
	// Configuring of torrent-client.
	cfg := torrent.NewDefaultClientConfig()
	cfg.DataDir = outputDir
	client, err := torrent.NewClient(cfg)
	if err != nil {
		log.Fatalf("Error starting the torrent-client: %v", err)
	}
	defer client.Close()
}
