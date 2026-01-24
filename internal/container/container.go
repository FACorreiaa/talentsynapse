package container

import (
	"context"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/FACorreiaa/skillsphere/internal/app/domain/admin"
	"github.com/FACorreiaa/skillsphere/internal/app/domain/auth"
	"github.com/FACorreiaa/skillsphere/internal/app/domain/badges"
	"github.com/FACorreiaa/skillsphere/internal/app/domain/chat"
	"github.com/FACorreiaa/skillsphere/internal/app/domain/connections"
	"github.com/FACorreiaa/skillsphere/internal/app/domain/dashboard"
	"github.com/FACorreiaa/skillsphere/internal/app/domain/discover"
	"github.com/FACorreiaa/skillsphere/internal/app/domain/errors"
	"github.com/FACorreiaa/skillsphere/internal/app/domain/home"
	"github.com/FACorreiaa/skillsphere/internal/app/domain/matches"
	"github.com/FACorreiaa/skillsphere/internal/app/domain/profile"
	"github.com/FACorreiaa/skillsphere/internal/app/domain/report"
	"github.com/FACorreiaa/skillsphere/internal/app/domain/settings"
	"github.com/FACorreiaa/skillsphere/internal/app/domain/skills"
	"github.com/FACorreiaa/skillsphere/internal/app/domain/user"
)

// Container holds all application dependencies
type Container struct {
	// Repositories
	UserRepo    *user.Repository
	SkillsRepo  *skills.Repository
	MatchesRepo *matches.Repository
	ChatRepo    *chat.Repository
	ReportRepo  *report.Repository
	BadgesRepo  *badges.Repository

	// Services
	AuthService *auth.Service

	// WebSocket
	ChatHub *chat.Hub

	// Handlers
	AuthHandler        *auth.Handler
	HomeHandler        *home.Handler
	DashboardHandler   *dashboard.Handler
	ProfileHandler     *profile.Handler
	SkillsHandler      *skills.Handler
	MatchesHandler     *matches.Handler
	DiscoverHandler    *discover.Handler
	SettingsHandler    *settings.Handler
	ChatHandler        *chat.Handler
	ConnectionsHandler *connections.Handler
	ErrorHandler       *errors.Handler
	AdminHandler       *admin.Handler
	ReportHandler      *report.Handler
}

// New creates a new dependency injection container
func New(pool *pgxpool.Pool) *Container {
	// Initialize session store
	auth.InitStore()

	// Create repositories
	userRepo := user.NewRepository(pool)
	skillsRepo := skills.NewRepository(pool)
	matchesRepo := matches.NewRepository(pool)
	chatRepo := chat.NewRepository(pool)
	reportRepo := report.NewRepository(pool)
	badgesRepo := badges.NewRepository(pool)

	// Seed admin user from environment variables
	if err := admin.SeedAdmin(context.Background(), userRepo); err != nil {
		log.Printf("Error seeding admin user: %v", err)
	}

	// Create services
	authService := auth.NewService(userRepo)

	// Create WebSocket hub and start it
	chatHub := chat.NewHub()
	go chatHub.Run()

	// Create handlers
	authHandler := auth.NewHandler(authService)
	homeHandler := home.NewHandler()
	dashboardHandler := dashboard.NewHandler()
	profileHandler := profile.NewHandler(userRepo, badgesRepo)
	skillsHandler := skills.NewHandler(skillsRepo)
	matchesHandler := matches.NewHandler(matchesRepo)
	discoverHandler := discover.NewHandler(skillsRepo)
	settingsHandler := settings.NewHandler()
	chatHandler := chat.NewHandler(chatRepo, chatHub)
	connectionsHandler := connections.NewHandler(matchesRepo)
	errorHandler := errors.NewHandler()
	adminHandler := admin.NewHandler(userRepo, reportRepo)
	reportHandler := report.NewHandler(reportRepo)

	return &Container{
		// Repositories
		UserRepo:    userRepo,
		SkillsRepo:  skillsRepo,
		MatchesRepo: matchesRepo,
		ChatRepo:    chatRepo,
		ReportRepo:  reportRepo,
		BadgesRepo:  badgesRepo,

		// Services
		AuthService: authService,

		// WebSocket
		ChatHub: chatHub,

		// Handlers
		AuthHandler:        authHandler,
		HomeHandler:        homeHandler,
		DashboardHandler:   dashboardHandler,
		ProfileHandler:     profileHandler,
		SkillsHandler:      skillsHandler,
		MatchesHandler:     matchesHandler,
		DiscoverHandler:    discoverHandler,
		SettingsHandler:    settingsHandler,
		ChatHandler:        chatHandler,
		ConnectionsHandler: connectionsHandler,
		ErrorHandler:       errorHandler,
		AdminHandler:       adminHandler,
		ReportHandler:      reportHandler,
	}
}
