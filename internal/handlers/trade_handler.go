package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	"github.com/vikasavnish/trademicro/internal/models"
	"github.com/vikasavnish/trademicro/internal/services"
	"github.com/vikasavnish/trademicro/internal/websocket"
)

// TradeHandler handles trade-related requests
type TradeHandler struct {
	tradeService services.TradeService
	wsHub        *websocket.Hub
}

// NewTradeHandler creates a new trade handler
func NewTradeHandler(tradeService services.TradeService, wsHub *websocket.Hub) *TradeHandler {
	return &TradeHandler{
		tradeService: tradeService,
		wsHub:        wsHub,
	}
}

// RegisterRoutes registers trade routes
func (h *TradeHandler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/trades", h.GetTrades).Methods("GET")
	router.HandleFunc("/trades", h.CreateTrade).Methods("POST")
	router.HandleFunc("/trades/{id}", h.GetTrade).Methods("GET")
	router.HandleFunc("/trades/{id}", h.UpdateTrade).Methods("PUT")
}

// GetTrades returns all trades
func (h *TradeHandler) GetTrades(w http.ResponseWriter, r *http.Request) {
	trades, err := h.tradeService.GetTrades()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(trades)
}

// CreateTrade creates a new trade
func (h *TradeHandler) CreateTrade(w http.ResponseWriter, r *http.Request) {
	var trade models.TradeOrder
	if err := json.NewDecoder(r.Body).Decode(&trade); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Set the user from the JWT token
	trade.User = r.Context().Value("username").(string)
	trade.Status = "running"

	createdTrade, err := h.tradeService.CreateTrade(trade)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Broadcast the new trade to WebSocket clients
	h.wsHub.Broadcast(models.Message{Type: "new_trade", Content: createdTrade})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(trade)
}

// GetTrade returns a specific trade by ID
func (h *TradeHandler) GetTrade(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tradeIDStr := vars["id"]

	// Convert string ID to uint
	tradeID, err := strconv.ParseUint(tradeIDStr, 10, 32)
	if err != nil {
		http.Error(w, "Invalid trade ID", http.StatusBadRequest)
		return
	}

	trade, err := h.tradeService.GetTradeByID(uint(tradeID))
	if err != nil {
		http.Error(w, "Trade not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(trade)
}

// UpdateTrade updates a trade's status
func (h *TradeHandler) UpdateTrade(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tradeIDStr := vars["id"]

	// Convert string ID to uint
	tradeID, err := strconv.ParseUint(tradeIDStr, 10, 32)
	if err != nil {
		http.Error(w, "Invalid trade ID", http.StatusBadRequest)
		return
	}

	var updatedTrade models.TradeOrder
	if err := json.NewDecoder(r.Body).Decode(&updatedTrade); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	trade, err := h.tradeService.UpdateTrade(uint(tradeID), updatedTrade)
	if err != nil {
		http.Error(w, "Failed to update trade", http.StatusInternalServerError)
		return
	}

	// Broadcast the update to WebSocket clients
	h.wsHub.Broadcast(models.Message{Type: "update_trade", Content: trade})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(trade)
}
