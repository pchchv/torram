package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"syscall"

	"github.com/pchchv/env"
)

// Config stores all the service's settings
type Config struct {
	WatchDir   string
	TargetDir  string
	OutputDir  string
	BotToken   string
	MaxWorkers int
}

func loadConfig() (*Config, error) {
	watchDir, err := getEnvValue("DOWNLOAD_DIR")
	if err != nil {
		return nil, err
	}

	botToken, err := getEnvValue("TG_BOT_TOKEN")
	if err != nil {
		return nil, err
	}

	maxWorkersStr, err := getEnvValue("MAX_NUM_WORKERS")
	if err != nil {
		return nil, err
	}

	maxWorkers, err := strconv.Atoi(maxWorkersStr)
	if err != nil {
		return nil, fmt.Errorf("invalid MAX_NUM_WORKERS value: %w", err)
	}

	return &Config{
		WatchDir:   watchDir,
		TargetDir:  filepath.Join(watchDir, "Downloaded"),
		OutputDir:  filepath.Join(watchDir, "Files"),
		BotToken:   botToken,
		MaxWorkers: maxWorkers,
	}, nil
}

func init() {
	// Load values from .env into the system.
	if err := env.Load(); err != nil {
		log.Panic("No .env file found")
	}
}

func getEnvValue(v string) (string, error) {
	value, exist := os.LookupEnv(v)
	if !exist {
		return value, fmt.Errorf("Value %v does not exist", v)
	}

	return value, nil
}

func ensureDirs(cfg *Config) error {
	for _, dir := range []string{cfg.WatchDir, cfg.TargetDir, cfg.OutputDir} {
		if err := os.MkdirAll(dir, os.ModePerm); err != nil {
			return fmt.Errorf("creating directory %s: %w", dir, err)
		}
	}
	return nil
}

func main() {
	cfg, err := loadConfig()
	if err != nil {
		log.Fatalf("[Main] Configuration error: %v", err)
	}

	if err := ensureDirs(cfg); err != nil {
		log.Fatalf("[Main] Directory initialization error: %v", err)
	}

	// Root context for a graceful shutdown of the entire application
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	log.Println("[Main] Starting application modules...")

	// Running a torrent server (it will launch a pool of workers and a scanner internally)
	go func() {
		if err := runServer(ctx, cfg); err != nil {
			log.Printf("[Main] Server stopped with error: %v", err)
			stop()
		}
	}()

	// Running a Telegram Bot
	go func() {
		if err := runBot(ctx, cfg); err != nil {
			log.Printf("[Main] Bot stopped with error: %v", err)
			stop()
		}
	}()

	// Wait for the termination signal
	<-ctx.Done()
	log.Println("[Main] Shutting down gracefully...")
}
