package container

import (
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/FACorreiaa/skillsphere/internal/app/auth"
	"github.com/FACorreiaa/skillsphere/internal/app/chat"
	"github.com/FACorreiaa/skillsphere/internal/app/connections"
	"github.com/FACorreiaa/skillsphere/internal/app/dashboard"
	"github.com/FACorreiaa/skillsphere/internal/app/discover"
	"github.com/FACorreiaa/skillsphere/internal/app/home"
	"github.com/FACorreiaa/skillsphere/internal/app/matches"
	"github.com/FACorreiaa/skillsphere/internal/app/profile"
	"github.com/FACorreiaa/skillsphere/internal/app/settings"
	"github.com/FACorreiaa/skillsphere/internal/app/skills"
	"github.com/FACorreiaa/skillsphere/internal/app/user"
)

// Container holds all application dependencies
type Container struct {
	// Repositories
	UserRepo    *user.Repository
	SkillsRepo  *skills.Repository
	MatchesRepo *matches.Repository
	ChatRepo    *chat.Repository

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

	// Create services
	authService := auth.NewService(userRepo)

	// Create WebSocket hub and start it
	chatHub := chat.NewHub()
	go chatHub.Run()

	// Create handlers
	authHandler := auth.NewHandler(authService)
	homeHandler := home.NewHandler()
	dashboardHandler := dashboard.NewHandler()
	profileHandler := profile.NewHandler(userRepo)
	skillsHandler := skills.NewHandler(skillsRepo)
	matchesHandler := matches.NewHandler(matchesRepo)
	discoverHandler := discover.NewHandler(skillsRepo)
	settingsHandler := settings.NewHandler()
	chatHandler := chat.NewHandler(chatRepo, chatHub)
	connectionsHandler := connections.NewHandler(matchesRepo)

	return &Container{
		// Repositories
		UserRepo:    userRepo,
		SkillsRepo:  skillsRepo,
		MatchesRepo: matchesRepo,
		ChatRepo:    chatRepo,

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
	}
}
