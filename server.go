package main

import (
	"context"
	"log"
	"os"
	"path/filepath"
	"sync"
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

func worker(id int, wg *sync.WaitGroup, tasks <-chan Task, client *torrent.Client, processedFiles *sync.Map) {
	defer wg.Done()

	for task := range tasks {
		log.Printf("[Worker #%d] Starting to download the file: %s", id, task.FileName)
		err := downloadTorrent(id, client, task.FilePath)
		if err != nil {
			log.Printf("[Worker #%d] Download error %s: %v", id, task.FileName, err)
			processedFiles.Delete(task.FileName) // Reset the retry flag
			continue
		}

		// Move the torrent file to the destination folder after the content has been successfully downloaded.
		err = os.Rename(task.FilePath, filepath.Join(targetDir, task.FileName))
		if err != nil {
			log.Printf("[Worker #%d] Failed to move the torrent file %s: %v", id, task.FileName, err)
		} else {
			log.Printf("[Worker #%d] The torrent %s was successfully downloaded and moved to %s", id, task.FileName, targetDir)
		}
	}
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
