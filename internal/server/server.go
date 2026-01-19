package server

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	_ "github.com/joho/godotenv/autoload"

	"github.com/FACorreiaa/skillsphere-pwa/internal/database"
)

// Server holds the dependencies for HTTP handlers
type Server struct {
	port int
	db   database.Service
}

// NewServer creates and configures a new HTTP server
func NewServer() *http.Server {
	port, _ := strconv.Atoi(os.Getenv("PORT"))
	if port == 0 {
		port = 8080
	}

	s := &Server{
		port: port,
		db:   database.New(),
	}

	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", s.port),
		Handler:      s.RegisterRoutes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	return server
}

// GetDB returns the database service (for handlers)
func (s *Server) GetDB() database.Service {
	return s.db
}
