package database

import (
	"context"
	"log"
	"os"
	"strconv"
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
		databaseURL = "postgres://localhost:5470/myapp?sslmode=disable"
	}

	// Run migrations before connecting
	if err := RunMigrations(databaseURL); err != nil {
		log.Printf("⚠️  Migration failed: %v", err)
		log.Println("⚠️  Continuing with connection attempt...")
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
	// Check if database pool is initialized
	if s.db == nil {
		return map[string]string{
			"status":  "unhealthy",
			"message": "database connection not initialized",
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Ping database first
	if err := s.db.Ping(ctx); err != nil {
		return map[string]string{
			"status":  "unhealthy",
			"message": err.Error(),
		}
	}

	// Get stats (may panic if connection is invalid)
	stats := s.db.Stat()

	return map[string]string{
		"status":         "healthy",
		"total_conns":    strconv.Itoa(int(stats.TotalConns())),
		"acquired_conns": strconv.Itoa(int(stats.AcquiredConns())),
		"idle_conns":     strconv.Itoa(int(stats.IdleConns())),
		"constructing":   strconv.Itoa(int(stats.ConstructingConns())),
		"max_conns":      strconv.Itoa(int(stats.MaxConns())),
	}
}
