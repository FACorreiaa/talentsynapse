#!/bin/bash
# Initialize database with migrations
# This script should be run after starting the database container

set -e

echo "Waiting for database to be ready..."
until PGPASSWORD=postgres psql -h localhost -p 5470 -U postgres -d postgres -c '\q' 2>/dev/null; do
  echo "Postgres is unavailable - sleeping"
  sleep 1
done

echo "Database is ready!"
echo "Running migrations..."
goose -dir ./internal/database/migrations postgres "postgres://postgres:postgres@localhost:5470/myapp?sslmode=disable" up

echo "✅ Database initialized successfully!"
