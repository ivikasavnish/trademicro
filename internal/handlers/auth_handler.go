package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/vikasavnish/trademicro/internal/models"
	"github.com/vikasavnish/trademicro/internal/services"
)

// AuthHandler handles authentication requests
type AuthHandler struct {
	authService  services.AuthService
	jwtSecretKey []byte
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(authService services.AuthService, jwtSecretKey []byte) *AuthHandler {
	return &AuthHandler{
		authService:  authService,
		jwtSecretKey: jwtSecretKey,
	}
}

// Login handles user login and returns a JWT token
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var loginReq models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&loginReq); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Authenticate the user
	user, err := h.authService.Authenticate(loginReq.Username, loginReq.Password)
	if err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	// Generate token
	tokenString, err := h.authService.GenerateToken(user, h.jwtSecretKey)
	if err != nil {
		http.Error(w, "Could not generate token", http.StatusInternalServerError)
		return
	}

	// Return the token
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(models.TokenResponse{
		AccessToken: tokenString,
		TokenType:   "bearer",
	})
}
