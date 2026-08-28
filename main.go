package main

import (
	"log"
	"os"

	"github.com/pchchv/env"
)

func init() {
	// Load values from .env into the system.
	if err := env.Load(); err != nil {
		log.Panic("No .env file found")
	}
}

func getEnvValue(v string) string {
	// Getting a value.
	// Outputs a panic if the value is missing.
	value, exist := os.LookupEnv(v)
	if !exist {
		log.Panicf("Value %v does not exist", v)
	}

	return value
}

func main() {
	downloadDir := getEnvValue("DOWNLOAD_DIR")
	// Create a Downloads folder if it doesn't exist.
	if err := os.MkdirAll(downloadDir, os.ModePerm); err != nil {
		log.Panicf("folder creation error: %v\n", err)
	}

	// Create a folder for downloaded files if one does not already exist.
	if err := os.MkdirAll(downloadDir+"/Downloaded", os.ModePerm); err != nil {
		log.Panicf("folder creation error: %v\n", err)
	}

	startBot(getEnvValue("TG_BOT_TOKEN"))
}
