package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/vikasavnish/trademicro/internal/services"
)

// InstrumentHandler handles HTTP requests for instrument operations
type InstrumentHandler struct {
	service *services.InstrumentService
}

// NewInstrumentHandler creates a new instrument handler
func NewInstrumentHandler(service *services.InstrumentService) *InstrumentHandler {
	return &InstrumentHandler{
		service: service,
	}
}

// FetchRequestOptions represents options for fetching instruments
type FetchRequestOptions struct {
	Mode            string `json:"mode"`            // "compact" or "detailed"
	ExchangeSegment string `json:"exchangeSegment"` // Optional segment to filter by
	SaveToFile      bool   `json:"saveToFile"`      // Whether to save the CSV to file
	BatchSize       int    `json:"batchSize"`       // Batch size for DB operations
}

// RegisterRoutes registers all instrument-related routes
func (h *InstrumentHandler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/instruments", h.GetInstrumentsHandler).Methods("GET")
	router.HandleFunc("/instruments/{id:[0-9]+}", h.GetInstrumentByIDHandler).Methods("GET")
	router.HandleFunc("/instruments/security/{securityId}", h.GetInstrumentBySecurityIDHandler).Methods("GET")
	router.HandleFunc("/instruments/fetch", h.FetchInstrumentsHandler).Methods("POST")
	router.HandleFunc("/instruments/exchanges", h.GetExchangesHandler).Methods("GET")
	router.HandleFunc("/instruments/exchange-segments", h.GetExchangeSegmentsHandler).Methods("GET")
	router.HandleFunc("/instruments/stats", h.GetInstrumentStatsHandler).Methods("GET")
}

// FetchInstrumentsHandler handles fetching and importing instruments from DhanHQ API
func (h *InstrumentHandler) FetchInstrumentsHandler(w http.ResponseWriter, r *http.Request) {
	// Parse request options
	var options FetchRequestOptions
	if r.Body != nil && r.ContentLength > 0 {
		if err := json.NewDecoder(r.Body).Decode(&options); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}

	// Set default values if not provided
	if options.Mode == "" {
		options.Mode = "compact"
	}
	if options.BatchSize <= 0 {
		options.BatchSize = 500
	}

	// Get username from context for logging
	username := "system"
	if ctxUsername, ok := r.Context().Value("username").(string); ok && ctxUsername != "" {
		username = ctxUsername
	}

	// Convert options to service format
	var fetchMode services.FetchMode
	switch options.Mode {
	case "detailed":
		fetchMode = services.DetailedMode
	default:
		fetchMode = services.CompactMode
	}

	fetchOptions := services.FetchOptions{
		Mode:            fetchMode,
		ExchangeSegment: options.ExchangeSegment,
		SaveToFile:      options.SaveToFile,
		BatchSize:       options.BatchSize,
		OutputFile:      "",
	}

	// Generate output file name if saving
	if options.SaveToFile {
		fetchOptions.OutputFile = "dhan_instruments_" +
			time.Now().Format("20060102_150405") + ".csv"
	}

	// Call service to fetch instruments
	result, err := h.service.FetchInstrumentsFromDhanHQ(fetchOptions)
	if err != nil {
		http.Error(w, "Failed to fetch instruments: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Add username to result
	type enhancedResult struct {
		*services.FetchResult
		User string `json:"user"`
	}

	enhancedResponse := enhancedResult{
		FetchResult: result,
		User:        username,
	}

	// Send response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(enhancedResponse)
}

// InstrumentQueryParams represents query parameters for fetching instruments
type InstrumentQueryParams struct {
	SecurityID       string
	Symbol           string
	ExchangeID       string
	Segment          string
	ISIN             string
	InstrumentType   string
	StrikePriceMin   float64
	StrikePriceMax   float64
	ExpiryDateStart  string
	ExpiryDateEnd    string
	OptionType       string
	UnderlyingSymbol string
	Series           string
	Page             int
	Limit            int
}

// GetInstrumentsHandler handles fetching instruments with filters
func (h *InstrumentHandler) GetInstrumentsHandler(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	query := r.URL.Query()

	params := InstrumentQueryParams{
		SecurityID:       query.Get("securityId"),
		Symbol:           query.Get("symbol"),
		ExchangeID:       query.Get("exchangeId"),
		Segment:          query.Get("segment"),
		ISIN:             query.Get("isin"),
		InstrumentType:   query.Get("instrumentType"),
		OptionType:       query.Get("optionType"),
		UnderlyingSymbol: query.Get("underlyingSymbol"),
		Series:           query.Get("series"),
	}

	// Parse numeric parameters
	if val := query.Get("strikePriceMin"); val != "" {
		if f, err := strconv.ParseFloat(val, 64); err == nil {
			params.StrikePriceMin = f
		}
	}

	if val := query.Get("strikePriceMax"); val != "" {
		if f, err := strconv.ParseFloat(val, 64); err == nil {
			params.StrikePriceMax = f
		}
	}

	if val := query.Get("page"); val != "" {
		if i, err := strconv.Atoi(val); err == nil && i > 0 {
			params.Page = i
		}
	}

	if val := query.Get("limit"); val != "" {
		if i, err := strconv.Atoi(val); err == nil && i > 0 {
			params.Limit = i
		}
	} else {
		params.Limit = 50 // Default limit
	}

	// Set pagination
	offset := 0
	if params.Page > 0 {
		offset = (params.Page - 1) * params.Limit
	}

	// Parse date parameters
	var expiryDateStart, expiryDateEnd time.Time
	if val := query.Get("expiryDateStart"); val != "" {
		if t, err := time.Parse("2006-01-02", val); err == nil {
			expiryDateStart = t
		}
	}

	if val := query.Get("expiryDateEnd"); val != "" {
		if t, err := time.Parse("2006-01-02", val); err == nil {
			expiryDateEnd = t
		}
	}

	// Build service query
	serviceQuery := services.InstrumentQuery{
		SecurityID:       params.SecurityID,
		Symbol:           params.Symbol,
		ExchangeID:       params.ExchangeID,
		Segment:          params.Segment,
		ISIN:             params.ISIN,
		InstrumentType:   params.InstrumentType,
		StrikePriceMin:   params.StrikePriceMin,
		StrikePriceMax:   params.StrikePriceMax,
		ExpiryDateStart:  expiryDateStart,
		ExpiryDateEnd:    expiryDateEnd,
		OptionType:       params.OptionType,
		UnderlyingSymbol: params.UnderlyingSymbol,
		Series:           params.Series,
		Limit:            params.Limit,
		Offset:           offset,
	}

	// Call service to get instruments
	instruments, count, err := h.service.GetInstruments(serviceQuery)
	if err != nil {
		http.Error(w, "Failed to fetch instruments: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Build response with pagination
	type PaginatedResponse struct {
		Data       interface{} `json:"data"`
		TotalCount int64       `json:"totalCount"`
		Page       int         `json:"page"`
		Limit      int         `json:"limit"`
		TotalPages int         `json:"totalPages"`
	}

	totalPages := (int(count) + params.Limit - 1) / params.Limit
	response := PaginatedResponse{
		Data:       instruments,
		TotalCount: count,
		Page:       params.Page,
		Limit:      params.Limit,
		TotalPages: totalPages,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetInstrumentByIDHandler handles fetching a single instrument by ID
func (h *InstrumentHandler) GetInstrumentByIDHandler(w http.ResponseWriter, r *http.Request) {
	// Parse ID from URL
	vars := mux.Vars(r)
	idStr := vars["id"]

	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		http.Error(w, "Invalid instrument ID", http.StatusBadRequest)
		return
	}

	// Call service to get instrument
	instrument, err := h.service.GetInstrumentByID(uint(id))
	if err != nil {
		http.Error(w, "Failed to fetch instrument: "+err.Error(), http.StatusNotFound)
		return
	}

	// Send response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(instrument)
}

// GetInstrumentBySecurityIDHandler handles fetching a single instrument by security ID
func (h *InstrumentHandler) GetInstrumentBySecurityIDHandler(w http.ResponseWriter, r *http.Request) {
	// Parse security ID from URL
	vars := mux.Vars(r)
	securityID := vars["securityId"]

	// Call service to get instrument
	instrument, err := h.service.GetInstrumentBySecurityID(securityID)
	if err != nil {
		http.Error(w, "Failed to fetch instrument: "+err.Error(), http.StatusNotFound)
		return
	}

	// Send response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(instrument)
}

// GetExchangesHandler returns a list of available exchanges
func (h *InstrumentHandler) GetExchangesHandler(w http.ResponseWriter, r *http.Request) {
	// Define exchanges based on DhanHQ documentation
	exchanges := []map[string]string{
		{"id": "NSE", "name": "National Stock Exchange"},
		{"id": "BSE", "name": "Bombay Stock Exchange"},
		{"id": "MCX", "name": "Multi Commodity Exchange"},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(exchanges)
}

// GetExchangeSegmentsHandler returns a list of exchange segments
func (h *InstrumentHandler) GetExchangeSegmentsHandler(w http.ResponseWriter, r *http.Request) {
	// Define segments based on DhanHQ documentation
	segments := []map[string]string{
		{"id": "NSE_EQ", "name": "NSE Equity"},
		{"id": "BSE_EQ", "name": "BSE Equity"},
		{"id": "NSE_FNO", "name": "NSE Futures & Options"},
		{"id": "BSE_FNO", "name": "BSE Futures & Options"},
		{"id": "MCX_COMM", "name": "MCX Commodities"},
		{"id": "NSE_CURRENCY", "name": "NSE Currency"},
		{"id": "BSE_CURRENCY", "name": "BSE Currency"},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(segments)
}

// GetInstrumentStatsHandler returns statistics about instruments in the database
func (h *InstrumentHandler) GetInstrumentStatsHandler(w http.ResponseWriter, r *http.Request) {
	// This would connect to the service to get stats about instruments
	// For now, returning a placeholder
	stats := map[string]interface{}{
		"totalCount":  0,
		"byExchange":  map[string]int{},
		"bySegment":   map[string]int{},
		"lastUpdated": time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}
