package server

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	_ "github.com/joho/godotenv/autoload"
	"go.uber.org/zap"

	"my-project/internal/database"
	"my-project/internal/logger"
)

type Server struct {
	port int

	db database.Service
}

func NewServer() *http.Server {
	// Initialize logger
	if err := logger.InitLoggers(); err != nil {
		log.Fatalf("Failed to initialize loggers: %v", err)
	}

	port, err := strconv.Atoi(os.Getenv("PORT"))
	if err != nil {
		logger.AppLogger.Error("Invalid port number, using default 8080",
			zap.Error(err),
			zap.String("port", os.Getenv("PORT")))
		port = 8080
	}

	NewServer := &Server{
		port: port,
		db:   database.New(),
	}

	logger.AppLogger.Info("Server initialization",
		zap.Int("port", port),
		zap.String("environment", os.Getenv("ENV")),
		zap.String("version", "1.0.0"))

	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", NewServer.port),
		Handler:      NewServer.RegisterRoutes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	return server
}
