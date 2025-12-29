// Package main implements the main application entry point.
// @title       Go Clean Template API
// @description Using a user management service as an example
// @version     1.0
// @host        localhost:8080
// @BasePath    /api
package main

import (
	"log"

	"gct/config"
	"gct/internal/app"
)

func main() {
	// Configuration
	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatalf("Config error: %s", err)
	}

	// Run
	app.Run(cfg)
}
