package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	"github.com/vikasavnish/trademicro/internal/services"
	"gorm.io/gorm"
)

type FavouredSymbolHandler struct {
	favouredSymbolService *services.FavouredSymbolService
	db                    *gorm.DB
}

func NewFavouredSymbolHandler(favouredSymbolService *services.FavouredSymbolService, db *gorm.DB) *FavouredSymbolHandler {
	return &FavouredSymbolHandler{
		favouredSymbolService: favouredSymbolService,
		db:                    db,
	}
}

func (h *FavouredSymbolHandler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/favoured-symbols", h.GetFavouredSymbols).Methods("GET")
	router.HandleFunc("/favoured-symbols", h.AddFavouredSymbol).Methods("POST")
	router.HandleFunc("/favoured-symbols/{id:[0-9]+}", h.GetFavouredSymbol).Methods("GET")
	router.HandleFunc("/favoured-symbols/{id:[0-9]+}", h.UpdateFavouredSymbol).Methods("PUT")
	router.HandleFunc("/favoured-symbols/{id:[0-9]+}", h.RemoveFavouredSymbol).Methods("DELETE")
}

// GetFavouredSymbols returns all favoured symbols for the current user
func (h *FavouredSymbolHandler) GetFavouredSymbols(w http.ResponseWriter, r *http.Request) {
	// Get user from context
	userID, err := getUserIDFromContext(r, h.db)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get favoured symbols
	favourites, err := h.favouredSymbolService.GetUserFavouredSymbols(userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"favouredSymbols": favourites,
	})
}

// AddFavouredSymbol adds a symbol to user's favoured list
func (h *FavouredSymbolHandler) AddFavouredSymbol(w http.ResponseWriter, r *http.Request) {
	// Get user from context
	userID, err := getUserIDFromContext(r, h.db)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse request
	var req struct {
		SymbolID uint   `json:"symbolId"`
		Notes    string `json:"notes"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Add to favourites
	favourite, err := h.favouredSymbolService.AddFavouredSymbol(userID, req.SymbolID, req.Notes)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(favourite)
}

// GetFavouredSymbol gets a specific favoured symbol
func (h *FavouredSymbolHandler) GetFavouredSymbol(w http.ResponseWriter, r *http.Request) {
	// Get user from context
	userID, err := getUserIDFromContext(r, h.db)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get ID from URL
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	// Get favoured symbol
	favourite, err := h.favouredSymbolService.GetFavouredSymbolByID(uint(id), userID)
	if err != nil {
		http.Error(w, "Symbol not found or access denied", http.StatusNotFound)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(favourite)
}

// UpdateFavouredSymbol updates a favoured symbol's notes
func (h *FavouredSymbolHandler) UpdateFavouredSymbol(w http.ResponseWriter, r *http.Request) {
	// Get user from context
	userID, err := getUserIDFromContext(r, h.db)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get ID from URL
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	// Parse request
	var req struct {
		Notes string `json:"notes"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Update favoured symbol
	favourite, err := h.favouredSymbolService.UpdateFavouredSymbol(uint(id), userID, req.Notes)
	if err != nil {
		http.Error(w, "Symbol not found or access denied", http.StatusNotFound)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(favourite)
}

// RemoveFavouredSymbol removes a symbol from user's favoured list
func (h *FavouredSymbolHandler) RemoveFavouredSymbol(w http.ResponseWriter, r *http.Request) {
	// Get user from context
	userID, err := getUserIDFromContext(r, h.db)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get ID from URL
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	// Get the favoured symbol first to get the symbolID
	favourite, err := h.favouredSymbolService.GetFavouredSymbolByID(uint(id), userID)
	if err != nil {
		http.Error(w, "Symbol not found or access denied", http.StatusNotFound)
		return
	}

	// Remove from favourites
	err = h.favouredSymbolService.RemoveFavouredSymbol(userID, favourite.SymbolID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Symbol removed from favourites",
	})
}

// Helper to get user ID from context
func getUserIDFromContext(r *http.Request, db *gorm.DB) (uint, error) {
	username, ok := r.Context().Value("username").(string)
	if !ok || username == "" {
		return 0, errors.New("unauthorized")
	}

	// Make sure this matches the actual user struct name in your models package
	var user struct {
		ID uint
	}
	if err := db.Table("users").Where("username = ?", username).First(&user).Error; err != nil {
		return 0, errors.New("user not found")
	}

	return user.ID, nil
}
