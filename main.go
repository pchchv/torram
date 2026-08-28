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
	startBot(getEnvValue("TG_BOT_TOKEN"))
}
