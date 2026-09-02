package main

import (
	"context"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/anacrolix/torrent"
)

type Task struct {
	FileName string
	FilePath string
}

func scanWatchDir(cfg *Config, taskChan chan<- Task, processedFiles *sync.Map) {
	files, err := os.ReadDir(cfg.WatchDir)
	if err != nil {
		log.Printf("[Scanner] Folder read error (will retry): %v", err)
		return
	}

	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(strings.ToLower(file.Name()), ".torrent") {
			continue
		}

		if _, loaded := processedFiles.LoadOrStore(file.Name(), true); loaded {
			continue
		}

		task := Task{
			FileName: file.Name(),
			FilePath: filepath.Join(cfg.WatchDir, file.Name()),
		}

		log.Printf("[Scanner] Sending file %s to the worker pool", file.Name())
		taskChan <- task
	}
}

func downloadTorrent(ctx context.Context, workerID int, client *torrent.Client, torrentPath string) error {
	t, err := client.AddTorrentFromFile(torrentPath)
	if err != nil {
		return err
	}
	defer t.Drop()

	select {
	case <-t.GotInfo():
	case <-ctx.Done():
		return ctx.Err()
	}

	t.DownloadAll()

	doneProgress := make(chan struct{})
	defer close(doneProgress)

	go func() {
		ticker := time.NewTicker(3 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				stats := t.Stats()
				total := t.Info().TotalLength()
				completed := t.BytesCompleted()
				var pct float64
				if total > 0 {
					pct = (float64(completed) / float64(total)) * 100
				}
				log.Printf("[Worker #%d] %s: %.2f%% (%d/%d byte) | Peers: %d", workerID, t.Name(), pct, completed, total, stats.ActivePeers)
			case <-doneProgress: // Безопасный выход без утечек памяти
				return
			case <-ctx.Done():
				return
			}
		}
	}()

	select {
	case <-t.Complete().On():
	return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func worker(ctx context.Context, id int, wg *sync.WaitGroup, tasks <-chan Task, client *torrent.Client, processedFiles *sync.Map, cfg *Config) {
	defer wg.Done()

	for {
		select {
		case <-ctx.Done():
			return
		case task, ok := <-tasks:
			if !ok {
				return
			}

			log.Printf("[Worker #%d] Starting to download: %s", id, task.FileName)
			err := downloadTorrent(ctx, id, client, task.FilePath)
		if err != nil {
			log.Printf("[Worker #%d] Download error %s: %v", id, task.FileName, err)
				processedFiles.Delete(task.FileName)
			continue
		}

		// Move the torrent file to the destination folder after the content has been successfully downloaded.
			destPath := filepath.Join(cfg.TargetDir, task.FileName)
			if err = os.Rename(task.FilePath, destPath); err != nil {
				log.Printf("[Worker #%d] Failed to move torrent file %s: %v", id, task.FileName, err)
		} else {
				log.Printf("[Worker #%d] Torrent %s successfully downloaded and moved to %s", id, task.FileName, cfg.TargetDir)
			}
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

	maxWorkers, err := strconv.Atoi(getEnvValue("MAX_NUM_WORKERS"))
	if err != nil {
		log.Panicf("Error getting the maximum number of workers: %e", err)
	}

	var processedFiles sync.Map // Protection against duplicate tasks in the pool
	taskChan := make(chan Task, 100)
	// Launching a fixed pool of workers
	var wg sync.WaitGroup
	for i := 1; i <= maxWorkers; i++ {
		wg.Add(1)
		go worker(i, &wg, taskChan, client, &processedFiles)
	}

	log.Printf("The service has started.\nThe worker pool (%d) is ready to run...", maxWorkers)

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		scanWatchDir(taskChan, &processedFiles)
	}

	close(taskChan)
	wg.Wait()
}
