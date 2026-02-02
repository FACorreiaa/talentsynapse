package container

import (
	"context"
	"log"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/FACorreiaa/talentsynapse/internal/app/domain/admin"
	"github.com/FACorreiaa/talentsynapse/internal/app/domain/auth"
	"github.com/FACorreiaa/talentsynapse/internal/app/domain/badges"
	"github.com/FACorreiaa/talentsynapse/internal/app/domain/chat"
	"github.com/FACorreiaa/talentsynapse/internal/app/domain/connections"
	"github.com/FACorreiaa/talentsynapse/internal/app/domain/dashboard"
	"github.com/FACorreiaa/talentsynapse/internal/app/domain/discover"
	"github.com/FACorreiaa/talentsynapse/internal/app/domain/errors"
	"github.com/FACorreiaa/talentsynapse/internal/app/domain/home"
	"github.com/FACorreiaa/talentsynapse/internal/app/domain/matches"
	"github.com/FACorreiaa/talentsynapse/internal/app/domain/points"
	"github.com/FACorreiaa/talentsynapse/internal/app/domain/portfolio"
	"github.com/FACorreiaa/talentsynapse/internal/app/domain/profile"
	"github.com/FACorreiaa/talentsynapse/internal/app/domain/push"
	"github.com/FACorreiaa/talentsynapse/internal/app/domain/report"
	"github.com/FACorreiaa/talentsynapse/internal/app/domain/review"

	"github.com/FACorreiaa/talentsynapse/internal/app/domain/scheduling"
	"github.com/FACorreiaa/talentsynapse/internal/app/domain/session"
	"github.com/FACorreiaa/talentsynapse/internal/app/domain/settings"
	"github.com/FACorreiaa/talentsynapse/internal/app/domain/skills"
	"github.com/FACorreiaa/talentsynapse/internal/app/domain/storage"
	"github.com/FACorreiaa/talentsynapse/internal/app/domain/user"
)

// Container holds all application dependencies
type Container struct {
	// Repositories
	UserRepo      *user.Repository
	SkillsRepo    *skills.Repository
	MatchesRepo   *matches.Repository
	ChatRepo      *chat.Repository
	ReportRepo    *report.Repository
	BadgesRepo    *badges.Repository
	ReviewRepo    *review.Repository
	PortfolioRepo *portfolio.Repository
	SessionRepo   *session.Repository
	DashboardRepo *dashboard.Repository

	// Services
	AuthService   *auth.Service
	BadgeService  *badges.Service
	PointsService *points.Service
	PushService   *push.Service

	// WebSocket
	ChatHub       *chat.Hub
	SchedulingHub *scheduling.Hub

	// Background Services
	SchedulingNotifier *scheduling.Notifier

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
	ReviewHandler      *review.Handler
	PortfolioHandler   *portfolio.Handler
	SessionHandler     *session.Handler
	SchedulingHandler  *scheduling.Handler
	PushHandler        *push.Handler
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
	reviewRepo := review.NewRepository(pool)
	portfolioRepo := portfolio.NewRepository(pool)
	sessionRepo := session.NewRepository(pool)
	dashboardRepo := dashboard.NewRepository(pool)
	pointsRepo := points.NewRepository(pool)

	// Seed admin user from environment variables
	if err := admin.SeedAdmin(context.Background(), userRepo); err != nil {
		log.Printf("Error seeding admin user: %v", err)
	}

	// Create services
	tokenService := auth.NewTokenService(pool)
	emailService := auth.NewEmailService()
	mfaService := auth.NewMFAService(pool, emailService)
	authService := auth.NewService(userRepo, tokenService, emailService, mfaService)

	// Create WebSocket hubs and start them
	chatHub := chat.NewHub()
	go chatHub.Run()

	schedulingRepo := scheduling.NewRepository(pool)
	schedulingService := scheduling.NewService(schedulingRepo)
	schedulingHub := scheduling.NewHub()
	go schedulingHub.Run()

	// Create push service (needed by multiple handlers)
	pushService := push.NewService(pool)

	// Create badge and points services with scheduling hub for notifications
	badgeService := badges.NewService(badgesRepo, schedulingHub)
	pointsService := points.NewService(pointsRepo, schedulingHub)

	// Create background services
	schedulingNotifier := scheduling.NewNotifier(schedulingRepo, schedulingHub, pushService)
	schedulingNotifier.Start()

	// Create handlers
	authHandler := auth.NewHandler(authService)
	homeHandler := home.NewHandler()
	dashboardHandler := dashboard.NewHandler(dashboardRepo)
	profileHandler := profile.NewHandler(userRepo, badgesRepo, reviewRepo, portfolioRepo, matchesRepo, sessionRepo)
	skillsHandler := skills.NewHandler(skillsRepo)
	matchesHandler := matches.NewHandler(matchesRepo, badgeService, reviewRepo, pushService)
	discoverHandler := discover.NewHandler(skillsRepo)
	settingsHandler := settings.NewHandler()

	// Storage service for file uploads
	uploadDir := os.Getenv("UPLOAD_DIR")
	if uploadDir == "" {
		uploadDir = "./uploads"
	}
	baseURL := os.Getenv("BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}
	storageService := storage.NewService(uploadDir, baseURL)

	chatHandler := chat.NewHandler(chatRepo, matchesRepo, chatHub, pushService, storageService)
	connectionsHandler := connections.NewHandler(matchesRepo)
	errorHandler := errors.NewHandler()
	adminHandler := admin.NewHandler(userRepo, reportRepo, badgeService)
	reportHandler := report.NewHandler(reportRepo)
	reviewHandler := review.NewHandler(reviewRepo, badgesRepo, pointsService, userRepo, schedulingHub, pushService)
	portfolioHandler := portfolio.NewHandler(portfolioRepo)
	sessionHandler := session.NewHandler(sessionRepo, badgesRepo, matchesRepo, pointsService)
	schedulingHandler := scheduling.NewHandler(schedulingService, schedulingHub)
	pushHandler := push.NewHandler(pushService)

	return &Container{
		// Repositories
		UserRepo:      userRepo,
		SkillsRepo:    skillsRepo,
		MatchesRepo:   matchesRepo,
		ChatRepo:      chatRepo,
		ReportRepo:    reportRepo,
		BadgesRepo:    badgesRepo,
		ReviewRepo:    reviewRepo,
		PortfolioRepo: portfolioRepo,
		SessionRepo:   sessionRepo,
		DashboardRepo: dashboardRepo,

		// Services
		AuthService:   authService,
		BadgeService:  badgeService,
		PointsService: pointsService,
		PushService:   pushService,

		// WebSocket
		ChatHub:       chatHub,
		SchedulingHub: schedulingHub,

		// Background Services
		SchedulingNotifier: schedulingNotifier,

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
		ReviewHandler:      reviewHandler,
		PortfolioHandler:   portfolioHandler,
		SessionHandler:     sessionHandler,
		SchedulingHandler:  schedulingHandler,
		PushHandler:        pushHandler,
	}
}
