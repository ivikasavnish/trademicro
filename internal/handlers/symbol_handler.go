package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/vikasavnish/trademicro/internal/models"
	"github.com/vikasavnish/trademicro/internal/services"
)

type SymbolHandler struct {
	symbolService *services.SymbolService
}

func NewSymbolHandler(symbolService *services.SymbolService) *SymbolHandler {
	return &SymbolHandler{
		symbolService: symbolService,
	}
}

func (h *SymbolHandler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/symbols", h.GetSymbols).Methods("GET")
	router.HandleFunc("/symbols/{id:[0-9]+}", h.GetSymbol).Methods("GET")
	router.HandleFunc("/symbols/code/{code}", h.GetSymbolByCode).Methods("GET")
	router.HandleFunc("/symbols", h.CreateSymbol).Methods("POST")
	router.HandleFunc("/symbols/{id:[0-9]+}", h.UpdateSymbol).Methods("PUT")
	router.HandleFunc("/symbols/{id:[0-9]+}", h.DeleteSymbol).Methods("DELETE")
	router.HandleFunc("/symbols/import/dhan", h.ImportFromDhan).Methods("POST")
	router.HandleFunc("/symbols/import/csv", h.ImportFromCSV).Methods("POST")
}

// GetSymbols returns a list of symbols with pagination and filtering
func (h *SymbolHandler) GetSymbols(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	query := r.URL.Query()
	limit, _ := strconv.Atoi(query.Get("limit"))
	offset, _ := strconv.Atoi(query.Get("offset"))

	// Create filter map
	filter := make(map[string]string)
	if exchange := query.Get("exchange"); exchange != "" {
		filter["exchange"] = exchange
	}
	if segment := query.Get("segment"); segment != "" {
		filter["segment"] = segment
	}
	if search := query.Get("search"); search != "" {
		filter["search"] = search
	}
	if active := query.Get("active"); active != "" {
		filter["active"] = active
	}

	// Get symbols
	symbols, count, err := h.symbolService.GetAllSymbols(filter, limit, offset)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Prepare response
	response := map[string]interface{}{
		"symbols": symbols,
		"count":   count,
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetSymbol returns a single symbol by ID
func (h *SymbolHandler) GetSymbol(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	symbol, err := h.symbolService.GetSymbolByID(uint(id))
	if err != nil {
		http.Error(w, "Symbol not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(symbol)
}

// GetSymbolByCode returns a single symbol by its code/symbol
func (h *SymbolHandler) GetSymbolByCode(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	code := vars["code"]

	symbol, err := h.symbolService.GetSymbolByCode(code)
	if err != nil {
		http.Error(w, "Symbol not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(symbol)
}

// CreateSymbol creates a new symbol
func (h *SymbolHandler) CreateSymbol(w http.ResponseWriter, r *http.Request) {
	var symbol models.Symbol
	if err := json.NewDecoder(r.Body).Decode(&symbol); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.symbolService.CreateSymbol(&symbol); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(symbol)
}

// UpdateSymbol updates an existing symbol
func (h *SymbolHandler) UpdateSymbol(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	// Get existing symbol
	existingSymbol, err := h.symbolService.GetSymbolByID(uint(id))
	if err != nil {
		http.Error(w, "Symbol not found", http.StatusNotFound)
		return
	}

	// Parse updated fields
	var updatedSymbol models.Symbol
	if err := json.NewDecoder(r.Body).Decode(&updatedSymbol); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Update fields but preserve ID
	updatedSymbol.ID = existingSymbol.ID

	if err := h.symbolService.UpdateSymbol(&updatedSymbol); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updatedSymbol)
}

// DeleteSymbol deletes a symbol
func (h *SymbolHandler) DeleteSymbol(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	if err := h.symbolService.DeleteSymbol(uint(id)); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Symbol deleted successfully",
	})
}

// ImportFromDhan imports symbols from Dhan API
func (h *SymbolHandler) ImportFromDhan(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Compact bool `json:"compact"`
	}

	// Parse request body
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		// Default to compact version if parsing fails
		req.Compact = true
	}

	count, err := h.symbolService.ImportSymbolsFromDhanAPI(req.Compact)
	if err != nil {
		http.Error(w, "Failed to import symbols: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":    "Symbols imported successfully",
		"count":      count,
		"source":     "Dhan API",
		"sourceType": req.Compact,
	})
}

// ImportFromCSV imports symbols from an uploaded CSV file
func (h *SymbolHandler) ImportFromCSV(w http.ResponseWriter, r *http.Request) {
	// Parse multipart form (max 10MB)
	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		http.Error(w, "Failed to parse form: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Get file from request
	file, _, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Failed to get file: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Import symbols from CSV
	count, err := h.symbolService.ImportSymbolsFromCSV(file)
	if err != nil {
		http.Error(w, "Failed to import symbols: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Symbols imported successfully",
		"count":   count,
		"source":  "CSV Upload",
	})
}
