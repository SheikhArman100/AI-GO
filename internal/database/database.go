package database

import (
	"fmt"
	"my-project/internal/logger"
	"my-project/internal/model"
	"os"
	"time"

	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger" // Renamed import to avoid conflict
)

// Service represents a service that interacts with a database.
type Service interface {
	// Health returns a map of health status information.
	Health() map[string]string

	// Close terminates the database connection.
	Close() error

	// DB returns the underlying GORM database instance.
	DB() *gorm.DB
}

// DB returns the underlying GORM database instance.
func (s *service) DB() *gorm.DB {
	return s.db
}

type service struct {
	db *gorm.DB
}

var (
	dbname     = os.Getenv("DB_DATABASE")
	password   = os.Getenv("DB_PASSWORD")
	username   = os.Getenv("DB_USERNAME")
	port       = os.Getenv("DB_PORT")
	host       = os.Getenv("DB_HOST")
	dbInstance *service
)

func New() Service {
	// Add startup logging
	logger.AppLogger.Info("Initializing database service",
		zap.String("host", host),
		zap.String("port", port),
		zap.String("database", dbname),
		zap.String("username", username))

	// Reuse Connection
	if dbInstance != nil {
		logger.AppLogger.Info("Reusing existing database connection",
			zap.String("host", host),
			zap.String("database", dbname))
		return dbInstance
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		username, password, host, port, dbname)

	// Log connection attempt
	logger.AppLogger.Info("Attempting database connection",
		zap.String("host", host),
		zap.String("port", port))

	// Create custom GORM logger
	customLogger := gormLogger.New( // Changed logger.New to gormLogger.New
		&GormWriter{},
		gormLogger.Config{ // Changed logger.Config to gormLogger.Config
			SlowThreshold:             time.Second,
			LogLevel:                  gormLogger.Info, // Changed logger.Info to gormLogger.Info
			IgnoreRecordNotFoundError: true,
			Colorful:                  false,
		},
	)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: customLogger,
	})
	if err != nil {
		logger.AppLogger.Fatal("Database connection failed",
			zap.Error(err),
			zap.String("host", host),
			zap.String("port", port),
			zap.String("database", dbname))
	}

	// Log successful connection
	logger.AppLogger.Info("Database connection established",
		zap.String("host", host),
		zap.String("port", port),
		zap.String("database", dbname))

	// Log migration start
	logger.AppLogger.Info("Starting database migration")

	// Auto Migrate
	if err := db.AutoMigrate(
		&model.User{},
		&model.UserDetail{},
		&model.RefreshToken{},
		&model.Image{},
		&model.SocialProfile{},
		&model.Search{},
		&model.Response{},
	); err != nil {
		logger.AppLogger.Fatal("Database migration failed",
			zap.Error(err),
			zap.String("database", dbname))
	}

	logger.AppLogger.Info("Database migration completed successfully",
		zap.String("database", dbname))

	dbInstance = &service{db: db}
	return dbInstance
}

// Custom writer for GORM that uses our QueryLogger
type GormWriter struct{}

func (w *GormWriter) Printf(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	logger.QueryLogger.Info(msg,
		zap.String("timestamp", time.Now().Format(time.RFC3339)))
}

func (s *service) Health() map[string]string {
	stats := make(map[string]string)
	logger.AppLogger.Info("Performing database health check")

	sqlDB, err := s.db.DB()
	if err != nil {
		stats["status"] = "down"
		stats["error"] = fmt.Sprintf("db down: %v", err)
		logger.AppLogger.Error("Database health check failed - cannot get DB instance",
			zap.Error(err),
			zap.String("database", dbname))
		return stats
	}

	if err := sqlDB.Ping(); err != nil {
		stats["status"] = "down"
		stats["error"] = fmt.Sprintf("db down: %v", err)
		logger.AppLogger.Error("Database health check failed - ping failed",
			zap.Error(err),
			zap.String("database", dbname))
		return stats
	}

	stats["status"] = "up"
	stats["message"] = "Database connection is healthy"
	logger.AppLogger.Info("Database health check passed",
		zap.String("database", dbname))
	return stats
}

func (s *service) Close() error {
	logger.AppLogger.Info("Initiating database connection closure")

	sqlDB, err := s.db.DB()
	if err != nil {
		logger.AppLogger.Error("Failed to get database instance for closure",
			zap.Error(err),
			zap.String("database", dbname))
		return err
	}

	logger.AppLogger.Info("Attempting to close database connection",
		zap.String("database", dbname))

	if err := sqlDB.Close(); err != nil {
		logger.AppLogger.Error("Failed to close database connection",
			zap.Error(err),
			zap.String("database", dbname))
		return err
	}

	logger.AppLogger.Info("Database connection closed successfully",
		zap.String("database", dbname))
	return nil
}
