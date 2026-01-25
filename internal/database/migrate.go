package database

import (
	"database/sql"
	"embed"
	"fmt"
	"log"

	_ "github.com/jackc/pgx/v5/stdlib" // PostgreSQL driver
	"github.com/pressly/goose/v3"
)

//go:embed migrations/*.sql
var embedMigrations embed.FS

// RunMigrations runs all pending database migrations
// This is called automatically when the application starts
func RunMigrations(databaseURL string) error {
	log.Println("🔄 Checking for pending database migrations...")

	// Open a connection for goose (uses database/sql with pgx driver)
	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		return fmt.Errorf("failed to open database for migrations: %w", err)
	}
	defer func() {
		if closeErr := db.Close(); closeErr != nil {
			log.Printf("Warning: failed to close migration database connection: %v", closeErr)
		}
	}()

	// Set goose to use embedded migrations
	goose.SetBaseFS(embedMigrations)

	// Run migrations
	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("failed to set goose dialect: %w", err)
	}

	if err := goose.Up(db, "migrations"); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	// Get current version
	version, err := goose.GetDBVersion(db)
	if err != nil {
		log.Printf("⚠️  Warning: could not get migration version: %v", err)
	} else {
		log.Printf("✅ Database migrations complete (current version: %d)", version)
	}

	return nil
}

// GetMigrationStatus returns the current migration status
func GetMigrationStatus(databaseURL string) (int64, error) {
	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		return 0, fmt.Errorf("failed to open database: %w", err)
	}
	defer func() {
		if closeErr := db.Close(); closeErr != nil {
			log.Printf("Warning: failed to close database connection: %v", closeErr)
		}
	}()

	if err := goose.SetDialect("postgres"); err != nil {
		return 0, fmt.Errorf("failed to set goose dialect: %w", err)
	}

	version, err := goose.GetDBVersion(db)
	if err != nil {
		return 0, fmt.Errorf("failed to get migration version: %w", err)
	}

	return version, nil
}
