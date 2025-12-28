// Package app configures and runs application.
package app

import (
	"context"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/evrone/go-clean-template/config"
	"github.com/evrone/go-clean-template/internal/controller/restapi"
	"github.com/evrone/go-clean-template/internal/repo"
	"github.com/evrone/go-clean-template/internal/usecase"
	"github.com/gin-gonic/gin"

	"github.com/evrone/go-clean-template/pkg/db/postgres"
	"github.com/evrone/go-clean-template/pkg/logger"
	httpserver "github.com/evrone/go-clean-template/pkg/server/http"
	"go.uber.org/zap"
)

// Run creates objects via constructors.
func Run(cfg *config.Config) {
	l := logger.New(cfg.Log.Level)

	// Repository
	pg, err := postgres.New(context.Background(), cfg.App.Environment, cfg.Database.Postgres, l)
	if err != nil {
		l.Fatalw("app - Run - postgres.New", zap.Error(err))
	}
	defer pg.Close()

	repositories := repo.New(pg, l)

	// Use Case
	useCases := usecase.NewUseCase(repositories, l, cfg)

	// HTTP Server logic
	gin.ForceConsoleColor()

	if cfg.App.IsProd() {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	handler := gin.New()
	restapi.NewRouter(handler, cfg, useCases, l)

	httpServer := httpserver.NewServer()

	// Parse port
	port, err := strconv.Atoi(cfg.HTTP.Port)
	if err != nil {
		l.Fatalw("app - Run - strconv.Atoi", zap.Error(err))
	}

	// Start server
	go func() {
		if err := httpServer.Run(port, handler); err != nil {
			l.Errorw("app - Run - httpServer.Run", zap.Error(err))
		}
	}()

	// Waiting signal
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	select {
	case s := <-interrupt:
		l.Infow("app - Run - signal", zap.String("signal", s.String()))
		// case err = <-httpServer.Notify():
		// 	l.Errorw("app - Run - httpServer.Notify", zap.Error(err))
	}

	// Shutdown
	err = httpServer.Shutdown(context.Background())
	if err != nil {
		l.Errorw("app - Run - httpServer.Shutdown", zap.Error(err))
	}
}
