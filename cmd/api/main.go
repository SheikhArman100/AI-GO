package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"

	"my-project/internal/logger"
	"my-project/internal/server"
)

func main() {
	// Initialize logger
	if err := logger.InitLoggers(); err != nil {
		log.Fatalf("Failed to initialize loggers: %v", err)
	}

	server := server.NewServer()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		logger.AppLogger.Info("Server starting",
			zap.String("address", server.Addr))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.AppLogger.Fatal("Server failed to start",
				zap.Error(err))
		}
	}()

	<-quit
	logger.AppLogger.Info("Server shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.AppLogger.Fatal("Server forced to shutdown",
			zap.Error(err))
	}

	logger.AppLogger.Info("Server exited properly")
}
