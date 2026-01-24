// Package main serves as the primary entry point for the Go Clean Template application.
// It initializes global configurations and bootstraps the application runner.
//
// @title           Go Clean Template API
// @version         1.0
// @description     A modular and scalable Go REST API template implementing Clean Architecture.
// @termsOfService  http://swagger.io/terms/
//
// @contact.name    API Support
// @contact.url     http://www.swagger.io/support
// @contact.email   support@swagger.io
//
// @license.name    Apache 2.0
// @license.url     http://www.apache.org/licenses/LICENSE-2.0.html
//
// @host            localhost:8080
// @BasePath        /api/v1
// @securityDefinitions.apikey BearerAuth
// @in              header
// @name            Authorization
package main

import (
	"log"

	"gct/config"
	"gct/internal/app"
)

// main is the application bootstrap function.
// It follows a two-step process: 1. Load environment-specific configurations.
// 2. Pass those configurations to the core application runner.
func main() {
	// 1. Initialize Configuration
	// Reads from environment variables and .env files to build the Config object.
	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatalf("Config initialization error: %s", err)
	}

	// 2. Execute Application Runner
	// Orchestrates the setup of database connections, usecases, and the HTTP server.
	app.Run(cfg)
}
