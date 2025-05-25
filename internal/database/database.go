package database

import (
	"fmt"
	"log"
	"my-project/internal/model"
	"os"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/joho/godotenv/autoload"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
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
	// Reuse Connection
	if dbInstance != nil {
		return dbInstance
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		username, password, host, port, dbname)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}

	// Auto Migrate the schema
	if err := db.AutoMigrate(&model.User{}, &model.UserDetail{}, &model.RefreshToken{}, &model.Image{}, &model.SocialProfile{}, &model.Search{}, &model.Response{}); err != nil {
		log.Fatal("Failed to migrate database schema:", err)
	}

	dbInstance = &service{
		db: db,
	}
	return dbInstance
}

// Health checks the health of the database connection.
// It returns a map with basic health status information.
func (s *service) Health() map[string]string {
	stats := make(map[string]string)

	sqlDB, err := s.db.DB()
	if err != nil {
		stats["status"] = "down"
		stats["error"] = fmt.Sprintf("db down: %v", err)
		log.Printf("db down: %v", err)
		return stats
	}

	// Ping the database
	if err := sqlDB.Ping(); err != nil {
		stats["status"] = "down"
		stats["error"] = fmt.Sprintf("db down: %v", err)
		log.Printf("db down: %v", err)
		return stats
	}

	// Database is up
	stats["status"] = "up"
	stats["message"] = "Database connection is healthy"
	return stats
}

// Close closes the database connection.
func (s *service) Close() error {
	sqlDB, err := s.db.DB()
	if err != nil {
		return err
	}
	log.Printf("Disconnected from database: %s", dbname)
	return sqlDB.Close()
}
