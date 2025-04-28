package main

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/go-redis/redis/v8"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
	"github.com/rs/cors"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/vikasavnish/trademicro/api"
)

// Global variables
var (
	db           *gorm.DB
	redisClient  *redis.Client
	connections  = make(map[*websocket.Conn]bool)
	broadcast    = make(chan Message)
	upgrader     = websocket.Upgrader{}
	jwtSecretKey = []byte(os.Getenv("SECRET_KEY"))
)

// Models

type FamilyMember struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	UserID      uint      `json:"user_id"`
	Name        string    `json:"name"`
	Email       string    `json:"email"`
	Phone       string    `json:"phone"`
	Pin         string    `json:"pin"`
	PortfolioID *uint     `json:"portfolioId"`
	IsActive    bool      `json:"isActive"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
	ClientToken string    `json:"client_token" gorm:"-"`
	ClientID    string    `json:"client_id" gorm:"-"`
}
type User struct {
	ID             uint   `gorm:"primaryKey" json:"id"`
	Username       string `gorm:"unique" json:"username"`
	Password       string `json:"password,omitempty" gorm:"column:password"`
	HashedPassword string `json:"-" gorm:"column:hashed_password"`
	Email          string `json:"email"`
	Role           string `json:"role"`
}

type TradeOrder struct {
	ID     uint    `gorm:"primaryKey" json:"id"`
	Symbol string  `json:"symbol"`
	Unit   int     `json:"unit"`
	Diff   float64 `json:"diff"`
	Zag    int     `json:"zag"`
	Type   string  `json:"type"`
	User   string  `json:"user"`
	Status string  `json:"status" gorm:"default:running"`
}

type BrokerToken struct {
	ID        uint   `gorm:"primaryKey" json:"id"`
	Broker    string `json:"broker"`
	Token     string `json:"token"`
	User      string `json:"user"`
	CreatedAt time.Time
}

type Symbol struct {
	ID     uint   `gorm:"primaryKey" json:"id"`
	Symbol string `json:"symbol"`
	Name   string `json:"name"`
}

type Message struct {
	Type    string      `json:"type"`
	Content interface{} `json:"content"`
}

type TokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// JWT Claims
type Claims struct {
	Username string `json:"username"`
	jwt.StandardClaims
}

func init() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	// Set JWT secret key
	jwtSecretKey = []byte(os.Getenv("SECRET_KEY"))
	if len(jwtSecretKey) == 0 {
		jwtSecretKey = []byte("default_secret_key")
		log.Println("Warning: Using default JWT secret key")
	}

	// Initialize database connection
	initDB()

	// Initialize Redis client
	initRedis()

	// Allow all origins for WebSocket connections
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }
}

func initDB() {
	// ...
	// Add FamilyMember to migrations

	// ...

	dbURL := os.Getenv("POSTGRES_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:password@localhost:5432/trademicro"
		log.Println("Warning: Using default database URL")
	}

	// Use the PostgreSQL DSN directly
	var err error
	db, err = gorm.Open(postgres.Open(dbURL), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Auto migrate the schema

	// Create a default admin user if none exists
	var userCount int64
	db.Model(&User{}).Count(&userCount)
	if userCount == 0 {
		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("Servloci@54321"), bcrypt.DefaultCost)
		db.Create(&User{
			Username:       "vikasavnish",
			HashedPassword: string(hashedPassword),
			Email:          "bizpowersolution@gmail.com",
			Role:           "admin",
		})
		log.Println("Created default admin user")
	}
}

func initRedis() {
	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		redisURL = "redis://localhost:6379/0"
		log.Println("Warning: Using default Redis URL")
	}

	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		log.Fatalf("Failed to parse Redis URL: %v", err)
	}

	redisClient = redis.NewClient(opt)
	ctx := context.Background()

	// Test the connection
	if _, err := redisClient.Ping(ctx).Result(); err != nil {
		log.Printf("Warning: Failed to connect to Redis: %v", err)
	}
}

// Authentication middleware
func authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authorizationHeader := r.Header.Get("Authorization")
		if authorizationHeader == "" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Extract the token
		tokenString := authorizationHeader[7:] // Remove "Bearer " prefix

		// Parse and validate the token
		claims := &Claims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			return jwtSecretKey, nil
		})

		if err != nil || !token.Valid {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Add the username to the request context
		ctx := context.WithValue(r.Context(), "username", claims.Username)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// Login handler
func loginHandler(w http.ResponseWriter, r *http.Request) {
	var loginReq LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&loginReq); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Find the user
	var user User
	result := db.Where("username = ?", loginReq.Username).First(&user)
	if result.Error != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	// Check password
	err := bcrypt.CompareHashAndPassword([]byte(user.HashedPassword), []byte(loginReq.Password))
	if err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	// Create a JWT token
	expirationTime := time.Now().Add(60 * time.Minute)
	claims := &Claims{
		Username: user.Username,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtSecretKey)
	if err != nil {
		http.Error(w, "Could not generate token", http.StatusInternalServerError)
		return
	}

	// Return the token
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(TokenResponse{
		AccessToken: tokenString,
		TokenType:   "bearer",
	})
}

// WebSocket handler
func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Error upgrading to WebSocket: %v", err)
		return
	}
	defer ws.Close()

	// Register new client
	connections[ws] = true

	for {
		_, _, err := ws.ReadMessage()
		if err != nil {
			delete(connections, ws)
			break
		}
	}
}

// Broadcast messages to all connected clients
func handleBroadcasts() {
	for {
		msg := <-broadcast
		for client := range connections {
			err := client.WriteJSON(msg)
			if err != nil {
				client.Close()
				delete(connections, client)
			}
		}
	}
}

// Trade order handlers
func getTradesHandler(w http.ResponseWriter, r *http.Request) {
	var trades []TradeOrder
	db.Find(&trades)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(trades)
}

func createTradeHandler(w http.ResponseWriter, r *http.Request) {
	var trade TradeOrder
	if err := json.NewDecoder(r.Body).Decode(&trade); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Set the user from the JWT token
	trade.User = r.Context().Value("username").(string)
	trade.Status = "running"

	result := db.Create(&trade)
	if result.Error != nil {
		http.Error(w, result.Error.Error(), http.StatusInternalServerError)
		return
	}

	// Broadcast the new trade to WebSocket clients
	broadcast <- Message{Type: "new_trade", Content: trade}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(trade)
}

func getTradeHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tradeID := vars["id"]

	var trade TradeOrder
	result := db.First(&trade, tradeID)
	if result.Error != nil {
		http.Error(w, "Trade not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(trade)
}

func updateTradeHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tradeID := vars["id"]

	var trade TradeOrder
	result := db.First(&trade, tradeID)
	if result.Error != nil {
		http.Error(w, "Trade not found", http.StatusNotFound)
		return
	}

	var updatedTrade TradeOrder
	if err := json.NewDecoder(r.Body).Decode(&updatedTrade); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Update only allowed fields
	trade.Status = updatedTrade.Status
	db.Save(&trade)

	// Broadcast the update to WebSocket clients
	broadcast <- Message{Type: "update_trade", Content: trade}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(trade)
}

// Symbol handlers
func getSymbolsHandler(w http.ResponseWriter, r *http.Request) {
	var symbols []Symbol
	db.Find(&symbols)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(symbols)
}

func createSymbolHandler(w http.ResponseWriter, r *http.Request) {
	var symbol Symbol
	if err := json.NewDecoder(r.Body).Decode(&symbol); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	result := db.Create(&symbol)
	if result.Error != nil {
		http.Error(w, result.Error.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(symbol)
}

// Broker token handlers
func getBrokerTokensHandler(w http.ResponseWriter, r *http.Request) {
	var tokens []BrokerToken
	db.Find(&tokens)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tokens)
}

func createBrokerTokenHandler(w http.ResponseWriter, r *http.Request) {
	var token BrokerToken
	if err := json.NewDecoder(r.Body).Decode(&token); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Set the user from the JWT token
	token.User = r.Context().Value("username").(string)
	token.CreatedAt = time.Now()

	result := db.Create(&token)
	if result.Error != nil {
		http.Error(w, result.Error.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(token)
}

// updateBrokerTokenHandler handles the update of broker tokens that expire in 1 month
func updateBrokerTokenHandler(w http.ResponseWriter, r *http.Request) {
	// Extract token ID from URL parameters
	vars := mux.Vars(r)
	tokenID, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid token ID", http.StatusBadRequest)
		return
	}

	// Get the existing token
	var existingToken BrokerToken
	result := db.First(&existingToken, tokenID)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			http.Error(w, "Token not found", http.StatusNotFound)
		} else {
			http.Error(w, result.Error.Error(), http.StatusInternalServerError)
		}
		return
	}

	// Verify the user owns this token
	username := r.Context().Value("username").(string)
	if existingToken.User != username {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse the updated token data
	var updatedToken BrokerToken
	if err := json.NewDecoder(r.Body).Decode(&updatedToken); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Update only the token field and reset CreatedAt
	existingToken.Token = updatedToken.Token
	existingToken.CreatedAt = time.Now()

	// Save the updated token
	result = db.Save(&existingToken)
	if result.Error != nil {
		http.Error(w, result.Error.Error(), http.StatusInternalServerError)
		return
	}

	// Return the updated token
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(existingToken)
}

// User handlers
func getUsersHandler(w http.ResponseWriter, r *http.Request) {
	var users []User
	db.Select("id, username, email, role").Find(&users) // Exclude password field

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}

func createUserHandler(w http.ResponseWriter, r *http.Request) {
	var user User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Could not hash password", http.StatusInternalServerError)
		return
	}
	user.Password = string(hashedPassword)

	result := db.Create(&user)
	if result.Error != nil {
		http.Error(w, result.Error.Error(), http.StatusInternalServerError)
		return
	}

	// Don't return the password
	user.Password = ""

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

// ServerConfig represents the configuration for different server types
type ServerConfig struct {
	IsMicro      bool   // Whether this is the micro instance (frontend API)
	WorkerHost   string // The host of the worker instance for long-running tasks
	WorkerUser   string // The user for SSH connections to the worker
	WorkerSSHKey string // Path to the SSH key for worker connections
}

// Initialize server configuration based on environment variables
func initServerConfig() ServerConfig {
	// Determine if this is the micro instance based on environment variable or hostname
	hostname, _ := os.Hostname()
	isMicro := os.Getenv("SERVER_ROLE") == "micro" || hostname == "instance-20250422-132526"

	// Get worker configuration
	workerHost := os.Getenv("WORKER_HOST")
	if workerHost == "" {
		workerHost = "instance-20250416-112838" // Default to the n1-highcpu-4 instance
	}

	workerUser := os.Getenv("WORKER_USER")
	if workerUser == "" {
		workerUser = "root" // Default user
	}

	workerSSHKey := os.Getenv("WORKER_SSH_KEY")
	if workerSSHKey == "" {
		// Default to SSH key in the app directory
		workerSSHKey = "/opt/trademicro/.ssh/worker_key"
	}

	return ServerConfig{
		IsMicro:      isMicro,
		WorkerHost:   workerHost,
		WorkerUser:   workerUser,
		WorkerSSHKey: workerSSHKey,
	}
}

// --- Family Member Handlers ---

func getFamilyMembersHandler(w http.ResponseWriter, r *http.Request) {
	username, ok := r.Context().Value("username").(string)
	if !ok || username == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	var user User
	if err := db.Where("username = ?", username).First(&user).Error; err != nil {
		http.Error(w, "User not found", http.StatusUnauthorized)
		return
	}
	var members []FamilyMember
	if err := db.Where("user_id = ?", user.ID).Find(&members).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(members)
}

func createFamilyMemberHandler(w http.ResponseWriter, r *http.Request) {
	username, ok := r.Context().Value("username").(string)
	if !ok || username == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	var user User
	if err := db.Where("username = ?", username).First(&user).Error; err != nil {
		http.Error(w, "User not found", http.StatusUnauthorized)
		return
	}
	var member FamilyMember
	if err := json.NewDecoder(r.Body).Decode(&member); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}
	member.UserID = user.ID
	member.CreatedAt = time.Now()
	member.UpdatedAt = time.Now()
	if err := db.Create(&member).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(member)
}

func updateFamilyMemberHandler(w http.ResponseWriter, r *http.Request) {
	username, ok := r.Context().Value("username").(string)
	if !ok || username == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	var user User
	if err := db.Where("username = ?", username).First(&user).Error; err != nil {
		http.Error(w, "User not found", http.StatusUnauthorized)
		return
	}
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}
	var member FamilyMember
	if err := db.First(&member, id).Error; err != nil {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
	if member.UserID != user.ID {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	if err := json.NewDecoder(r.Body).Decode(&member); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}
	member.UserID = user.ID // Ensure user cannot change ownership
	member.UpdatedAt = time.Now()
	if err := db.Save(&member).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(member)
}

func deleteFamilyMemberHandler(w http.ResponseWriter, r *http.Request) {
	username, ok := r.Context().Value("username").(string)
	if !ok || username == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	var user User
	if err := db.Where("username = ?", username).First(&user).Error; err != nil {
		http.Error(w, "User not found", http.StatusUnauthorized)
		return
	}
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}
	var member FamilyMember
	if err := db.First(&member, id).Error; err != nil {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
	if member.UserID != user.ID {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	if err := db.Delete(&FamilyMember{}, id).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "deleted"})
}

func main() {
	// Initialize server configuration
	config := initServerConfig()

	// Create a new router
	r := mux.NewRouter()

	// Add health check endpoint
	r.HandleFunc("/api/health", api.HealthHandler).Methods("GET")

	// Serve the web application
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("web/static"))))

	// Set up CORS
	corsMiddleware := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"}, // Allow all origins for API access
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
	})

	// Apply CORS middleware
	handler := corsMiddleware.Handler(r)

	// WebSocket route
	r.HandleFunc("/ws", handleWebSocket)

	// Add public endpoints directly to the root router (no authentication required)
	r.HandleFunc("/api/login", loginHandler).Methods("POST")
	r.HandleFunc("/api/health", api.HealthHandler).Methods("GET")

	// Create the API router for authenticated endpoints
	apiRouter := r.PathPrefix("/api").Subrouter()

	// Create a subrouter for authenticated endpoints
	authRouter := apiRouter.PathPrefix("").Subrouter()
	authRouter.Use(authMiddleware)

	// Trade routes
	authRouter.HandleFunc("/trades", getTradesHandler).Methods("GET")
	authRouter.HandleFunc("/trades", createTradeHandler).Methods("POST")
	authRouter.HandleFunc("/trades/{id}", getTradeHandler).Methods("GET")
	authRouter.HandleFunc("/trades/{id}", updateTradeHandler).Methods("PUT")

	// Symbol routes
	authRouter.HandleFunc("/symbols", getSymbolsHandler).Methods("GET")
	authRouter.HandleFunc("/symbols", createSymbolHandler).Methods("POST")

	// Broker token routes
	authRouter.HandleFunc("/broker-tokens", getBrokerTokensHandler).Methods("GET")
	authRouter.HandleFunc("/broker-tokens", createBrokerTokenHandler).Methods("POST")
	authRouter.HandleFunc("/broker-tokens/{id}", updateBrokerTokenHandler).Methods("PUT")

	// User routes
	authRouter.HandleFunc("/users", getUsersHandler).Methods("GET")
	authRouter.HandleFunc("/users", createUserHandler).Methods("POST")

	// Family member sync routes
	authRouter.HandleFunc("/family-members", getFamilyMembersHandler).Methods("GET")
	authRouter.HandleFunc("/family-members", createFamilyMemberHandler).Methods("POST")
	authRouter.HandleFunc("/family-members/{id}", updateFamilyMemberHandler).Methods("PUT")
	authRouter.HandleFunc("/family-members/{id}", deleteFamilyMemberHandler).Methods("DELETE")

	// Initialize task management if this is the micro instance
	if config.IsMicro {
		log.Println("Initializing as FRONTEND micro instance with task management capabilities")

		// Create task handler with worker configuration
		taskHandler := api.NewTaskHandler()

		// Register task management routes
		taskHandler.RegisterRoutes(authRouter)

		// Initialize task cleanup
		taskHandler.InitTaskCleanup()

		log.Printf("Task management initialized with worker: %s@%s", config.WorkerUser, config.WorkerHost)
	} else {
		log.Println("Initializing as WORKER instance for handling long-running tasks")
	}

	// Catch-all handler for serving the SPA
	r.PathPrefix("/").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// For API requests, let the router handle them
		if strings.HasPrefix(r.URL.Path, "/api/") {
			http.NotFound(w, r)
			return
		}

		// For all other requests, serve the index.html file
		http.ServeFile(w, r, "web/index.html")
	})

	// Start the WebSocket broadcast handler
	go handleBroadcasts()

	// Start the server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}

	log.Printf("Server starting on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, handler))
}
