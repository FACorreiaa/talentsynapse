package database

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Service represents the database connection interface
type Service interface {
	Health() map[string]string
	Close() error
	GetPool() *pgxpool.Pool
}

type service struct {
	db *pgxpool.Pool
}

var dbInstance *service

// New creates a new database service or returns existing connection
func New() Service {
	if dbInstance != nil {
		return dbInstance
	}

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		databaseURL = "postgres://localhost:5432/myapp?sslmode=disable"
	}

	// Parse configuration
	config, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		log.Fatalf("Unable to parse database URL: %v", err)
	}

	// Connection pool settings
	config.MaxConns = 25
	config.MinConns = 5
	config.MaxConnLifetime = time.Hour
	config.MaxConnIdleTime = 30 * time.Minute
	config.HealthCheckPeriod = time.Minute

	// Connect to database
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	db, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v", err)
	}

	// Verify connection
	if err := db.Ping(ctx); err != nil {
		log.Fatalf("Unable to ping database: %v", err)
	}

	log.Println("✅ Connected to database")

	dbInstance = &service{db: db}
	return dbInstance
}

// GetPool returns the underlying pgxpool connection
func (s *service) GetPool() *pgxpool.Pool {
	return s.db
}

// Close closes the database connection
func (s *service) Close() error {
	s.db.Close()
	log.Println("Database connection closed")
	return nil
}

// Health checks database connection health
func (s *service) Health() map[string]string {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	stats := s.db.Stat()

	if err := s.db.Ping(ctx); err != nil {
		return map[string]string{
			"status":  "unhealthy",
			"message": err.Error(),
		}
	}

	return map[string]string{
		"status":           "healthy",
		"total_conns":      itoa(int(stats.TotalConns())),
		"acquired_conns":   itoa(int(stats.AcquiredConns())),
		"idle_conns":       itoa(int(stats.IdleConns())),
		"constructing":     itoa(int(stats.ConstructingConns())),
		"max_conns":        itoa(int(stats.MaxConns())),
	}
}

func itoa(i int) string {
	return string(rune('0' + i%10))
}
