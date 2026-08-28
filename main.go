package main

import (
	"log"

	"github.com/pchchv/env"
)

func init() {
	// Load values from .env into the system
	if err := env.Load(); err != nil {
		log.Panic("No .env file found")
	}
}

func main() {}
