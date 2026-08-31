package main

import (
	"context"
	"log"
	"time"

	"github.com/anacrolix/torrent"
)

// Task is a workers assignment.
type Task struct {
	FileName string
	FilePath string
}

func downloadTorrent(workerID int, client *torrent.Client, torrentPath string) error {
	t, err := client.AddTorrentFromFile(torrentPath)
	if err != nil {
		return err
	}
	defer t.Drop()

	<-t.GotInfo() // waiting for metadata
	t.DownloadAll()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		ticker := time.NewTicker(3 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				var pct float64
				stats := t.Stats()
				total := t.Info().TotalLength()
				completed := t.BytesCompleted()
				if total > 0 {
					pct = (float64(completed) / float64(total)) * 100
				}
				log.Printf("[Worker #%d] %s: %.2f%% (%d/%d byte) | Peers: %d", workerID, t.Name(), pct, completed, total, stats.ActivePeers)
			case <-ctx.Done():
				return
			}
		}
	}()

	// The On method returns a chan that closes when the torrent is 100% downloaded.
	<-t.Complete().On()
	return nil
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
