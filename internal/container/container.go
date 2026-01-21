package container

import (
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/FACorreiaa/skillsphere-pwa/internal/app/auth"
	"github.com/FACorreiaa/skillsphere-pwa/internal/app/dashboard"
	"github.com/FACorreiaa/skillsphere-pwa/internal/app/home"
	"github.com/FACorreiaa/skillsphere-pwa/internal/app/user"
)

// Container holds all application dependencies
type Container struct {
	// Repositories
	UserRepo *user.Repository

	// Services
	AuthService *auth.Service

	// Handlers
	AuthHandler      *auth.Handler
	HomeHandler      *home.Handler
	DashboardHandler *dashboard.Handler
}

// New creates a new dependency injection container
func New(pool *pgxpool.Pool) *Container {
	// Initialize session store
	auth.InitStore()

	// Create repositories
	userRepo := user.NewRepository(pool)

	// Create services
	authService := auth.NewService(userRepo)

	// Create handlers
	authHandler := auth.NewHandler(authService)
	homeHandler := home.NewHandler()
	dashboardHandler := dashboard.NewHandler()

	return &Container{
		// Repositories
		UserRepo: userRepo,

		// Services
		AuthService: authService,

		// Handlers
		AuthHandler:      authHandler,
		HomeHandler:      homeHandler,
		DashboardHandler: dashboardHandler,
	}
}
