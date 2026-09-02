package main

import (
	"fmt"
	"log"
	"os"

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

func initDirs() error {
	watchDir = getEnvValue("DOWNLOAD_DIR")
	targetDir = watchDir + "/Downloaded"
	outputDir = watchDir + "/Files"
	for _, dir := range []string{watchDir, targetDir, outputDir} {
		if err := os.MkdirAll(dir, os.ModePerm); err != nil {
			return fmt.Errorf("Failed to create directory %s: %v", dir, err)
		}
	}
	return nil
}

func main() {
	if err := initDirs(); err != nil {
		log.Panic(err)
	}

	startBot(getEnvValue("TG_BOT_TOKEN"))
	server()
}
