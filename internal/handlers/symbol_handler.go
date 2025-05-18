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
	// Add autocomplete endpoint
	router.HandleFunc("/symbols/autocomplete/underlying", h.AutocompleteUnderlyingSymbols).Methods("GET")
}

// GetSymbols returns a list of symbols with pagination and filtering
func (h *SymbolHandler) GetSymbols(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	query := r.URL.Query()
	limit, _ := strconv.Atoi(query.Get("limit"))
	offset, _ := strconv.Atoi(query.Get("offset"))

	// Create filter map
	filter := make(map[string]string)

	// Basic filters
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

	// Instrument filters
	if instrType := query.Get("instrument_type"); instrType != "" {
		filter["instrument_type"] = instrType
	}
	if series := query.Get("series"); series != "" {
		filter["series"] = series
	}
	if instrument := query.Get("instrument"); instrument != "" {
		filter["instrument"] = instrument
	}
	if isin := query.Get("isin"); isin != "" {
		filter["isin"] = isin
	}
	if securityId := query.Get("security_id"); securityId != "" {
		filter["security_id"] = securityId
	}
	if underlyingSymbol := query.Get("underlying_symbol"); underlyingSymbol != "" {
		filter["underlying_symbol"] = underlyingSymbol
	}

	// Options filters
	if strikePrice := query.Get("strike_price"); strikePrice != "" {
		filter["strike_price"] = strikePrice
	}
	if optionType := query.Get("option_type"); optionType != "" {
		filter["option_type"] = optionType
	}
	if expiryFlag := query.Get("expiry_flag"); expiryFlag != "" {
		filter["expiry_flag"] = expiryFlag
	}

	// Margin filters
	if minMtf := query.Get("min_mtf"); minMtf != "" {
		filter["min_mtf"] = minMtf
	}
	if maxMtf := query.Get("max_mtf"); maxMtf != "" {
		filter["max_mtf"] = maxMtf
	}
	if lotSize := query.Get("lot_size"); lotSize != "" {
		filter["lot_size"] = lotSize
	}
	if minLotSize := query.Get("min_lot_size"); minLotSize != "" {
		filter["min_lot_size"] = minLotSize
	}
	if maxLotSize := query.Get("max_lot_size"); maxLotSize != "" {
		filter["max_lot_size"] = maxLotSize
	}

	// Margin requirement filters
	if minBuyCoMargin := query.Get("min_buy_co_margin"); minBuyCoMargin != "" {
		filter["min_buy_co_margin"] = minBuyCoMargin
	}
	if minSellCoMargin := query.Get("min_sell_co_margin"); minSellCoMargin != "" {
		filter["min_sell_co_margin"] = minSellCoMargin
	}

	// Sorting parameters
	if sort := query.Get("sort"); sort != "" {
		filter["sort"] = sort
	} else {
		// Default sort by MTF leverage if not specified
		filter["sort"] = "mtf"
	}

	if order := query.Get("order"); order != "" {
		filter["order"] = order
	} else {
		// Default to descending order for MTF
		filter["order"] = "desc"
	}

	// Get symbols
	symbols, count, err := h.symbolService.GetAllSymbols(filter, limit, offset)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Prepare response
	// Handle pagination calculation with zero limit protection
	page := 1
	pages := int64(1)
	pageSize := limit

	if limit > 0 {
		page = offset/limit + 1
		pages = (count + int64(limit) - 1) / int64(limit)
	} else if limit < 0 {
		// Use default values for invalid limit
		pageSize = 50 // Default page size
	}

	response := map[string]interface{}{
		"symbols":  symbols,
		"count":    count,
		"page":     page,
		"pageSize": pageSize,
		"pages":    pages,
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

// AutocompleteUnderlyingSymbols provides autocomplete suggestions for underlying symbols
func (h *SymbolHandler) AutocompleteUnderlyingSymbols(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	query := r.URL.Query()
	search := query.Get("search")

	// Get autocomplete suggestions
	suggestions, err := h.symbolService.GetUnderlyingSymbolSuggestions(search)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Prepare a more structured response with metadata
	response := map[string]interface{}{
		"suggestions": suggestions,
		"count":       len(suggestions),
		"query":       search,
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
