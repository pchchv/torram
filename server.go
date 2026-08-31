package main

import (
	"log"

	"github.com/anacrolix/torrent"
)

// Task is a worker's assignment.
type Task struct {
	FileName string
	FilePath string
}

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
