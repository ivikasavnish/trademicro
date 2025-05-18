// IMPORTANT: This file is being migrated to cmd/server/main.go
// This is kept only for backward compatibility
// Please use the modular version at cmd/server/main.go instead

package main

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/go-redis/redis"
	"github.com/golang-jwt/jwt"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/rs/cors"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/vikasavnish/trademicro/internal/api"
	"github.com/vikasavnish/trademicro/internal/handlers"
	"github.com/vikasavnish/trademicro/internal/services"
	"github.com/vikasavnish/trademicro/internal/websocket"
)

// StartTradeProcessRequest is the request body for starting a trade process
// script: one of dhanfeed_sync.py, updater.py, trade_log.py
// args: list of string arguments
// Example: {"script": "trade_log.py", "args": ["COALINDIA", "5", "--diff", ".1", "--zag", "5", "--type", "MTF", "sonam"]}
type StartTradeProcessRequest struct {
	Script string   `json:"script"`
	Args   []string `json:"args"`
}

// processActionHandler handles start/stop/resume actions for trading processes
func processActionHandler(w http.ResponseWriter, r *http.Request) {
	action := mux.Vars(r)["action"]
	var allowedActions = map[string]bool{"start": true, "stop": true, "resume": true}
	var allowedScripts = map[string]bool{"dhanfeed_sync.py": true, "updater.py": true, "trade_log.py": true}
	var forbiddenChars = []string{";", "&&", "|", "`", "$", ">", "<", "\n", "\r"}

	if !allowedActions[action] {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{"status": "error", "error": "invalid action"})
		return
	}
	var req StartTradeProcessRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{"status": "error", "error": "invalid request body"})
		return
	}
	if req.Script == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{"status": "error", "error": "script is required"})
		return
	}
	if !allowedScripts[req.Script] {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{"status": "error", "error": "script not allowed"})
		return
	}
	for _, arg := range req.Args {
		for _, ch := range forbiddenChars {
			if strings.Contains(arg, ch) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(map[string]interface{}{"status": "error", "error": "invalid argument detected"})
				return
			}
		}
	}
	cmdArgs := append([]string{action, req.Script}, req.Args...)
	cmd := exec.Command("python3", append([]string{"trade_manager.py"}, cmdArgs...)...)
	cmd.Dir = "."
	err := cmd.Start()
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{"status": "error", "error": err.Error()})
		return
	}
	go cmd.Wait()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"status": action + "ed", "pid": cmd.Process.Pid, "error": nil})
}

// processListHandler lists all managed trading processes
func processListHandler(w http.ResponseWriter, r *http.Request) {
	cmd := exec.Command("python3", "trade_manager.py", "list")
	cmd.Dir = "."
	output, err := cmd.Output()
	w.Header().Set("Content-Type", "application/json")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{"status": "error", "error": err.Error()})
		return
	}
	// The output is plain text, convert to JSON array of lines
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	json.NewEncoder(w).Encode(map[string]interface{}{"status": "ok", "processes": lines, "error": nil})
}

// startTradeProcessHandler starts a background trading process using trade_manager.py
func startTradeProcessHandler(w http.ResponseWriter, r *http.Request) {
	var allowedScripts = map[string]bool{
		"dhanfeed_sync.py": true,
		"updater.py":       true,
		"trade_log.py":     true,
	}
	var forbiddenChars = []string{";", "&&", "|", "`", "$", ">", "<", "\n", "\r"}

	var req StartTradeProcessRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{"status": "error", "error": "invalid request body"})
		return
	}
	if req.Script == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{"status": "error", "error": "script is required"})
		return
	}
	if !allowedScripts[req.Script] {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{"status": "error", "error": "script not allowed"})
		return
	}
	for _, arg := range req.Args {
		for _, ch := range forbiddenChars {
			if strings.Contains(arg, ch) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(map[string]interface{}{"status": "error", "error": "invalid argument detected"})
				return
			}
		}
	}
	cmdArgs := append([]string{"start", req.Script}, req.Args...)
	cmd := exec.Command("python3", append([]string{"trade_manager.py"}, cmdArgs...)...)
	cmd.Dir = "." // run in project root
	err := cmd.Start()
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{"status": "error", "error": err.Error()})
		return
	}
	go cmd.Wait() // Do not block
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"status": "started", "pid": cmd.Process.Pid, "error": nil})
}

// startTradeLogHandler starts trade_log.py with user-supplied args
func startTradeLogHandler(w http.ResponseWriter, r *http.Request) {
	var forbiddenChars = []string{";", "&&", "|", "`", "$", ">", "<", "\n", "\r"}

	var req StartTradeProcessRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{"status": "error", "error": "invalid request body"})
		return
	}
	if len(req.Args) == 0 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{"status": "error", "error": "args required for trade_log.py"})
		return
	}
	if req.Script != "" && req.Script != "trade_log.py" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{"status": "error", "error": "only trade_log.py is allowed"})
		return
	}
	for _, arg := range req.Args {
		for _, ch := range forbiddenChars {
			if strings.Contains(arg, ch) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(map[string]interface{}{"status": "error", "error": "invalid argument detected"})
				return
			}
		}
	}
	cmdArgs := append([]string{"start", "trade_log.py"}, req.Args...)
	cmd := exec.Command("python3", append([]string{"trade_manager.py"}, cmdArgs...)...)
	cmd.Dir = "."
	err := cmd.Start()
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{"status": "error", "error": err.Error()})
		return
	}
	go cmd.Wait()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"status": "started", "pid": cmd.Process.Pid, "error": nil})
}

// Global variables
var (
	db              *gorm.DB
	redisClient     *redis.Client
	connections     = make(map[*websocket.Conn]bool)
	broadcast       = make(chan Message)
	upgrader        = websocket.Upgrader{}
	jwtSecretKey    = []byte(os.Getenv("SECRET_KEY"))
	symbolUpdateJob *SymbolUpdateJob
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
	ID       uint   `gorm:"primaryKey" json:"id"`
	Symbol   string `json:"symbol" gorm:"uniqueIndex"`
	Name     string `json:"name"`
	Exchange string `json:"exchange"`
	Type     string `json:"type"`
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
	// ...existing code...

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

	// Auto migrations removed - using SQL migration files instead

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

// SymbolUpdateJob handles updating stock symbols once a day
type SymbolUpdateJob struct {
	ticker     *time.Ticker
	db         *gorm.DB
	dataSource SymbolDataSource
	running    bool
	stopCh     chan struct{}
}

// SymbolDataSource provides symbol data
type SymbolDataSource interface {
	FetchSymbols() ([]Symbol, error)
}

// DefaultSymbolDataSource implements SymbolDataSource
type DefaultSymbolDataSource struct{}

// FetchSymbols fetches stock symbols from an online CSV data source
func (ds *DefaultSymbolDataSource) FetchSymbols() ([]Symbol, error) {
	log.Println("Fetching symbols from online CSV source")

	// URL for Dhan API's CSV data
	csvURL := "https://images.dhan.co/api-data/api-scrip-master.csv"

	// Make HTTP request to fetch the CSV
	resp, err := http.Get(csvURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch CSV: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Parse CSV
	reader := csv.NewReader(resp.Body)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to parse CSV: %v", err)
	}

	// Ensure we have at least a header row and one data row
	if len(records) < 2 {
		return nil, fmt.Errorf("CSV has insufficient data")
	}

	// Get header row to map column indices
	headers := records[0]
	columnMap := make(map[string]int)
	for i, header := range headers {
		columnMap[header] = i
	}

	// Check required columns exist
	requiredColumns := []string{"SEM_TRADING_SYMBOL", "SM_SYMBOL_NAME", "SEM_EXM_EXCH_ID", "SEM_SEGMENT"}
	for _, col := range requiredColumns {
		if _, ok := columnMap[col]; !ok {
			return nil, fmt.Errorf("required column '%s' not found in CSV", col)
		}
	}

	// Convert CSV records to Symbol objects
	var symbols []Symbol
	for i := 1; i < len(records); i++ {
		row := records[i]
		// Skip if the row doesn't have enough columns
		if len(row) < len(headers) {
			continue
		}

		// Get values from specific columns
		var symbol Symbol

		// Map CSV columns to Symbol struct fields
		if idx, ok := columnMap["SEM_TRADING_SYMBOL"]; ok && idx < len(row) {
			symbol.Symbol = row[idx]
		}

		if idx, ok := columnMap["SM_SYMBOL_NAME"]; ok && idx < len(row) {
			symbol.Name = row[idx]
		}

		if idx, ok := columnMap["SEM_EXM_EXCH_ID"]; ok && idx < len(row) {
			symbol.Exchange = row[idx]
		}

		if idx, ok := columnMap["SEM_SEGMENT"]; ok && idx < len(row) {
			// Map segment to Type field
			segmentType := row[idx]
			if segmentType == "E" {
				symbol.Type = "EQ"
			} else if segmentType == "D" {
				symbol.Type = "IDX"
			} else {
				symbol.Type = segmentType
			}
		}

		// Skip if symbol is empty
		if symbol.Symbol == "" {
			continue
		}

		symbols = append(symbols, symbol)

		// Limit to 1000 symbols to avoid overwhelming the system
		if len(symbols) >= 1000 {
			log.Println("Limiting to 1000 symbols for performance")
			break
		}
	}

	log.Printf("Successfully parsed %d symbols from CSV", len(symbols))
	return symbols, nil
}

// NewSymbolUpdateJob creates a new symbol update job
func NewSymbolUpdateJob(db *gorm.DB) *SymbolUpdateJob {
	return &SymbolUpdateJob{
		db:         db,
		dataSource: &DefaultSymbolDataSource{},
		stopCh:     make(chan struct{}),
	}
}

// Start begins the symbol update job
func (job *SymbolUpdateJob) Start() {
	if job.running {
		log.Println("Symbol update job is already running")
		return
	}

	job.running = true

	// Run job immediately on startup
	job.UpdateSymbols()

	go func() {
		for {
			// Calculate time until next 8:45 AM
			now := time.Now()
			nextRun := time.Date(now.Year(), now.Month(), now.Day(), 8, 45, 0, 0, now.Location())
			if now.After(nextRun) {
				// If it's already past 8:45 AM today, schedule for tomorrow
				nextRun = nextRun.Add(24 * time.Hour)
			}

			waitDuration := nextRun.Sub(now)
			log.Printf("Symbol update job scheduled to run at 8:45 AM, waiting for %v", waitDuration)

			// Create timer for next run
			timer := time.NewTimer(waitDuration)

			select {
			case <-timer.C:
				// Run the update job
				job.UpdateSymbols()

				// Continue the loop to schedule the next run
				continue
			case <-job.stopCh:
				timer.Stop()
				job.running = false
				return
			}
		}
	}()

	log.Println("Symbol update job started, will run daily at 8:45 AM")
}

// Stop terminates the symbol update job
func (job *SymbolUpdateJob) Stop() {
	if !job.running {
		return
	}
	close(job.stopCh)
	job.running = false
	log.Println("Symbol update job stopped")
}

// UpdateSymbols fetches and updates symbols in the database
func (job *SymbolUpdateJob) UpdateSymbols() {
	log.Println("Running scheduled symbol update job")

	// Fetch symbols from data source
	symbols, err := job.dataSource.FetchSymbols()
	if err != nil {
		log.Printf("Error fetching symbols: %v", err)
		return
	}

	log.Printf("Fetched %d symbols from data source", len(symbols))

	// Begin a transaction
	tx := job.db.Begin()
	if tx.Error != nil {
		log.Printf("Error starting transaction: %v", tx.Error)
		return
	}

	// For each symbol
	for _, symbol := range symbols {
		// Try to find existing symbol
		var existingSymbol Symbol
		result := tx.Where("symbol = ?", symbol.Symbol).First(&existingSymbol)

		// If symbol exists, update it
		if result.Error == nil {
			existingSymbol.Name = symbol.Name
			existingSymbol.Exchange = symbol.Exchange
			existingSymbol.Type = symbol.Type

			if err := tx.Save(&existingSymbol).Error; err != nil {
				log.Printf("Error updating symbol %s: %v", symbol.Symbol, err)
				tx.Rollback()
				return
			}
		} else {
			// Create new symbol if it doesn't exist
			if err := tx.Create(&symbol).Error; err != nil {
				log.Printf("Error creating symbol %s: %v", symbol.Symbol, err)
				tx.Rollback()
				return
			}
		}
	}

	// Commit the transaction
	if err := tx.Commit().Error; err != nil {
		log.Printf("Error committing symbol updates: %v", err)
		tx.Rollback()
		return
	}

	log.Println("Symbol update completed successfully")

	// Broadcast update to websocket clients
	var symbolCount int64
	job.db.Model(&Symbol{}).Count(&symbolCount)
	broadcast <- Message{Type: "symbols_updated", Content: map[string]interface{}{
		"count": symbolCount,
		"time":  time.Now(),
	}}
}

// triggerSymbolUpdateHandler allows manual triggering of the symbol update job
func triggerSymbolUpdateHandler(w http.ResponseWriter, r *http.Request) {
	// Get the username from context for logging purposes
	username, _ := r.Context().Value("username").(string)

	log.Printf("Manual symbol update triggered by user: %s", username)

	// Access the global job instance and trigger an update
	go symbolUpdateJob.UpdateSymbols()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "success",
		"message": "Symbol update job triggered successfully",
	})
}

// startTradeSystemHandler starts the trading system on the worker machine
func startTradeSystemHandler(w http.ResponseWriter, r *http.Request) {
	// This is a placeholder for the actual implementation
	log.Println("Starting trade system...")

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "success",
		"message": "Trade system start command issued",
	})
}

// stopTradeSystemHandler stops the trading system on the worker machine
func stopTradeSystemHandler(w http.ResponseWriter, r *http.Request) {
	// This is a placeholder for the actual implementation
	log.Println("Stopping trade system...")

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "success",
		"message": "Trade system stop command issued",
	})
}

// listGCEInstancesHandler is a debug handler to list GCE instances
func listGCEInstancesHandler(w http.ResponseWriter, r *http.Request) {
	// This is a placeholder for the actual implementation
	log.Println("Listing GCE instances (debug)...")

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "success",
		"instances": []map[string]string{
			{"name": "instance-1", "status": "RUNNING"},
			{"name": "instance-2", "status": "STOPPED"},
		},
	})
}

// Additional types

// DhanHQInstrumentFetchConfig represents options for fetching instruments from DhanHQ API
type DhanHQInstrumentFetchConfig struct {
	Mode            string `json:"mode"`            // "compact" or "detailed"
	ExchangeSegment string `json:"exchangeSegment"` // Optional segment to filter by
	SaveToFile      bool   `json:"saveToFile"`      // Whether to save the CSV to file
	BatchSize       int    `json:"batchSize"`       // Batch size for DB operations
}

// fetchDhanHQInstrumentsHandler fetches instruments from DhanHQ API and adds them to the database
func fetchDhanHQInstrumentsHandler(w http.ResponseWriter, r *http.Request) {
	// Default configuration
	config := DhanHQInstrumentFetchConfig{
		Mode:       "compact",
		SaveToFile: true,
		BatchSize:  500, // Default batch size for DB operations
	}

	// Parse request body if provided
	if r.Body != nil && r.ContentLength > 0 {
		if err := json.NewDecoder(r.Body).Decode(&config); err != nil {
			http.Error(w, fmt.Sprintf("Invalid request: %v", err), http.StatusBadRequest)
			return
		}
	}

	log.Printf("Fetching DhanHQ instruments using mode: %s", config.Mode)

	// Get current username for auditing
	username := "system"
	if ctxUsername, ok := r.Context().Value("username").(string); ok && ctxUsername != "" {
		username = ctxUsername
	}

	// Use Python script for batch processing
	scriptPath := "./update_instruments.py"

	// Check if script exists
	if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
		http.Error(w, "Instrument update script not found", http.StatusInternalServerError)
		log.Printf("Error: update_instruments.py script not found")
		return
	}

	// Build command with appropriate arguments
	args := []string{
		"--mode", config.Mode,
		"--batch-size", fmt.Sprintf("%d", config.BatchSize),
	}

	// Add exchange segment filter if specified
	if config.ExchangeSegment != "" {
		args = append(args, "--exchange-segment", config.ExchangeSegment)
	}

	// Configure CSV output
	outputFile := fmt.Sprintf("dhan_instruments_%s_%s.csv",
		config.Mode,
		time.Now().Format("20060102_150405"))
	args = append(args, "--output", outputFile)

	// Log the operation
	log.Printf("Starting instrument update process with user: %s, mode: %s", username, config.Mode)

	cmd := exec.Command("python3", append([]string{scriptPath}, args...)...)

	// Execute command and get output
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("Error fetching instruments: %v, Output: %s", err, string(output))
		http.Error(w, fmt.Sprintf("Failed to fetch instruments: %v", err), http.StatusInternalServerError)
		return
	}

	// Check if the output file exists
	if _, err := os.Stat(outputFile); os.IsNotExist(err) {
		http.Error(w, "Failed to generate instrument data file", http.StatusInternalServerError)
		log.Printf("Error: Output file not created: %s", outputFile)
		return
	}

	// Process the file in batches
	log.Printf("Processing instrument data from file: %s", outputFile)

	// Open the CSV file
	file, err := os.Open(outputFile)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to open data file: %v", err), http.StatusInternalServerError)
		return
	}
	defer file.Close()

	// Parse CSV data
	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		log.Printf("Failed to parse CSV data: %v", err)
		http.Error(w, fmt.Sprintf("Failed to parse instrument data: %v", err), http.StatusInternalServerError)
		return
	}

	// Ensure we have at least a header row
	if len(records) < 2 {
		log.Printf("Insufficient CSV data received: %v rows", len(records))
		http.Error(w, "Insufficient instrument data received", http.StatusInternalServerError)
		return
	}

	// Get header row to map column indices
	headers := records[0]
	columnMap := make(map[string]int)
	for i, header := range headers {
		columnMap[header] = i
	}

	// Track counts for the whole process
	totalCreated := 0
	totalUpdated := 0
	totalSkipped := 0

	// Calculate total number of batches
	dataRows := len(records) - 1 // Exclude header row
	batchSize := config.BatchSize
	if batchSize <= 0 {
		batchSize = 500 // Fallback if invalid batch size provided
	}

	totalBatches := (dataRows + batchSize - 1) / batchSize // Ceiling division
	log.Printf("Processing %d instruments in %d batches of size %d", dataRows, totalBatches, batchSize)

	// Process in batches to avoid overwhelming the database
	for batchNum := 0; batchNum < totalBatches; batchNum++ {
		startIdx := batchNum*batchSize + 1 // +1 to skip header
		endIdx := min((batchNum+1)*batchSize+1, len(records))

		log.Printf("Processing batch %d/%d (rows %d-%d)", batchNum+1, totalBatches, startIdx, endIdx-1)

		// Begin database transaction for this batch
		tx := db.Begin()
		if tx.Error != nil {
			http.Error(w, fmt.Sprintf("Database error: %v", tx.Error), http.StatusInternalServerError)
			return
		}

		// Track counts for this batch
		created := 0
		updated := 0
		skipped := 0

		// Process each row in this batch
		for i := startIdx; i < endIdx; i++ {
			row := records[i]

			// Skip if row is too short
			if len(row) < len(headers) {
				skipped++
				continue
			}

			// Create symbol object
			var symbol Symbol
			var symbolCode, symbolName, exchange, segment string

			// Try to extract data using different column names based on format
			// For compact format
			if idx, ok := columnMap["SYMBOL"]; ok && idx < len(row) {
				symbolCode = row[idx]
			}

			// For detailed format
			if symbolCode == "" {
				if idx, ok := columnMap["SEM_TRADING_SYMBOL"]; ok && idx < len(row) {
					symbolCode = row[idx]
				}
			}

			// Skip if no symbol code found
			if symbolCode == "" {
				skipped++
				continue
			}

			// Get symbol name
			if idx, ok := columnMap["SYMBOL_NAME"]; ok && idx < len(row) {
				symbolName = row[idx]
			} else if idx, ok := columnMap["SEM_CUSTOM_SYMBOL"]; ok && idx < len(row) {
				symbolName = row[idx]
			}

			// Get exchange
			if idx, ok := columnMap["EXCH_ID"]; ok && idx < len(row) {
				exchange = row[idx]
			} else if idx, ok := columnMap["SEM_EXM_EXCH_ID"]; ok && idx < len(row) {
				exchange = row[idx]
			}

			// Get segment
			if idx, ok := columnMap["SEGMENT"]; ok && idx < len(row) {
				segment = row[idx]
			} else if idx, ok := columnMap["SEM_SEGMENT"]; ok && idx < len(row) {
				segment = row[idx]
			}

			// Determine type based on segment
			segmentType := "OTHER"
			if segment == "E" {
				segmentType = "EQ"
			} else if segment == "D" {
				segmentType = "IDX"
			} else if segment == "C" {
				segmentType = "CUR"
			} else if segment == "M" {
				segmentType = "COMM"
			}

			// Set symbol fields
			symbol.Symbol = symbolCode
			symbol.Name = symbolName
			symbol.Exchange = exchange
			symbol.Type = segmentType

			// Skip if there's a segment filter and it doesn't match
			if config.ExchangeSegment != "" {
				// Check if the exchange segment matches
				fullSegment := exchange + "_" + segmentType
				if !strings.Contains(strings.ToLower(fullSegment), strings.ToLower(config.ExchangeSegment)) {
					skipped++
					continue
				}
			}

			// Try to find existing symbol
			var existingSymbol Symbol
			result := tx.Where("symbol = ?", symbol.Symbol).First(&existingSymbol)

			// Update or create
			if result.Error == nil {
				existingSymbol.Name = symbol.Name
				existingSymbol.Exchange = symbol.Exchange
				existingSymbol.Type = symbol.Type

				if err := tx.Save(&existingSymbol).Error; err != nil {
					log.Printf("Error updating symbol %s: %v", symbol.Symbol, err)
					tx.Rollback()
					http.Error(w, fmt.Sprintf("Database error: %v", err), http.StatusInternalServerError)
					return
				}
				updated++
			} else {
				if err := tx.Create(&symbol).Error; err != nil {
					log.Printf("Error creating symbol %s: %v", symbol.Symbol, err)
					tx.Rollback()
					http.Error(w, fmt.Sprintf("Database error: %v", err), http.StatusInternalServerError)
					return
				}
				created++
			}
		}

		// Commit this batch's transaction
		if err := tx.Commit().Error; err != nil {
			http.Error(w, fmt.Sprintf("Failed to commit batch %d: %v", batchNum+1, err), http.StatusInternalServerError)
			return
		}

		// Update totals
		totalCreated += created
		totalUpdated += updated
		totalSkipped += skipped

		log.Printf("Batch %d/%d completed: created=%d, updated=%d, skipped=%d",
			batchNum+1, totalBatches, created, updated, skipped)
	}

	// Send success response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":        "success",
		"created":       totalCreated,
		"updated":       totalUpdated,
		"skipped":       totalSkipped,
		"total":         totalCreated + totalUpdated + totalSkipped,
		"total_batches": totalBatches,
		"batch_size":    batchSize,
		"time":          time.Now().Format(time.RFC3339),
		"user":          username,
		"mode":          config.Mode,
	})

	// Broadcast update to websocket clients
	broadcast <- Message{
		Type: "instruments_updated",
		Content: map[string]interface{}{
			"created":     totalCreated,
			"updated":     totalUpdated,
			"total":       totalCreated + totalUpdated,
			"time":        time.Now(),
			"mode":        config.Mode,
			"user":        username,
			"batch_count": totalBatches,
		},
	}

	// Clean up temporary file if needed
	if !config.SaveToFile {
		os.Remove(outputFile)
	}
}

// fetchDhanHQSegmentsHandler returns the available exchange segments from DhanHQ
func fetchDhanHQSegmentsHandler(w http.ResponseWriter, r *http.Request) {
	// Define the segments based on DhanHQ documentation
	segments := []map[string]interface{}{
		{"value": "NSE_EQ", "label": "NSE Equity"},
		{"value": "BSE_EQ", "label": "BSE Equity"},
		{"value": "NSE_FNO", "label": "NSE Futures & Options"},
		{"value": "BSE_FNO", "label": "BSE Futures & Options"},
		{"value": "MCX_COMM", "label": "MCX Commodities"},
		{"value": "NSE_CURRENCY", "label": "NSE Currency"},
		{"value": "BSE_CURRENCY", "label": "BSE Currency"},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(segments)
}

// DhanHQInstrumentUpdater handles updating instruments from DhanHQ API on a schedule
type DhanHQInstrumentUpdater struct {
	db        *gorm.DB
	ticker    *time.Ticker
	running   bool
	stopCh    chan struct{}
	mode      string
	batchSize int
}

// NewDhanHQInstrumentUpdater creates a new instrument updater job
func NewDhanHQInstrumentUpdater(db *gorm.DB) *DhanHQInstrumentUpdater {
	return &DhanHQInstrumentUpdater{
		db:        db,
		stopCh:    make(chan struct{}),
		mode:      "compact",
		batchSize: 500,
	}
}

// Start begins the instrument update job
func (job *DhanHQInstrumentUpdater) Start() {
	if job.running {
		log.Println("DhanHQ instrument update job is already running")
		return
	}

	job.running = true

	// Run the job once at startup
	go job.UpdateInstruments()

	go func() {
		for {
			// Calculate time until next 3:00 AM - different time than symbol update
			now := time.Now()
			nextRun := time.Date(now.Year(), now.Month(), now.Day(), 3, 0, 0, 0, now.Location())
			if now.After(nextRun) {
				// If it's already past 3:00 AM today, schedule for tomorrow
				nextRun = nextRun.Add(24 * time.Hour)
			}

			waitDuration := nextRun.Sub(now)
			log.Printf("DhanHQ instrument update job scheduled to run at 3:00 AM, waiting for %v", waitDuration)

			// Create timer for next run
			timer := time.NewTimer(waitDuration)

			select {
			case <-timer.C:
				// Run the update job
				job.UpdateInstruments()

				// Continue the loop to schedule the next run
				continue
			case <-job.stopCh:
				timer.Stop()
				job.running = false
				return
			}
		}
	}()

	log.Println("DhanHQ instrument update job started, will run daily at 3:00 AM")
}

// Stop terminates the instrument update job
func (job *DhanHQInstrumentUpdater) Stop() {
	if !job.running {
		return
	}
	close(job.stopCh)
	job.running = false
	log.Println("DhanHQ instrument update job stopped")
}

// UpdateInstruments fetches and updates instruments from DhanHQ API
func (job *DhanHQInstrumentUpdater) UpdateInstruments() {
	log.Println("Running scheduled DhanHQ instrument update job")

	outputFile := fmt.Sprintf("dhan_instruments_scheduled_%s.csv",
		time.Now().Format("20060102"))

	// Build command with appropriate arguments
	args := []string{
		"--mode", job.mode,
		"--batch-size", fmt.Sprintf("%d", job.batchSize),
		"--output", outputFile,
	}

	// Execute Python script to fetch instruments
	cmd := exec.Command("python3", append([]string{"./update_instruments.py"}, args...)...)

	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("Error in scheduled instrument update: %v, Output: %s", err, string(output))
		return
	}

	log.Printf("Instrument update script output: %s", string(output))

	// Process the file in batches (similar to the handler)
	file, err := os.Open(outputFile)
	if err != nil {
		log.Printf("Failed to open data file: %v", err)
		return
	}
	defer file.Close()

	// Parse CSV data
	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		log.Printf("Failed to parse CSV data: %v", err)
		return
	}

	// Ensure we have data
	if len(records) < 2 {
		log.Printf("Insufficient CSV data received: %v rows", len(records))
		return
	}

	// Get header row and create column map
	headers := records[0]
	columnMap := make(map[string]int)
	for i, header := range headers {
		columnMap[header] = i
	}

	// Calculate batches
	dataRows := len(records) - 1
	totalBatches := (dataRows + job.batchSize - 1) / job.batchSize
	log.Printf("Processing %d instruments in %d batches", dataRows, totalBatches)

	// Track counts
	totalCreated := 0
	totalUpdated := 0
	totalSkipped := 0

	// Process batches
	for batchNum := 0; batchNum < totalBatches; batchNum++ {
		startIdx := batchNum*job.batchSize + 1 // Skip header
		endIdx := min((batchNum+1)*job.batchSize+1, len(records))

		log.Printf("Processing batch %d/%d (rows %d-%d)", batchNum+1, totalBatches, startIdx, endIdx-1)

		// Begin transaction
		tx := job.db.Begin()
		if tx.Error != nil {
			log.Printf("Database error: %v", tx.Error)
			return
		}

		// Process rows
		created, updated, skipped := 0, 0, 0

		for i := startIdx; i < endIdx; i++ {
			row := records[i]

			// Skip if row is too short
			if len(row) < len(headers) {
				skipped++
				continue
			}

			// Extract symbol data using a helper function
			symbol, ok := job.extractSymbolFromCSV(row, columnMap)
			if !ok {
				skipped++
				continue
			}

			// Try to find existing symbol
			var existingSymbol Symbol
			result := tx.Where("symbol = ?", symbol.Symbol).First(&existingSymbol)

			// Update or create
			if result.Error == nil {
				existingSymbol.Name = symbol.Name
				existingSymbol.Exchange = symbol.Exchange
				existingSymbol.Type = symbol.Type

				if err := tx.Save(&existingSymbol).Error; err != nil {
					log.Printf("Error updating symbol %s: %v", symbol.Symbol, err)
					tx.Rollback()
					return
				}
				updated++
			} else {
				if err := tx.Create(&symbol).Error; err != nil {
					log.Printf("Error creating symbol %s: %v", symbol.Symbol, err)
					tx.Rollback()
					return
				}
				created++
			}
		}

		// Commit transaction
		if err := tx.Commit().Error; err != nil {
			log.Printf("Failed to commit batch %d: %v", batchNum+1, err)
			return
		}

		// Update totals
		totalCreated += created
		totalUpdated += updated
		totalSkipped += skipped

		log.Printf("Batch %d completed: created=%d, updated=%d, skipped=%d",
			batchNum+1, created, updated, skipped)
	}

	// Cleanup
	os.Remove(outputFile)

	log.Printf("DhanHQ instrument update completed: created=%d, updated=%d, skipped=%d",
		totalCreated, totalUpdated, totalSkipped)

	// Broadcast update to websocket clients
	broadcast <- Message{
		Type: "instruments_updated",
		Content: map[string]interface{}{
			"created":   totalCreated,
			"updated":   totalUpdated,
			"total":     totalCreated + totalUpdated,
			"time":      time.Now(),
			"mode":      job.mode,
			"scheduled": true,
		},
	}
}

// extractSymbolFromCSV extracts symbol data from a CSV row
func (job *DhanHQInstrumentUpdater) extractSymbolFromCSV(row []string, columnMap map[string]int) (Symbol, bool) {
	var symbol Symbol
	var symbolCode, symbolName, exchange, segment string

	// Try to extract data based on different CSV formats
	// For compact format
	if idx, ok := columnMap["SYMBOL"]; ok && idx < len(row) {
		symbolCode = row[idx]
	}

	// For detailed format
	if symbolCode == "" {
		if idx, ok := columnMap["SEM_TRADING_SYMBOL"]; ok && idx < len(row) {
			symbolCode = row[idx]
		}
	}

	// Skip if no symbol code found
	if symbolCode == "" {
		return symbol, false
	}

	// Get symbol name
	if idx, ok := columnMap["SYMBOL_NAME"]; ok && idx < len(row) {
		symbolName = row[idx]
	} else if idx, ok := columnMap["SEM_CUSTOM_SYMBOL"]; ok && idx < len(row) {
		symbolName = row[idx]
	}

	// Get exchange
	if idx, ok := columnMap["EXCH_ID"]; ok && idx < len(row) {
		exchange = row[idx]
	} else if idx, ok := columnMap["SEM_EXM_EXCH_ID"]; ok && idx < len(row) {
		exchange = row[idx]
	}

	// Get segment
	if idx, ok := columnMap["SEGMENT"]; ok && idx < len(row) {
		segment = row[idx]
	} else if idx, ok := columnMap["SEM_SEGMENT"]; ok && idx < len(row) {
		segment = row[idx]
	}

	// Determine type based on segment
	segmentType := "OTHER"
	if segment == "E" {
		segmentType = "EQ"
	} else if segment == "D" {
		segmentType = "IDX"
	} else if segment == "C" {
		segmentType = "CUR"
	} else if segment == "M" {
		segmentType = "COMM"
	}

	// Set symbol fields
	symbol.Symbol = symbolCode
	symbol.Name = symbolName
	symbol.Exchange = exchange
	symbol.Type = segmentType

	return symbol, true
}

func main() {
	// Initialize server configuration
	config := initServerConfig()

	// Create a new router
	r := mux.NewRouter()

	// Initialize symbol update job
	symbolUpdateJob = NewSymbolUpdateJob(db)
	symbolUpdateJob.Start()

	// Initialize DhanHQ instrument update job
	dhanInstrumentJob := NewDhanHQInstrumentUpdater(db)
	dhanInstrumentJob.Start()

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

	// Trade system control endpoints (big machine)
	authRouter.HandleFunc("/start_trade_system", startTradeSystemHandler).Methods("POST")
	authRouter.HandleFunc("/stop_trade_system", stopTradeSystemHandler).Methods("POST")

	// Trade process management endpoints
	authRouter.HandleFunc("/start_trade_process", startTradeProcessHandler).Methods("POST")
	authRouter.HandleFunc("/start_trade_log", startTradeLogHandler).Methods("POST")

	authRouter.HandleFunc("/process/{action:start|stop|resume}", processActionHandler).Methods("POST")
	authRouter.HandleFunc("/process/list", processListHandler).Methods("GET")

	// DEBUG: List GCE instances in zone
	authRouter.HandleFunc("/debug/list_instances", listGCEInstancesHandler).Methods("GET")

	// Symbol routes
	authRouter.HandleFunc("/symbols", getSymbolsHandler).Methods("GET")
	authRouter.HandleFunc("/symbols", createSymbolHandler).Methods("POST")
	authRouter.HandleFunc("/symbols/update", triggerSymbolUpdateHandler).Methods("POST")

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

	// Initialize symbol service and handler
	symbolService := services.NewSymbolService(db)
	symbolHandler := handlers.NewSymbolHandler(symbolService)

	// Register symbol routes
	symbolHandler.RegisterRoutes(authRouter)

	// Initialize DhanHQ instrument service and handler
	instrumentService := services.NewInstrumentService(db)
	instrumentHandler := handlers.NewInstrumentHandler(instrumentService)

	// Register instrument routes
	instrumentHandler.RegisterRoutes(authRouter)

	// Start scheduled instrument updates
	instrumentUpdateStopCh := instrumentService.StartScheduledUpdates(services.ScheduledUpdateOptions{
		Mode:         services.CompactMode,
		BatchSize:    500,
		HourOfDay:    3, // Run at 3 AM
		MinuteOfHour: 0,
	})

	// Store stopCh for proper shutdown (not used yet but good practice)
	_ = instrumentUpdateStopCh

	// DhanHQ instrument routes
	authRouter.HandleFunc("/dhanhq/instruments", fetchDhanHQInstrumentsHandler).Methods("POST")
	authRouter.HandleFunc("/dhanhq/segments", fetchDhanHQSegmentsHandler).Methods("GET")

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
