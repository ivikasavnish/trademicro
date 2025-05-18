package api

import (
	"net/http"
	"strings"

	"github.com/go-redis/redis/v8"
	"github.com/gorilla/mux"
	"gorm.io/gorm"

	"github.com/vikasavnish/trademicro/internal/config"
	"github.com/vikasavnish/trademicro/internal/handlers"
	"github.com/vikasavnish/trademicro/internal/middleware"
	"github.com/vikasavnish/trademicro/internal/services"
	"github.com/vikasavnish/trademicro/internal/tasks"
	"github.com/vikasavnish/trademicro/internal/websocket"
)

// SetupRouter configures all routes and returns the router
func SetupRouter(
	db *gorm.DB,
	redisClient *redis.Client,
	wsHub *websocket.Hub,
	cfg *config.Config,
	taskManager *tasks.Manager,
) *mux.Router {
	// Create a new router
	router := mux.NewRouter()

	// Add health check endpoint
	router.HandleFunc("/api/health", HealthHandler).Methods("GET")

	// WebSocket route
	router.HandleFunc("/ws", wsHub.HandleWebSocket)

	// Serve static files
	router.PathPrefix("/static/").Handler(
		http.StripPrefix("/static/", http.FileServer(http.Dir("web/static"))),
	)

	// Create services
	authService := services.NewAuthService(db)
	tradeService := services.NewTradeService(db)
	symbolService := services.NewSymbolService(db)
	userService := services.NewUserService(db)
	familyService := services.NewFamilyMemberService(db)
	brokerTokenService := services.NewBrokerTokenService(db)
	processService := services.NewProcessService()
	favouredSymbolService := services.NewFavouredSymbolService(db)

	// Create handlers using services
	authHandler := handlers.NewAuthHandler(authService, cfg.JWT.SecretKey)
	tradeHandler := handlers.NewTradeHandler(tradeService, wsHub)
	symbolHandler := handlers.NewSymbolHandler(symbolService)
	userHandler := handlers.NewUserHandler(userService)
	familyHandler := handlers.NewFamilyMemberHandler(familyService, userService)
	brokerTokenHandler := handlers.NewBrokerTokenHandler(brokerTokenService, userService, familyService)
	processHandler := handlers.NewProcessHandler(processService)
	favouredSymbolHandler := handlers.NewFavouredSymbolHandler(favouredSymbolService, userService)

	// Add public endpoints directly to the root router (no authentication required)
	router.HandleFunc("/api/login", authHandler.Login).Methods("POST")

	// Create the API router for authenticated endpoints
	apiRouter := router.PathPrefix("/api").Subrouter()

	// Create a subrouter for authenticated endpoints
	authRouter := apiRouter.PathPrefix("").Subrouter()
	authRouter.Use(middleware.AuthMiddleware(cfg.JWT.SecretKey))

	// Register routes
	tradeHandler.RegisterRoutes(authRouter)
	symbolHandler.RegisterRoutes(authRouter)
	userHandler.RegisterRoutes(authRouter)
	familyHandler.RegisterRoutes(authRouter)
	brokerTokenHandler.RegisterRoutes(authRouter)
	processHandler.RegisterRoutes(authRouter)
	favouredSymbolHandler.RegisterRoutes(authRouter)

	// Initialize task management if this is the micro instance
	if cfg.Server.IsMicro {
		// Create task handler with worker configuration
		taskHandler := handlers.NewTaskHandler(cfg.Server)

		// Register task management routes
		taskHandler.RegisterRoutes(authRouter)
	}

	// DhanHQ instrument routes
	instrumentHandler := handlers.NewInstrumentHandler(services.NewInstrumentService(db))
	instrumentHandler.RegisterRoutes(authRouter)

	// Catch-all handler for serving the SPA
	router.PathPrefix("/").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// For API requests, let the router handle them
		if strings.HasPrefix(r.URL.Path, "/api/") {
			http.NotFound(w, r)
			return
		}

		// For all other requests, serve the index.html file
		http.ServeFile(w, r, "web/index.html")
	})

	return router
}
