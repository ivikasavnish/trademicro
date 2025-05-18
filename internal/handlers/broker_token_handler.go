package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/vikasavnish/trademicro/internal/models"
	"github.com/vikasavnish/trademicro/internal/services"
)

// BrokerTokenHandler handles broker-related requests
type BrokerTokenHandler struct {
	brokerTokenService services.BrokerTokenService
	userService        services.UserService
	familyService      services.FamilyMemberService
}

func NewBrokerTokenHandler(
	brokerTokenService services.BrokerTokenService,
	userService services.UserService,
	familyService services.FamilyMemberService,
) *BrokerTokenHandler {
	return &BrokerTokenHandler{
		brokerTokenService: brokerTokenService,
		userService:        userService,
		familyService:      familyService,
	}
}

func (h *BrokerTokenHandler) RegisterRoutes(router *mux.Router) {
	// Broker management routes
	router.HandleFunc("/brokers", h.GetBrokers).Methods("GET")
	router.HandleFunc("/brokers/{id}", h.GetBroker).Methods("GET")
	router.HandleFunc("/brokers", h.CreateBroker).Methods("POST")
	router.HandleFunc("/brokers/{id}", h.UpdateBroker).Methods("PUT")
	router.HandleFunc("/brokers/{id}", h.DeleteBroker).Methods("DELETE")

	// Broker token routes for current user
	router.HandleFunc("/broker-tokens", h.GetUserBrokerTokens).Methods("GET")
	router.HandleFunc("/broker-tokens/{id}", h.GetBrokerToken).Methods("GET")
	router.HandleFunc("/broker-tokens", h.CreateBrokerToken).Methods("POST")
	router.HandleFunc("/broker-tokens/{id}", h.UpdateBrokerToken).Methods("PUT")
	router.HandleFunc("/broker-tokens/{id}", h.DeleteBrokerToken).Methods("DELETE")

	// Family member broker token routes
	router.HandleFunc("/family/{familyId}/broker-tokens", h.GetFamilyMemberBrokerTokens).Methods("GET")
	router.HandleFunc("/family/{familyId}/broker-tokens/{id}", h.GetFamilyMemberBrokerToken).Methods("GET")
	router.HandleFunc("/family/{familyId}/broker-tokens", h.CreateFamilyMemberBrokerToken).Methods("POST")
	router.HandleFunc("/family/{familyId}/broker-tokens/{id}", h.UpdateFamilyMemberBrokerToken).Methods("PUT")
	router.HandleFunc("/family/{familyId}/broker-tokens/{id}", h.DeleteFamilyMemberBrokerToken).Methods("DELETE")
}

// Helper to get user ID from context
func (h *BrokerTokenHandler) getUserIDFromContext(r *http.Request) (uint, error) {
	username, ok := r.Context().Value("username").(string)
	if !ok || username == "" {
		return 0, errors.New("unauthorized")
	}

	userID, err := h.userService.GetUserIDByUsername(username)
	if err != nil {
		return 0, errors.New("user not found")
	}

	return userID, nil
}

// BROKER MANAGEMENT ENDPOINTS

// GetBrokers returns a list of all available brokers
func (h *BrokerTokenHandler) GetBrokers(w http.ResponseWriter, r *http.Request) {
	brokers, err := h.brokerTokenService.GetAllBrokers()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(brokers)
}

// GetBroker returns a specific broker by ID
func (h *BrokerTokenHandler) GetBroker(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	broker, err := h.brokerTokenService.GetBrokerByID(uint(id))
	if err != nil {
		http.Error(w, "Broker not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(broker)
}

// CreateBroker creates a new broker (admin only)
func (h *BrokerTokenHandler) CreateBroker(w http.ResponseWriter, r *http.Request) {
	// Get user ID and verify admin role
	userID, err := h.getUserIDFromContext(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Check if user is admin
	isAdmin, err := h.userService.IsUserAdmin(userID)
	if err != nil || !isAdmin {
		http.Error(w, "Forbidden: Admin access required", http.StatusForbidden)
		return
	}

	var req models.BrokerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	broker := models.Broker{
		Name:        req.Name,
		Code:        req.Code,
		Description: req.Description,
		APIBaseURL:  req.APIBaseURL,
	}

	if req.IsActive != nil {
		broker.IsActive = *req.IsActive
	} else {
		broker.IsActive = true
	}

	createdBroker, err := h.brokerTokenService.CreateBroker(broker)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(createdBroker)
}

// UpdateBroker updates an existing broker (admin only)
func (h *BrokerTokenHandler) UpdateBroker(w http.ResponseWriter, r *http.Request) {
	// Get user ID and verify admin role
	userID, err := h.getUserIDFromContext(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Check if user is admin
	isAdmin, err := h.userService.IsUserAdmin(userID)
	if err != nil || !isAdmin {
		http.Error(w, "Forbidden: Admin access required", http.StatusForbidden)
		return
	}

	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	var req models.BrokerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	broker := models.Broker{
		ID:          uint(id),
		Name:        req.Name,
		Code:        req.Code,
		Description: req.Description,
		APIBaseURL:  req.APIBaseURL,
	}

	if req.IsActive != nil {
		broker.IsActive = *req.IsActive
	}

	updatedBroker, err := h.brokerTokenService.UpdateBroker(broker)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updatedBroker)
}

// DeleteBroker deletes a broker (admin only)
func (h *BrokerTokenHandler) DeleteBroker(w http.ResponseWriter, r *http.Request) {
	// Get user ID and verify admin role
	userID, err := h.getUserIDFromContext(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Check if user is admin
	isAdmin, err := h.userService.IsUserAdmin(userID)
	if err != nil || !isAdmin {
		http.Error(w, "Forbidden: Admin access required", http.StatusForbidden)
		return
	}

	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	if err := h.brokerTokenService.DeleteBroker(uint(id)); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Broker deleted successfully"})
}

// USER BROKER TOKEN ENDPOINTS

// GetUserBrokerTokens returns all broker tokens for the current user
func (h *BrokerTokenHandler) GetUserBrokerTokens(w http.ResponseWriter, r *http.Request) {
	userID, err := h.getUserIDFromContext(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	tokens, err := h.brokerTokenService.GetBrokerTokensByUserID(userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tokens)
}

// GetBrokerToken returns a specific broker token for the current user
func (h *BrokerTokenHandler) GetBrokerToken(w http.ResponseWriter, r *http.Request) {
	userID, err := h.getUserIDFromContext(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	token, err := h.brokerTokenService.GetBrokerTokenByID(uint(id))
	if err != nil {
		http.Error(w, "Broker token not found", http.StatusNotFound)
		return
	}

	// Verify ownership
	if token.UserID != userID {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(token)
}

// CreateBrokerToken creates a new broker token for the current user
func (h *BrokerTokenHandler) CreateBrokerToken(w http.ResponseWriter, r *http.Request) {
	userID, err := h.getUserIDFromContext(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req models.BrokerTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	token := models.BrokerToken{
		UserID:       userID,
		BrokerID:     req.BrokerID,
		ClientID:     req.ClientID,
		AccessToken:  req.AccessToken,
		RefreshToken: req.RefreshToken,
		ExpiresAt:    req.ExpiresAt,
		IsActive:     true,
	}

	createdToken, err := h.brokerTokenService.CreateBrokerToken(token)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(createdToken)
}

// UpdateBrokerToken updates an existing broker token for the current user
func (h *BrokerTokenHandler) UpdateBrokerToken(w http.ResponseWriter, r *http.Request) {
	userID, err := h.getUserIDFromContext(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	// Verify token ownership
	token, err := h.brokerTokenService.GetBrokerTokenByID(uint(id))
	if err != nil {
		http.Error(w, "Broker token not found", http.StatusNotFound)
		return
	}

	if token.UserID != userID {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req models.BrokerTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	token.BrokerID = req.BrokerID
	token.ClientID = req.ClientID
	token.AccessToken = req.AccessToken
	token.RefreshToken = req.RefreshToken

	if !req.ExpiresAt.IsZero() {
		token.ExpiresAt = req.ExpiresAt
	}

	updatedToken, err := h.brokerTokenService.UpdateBrokerToken(*token)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updatedToken)
}

// DeleteBrokerToken deletes a broker token
func (h *BrokerTokenHandler) DeleteBrokerToken(w http.ResponseWriter, r *http.Request) {
	userID, err := h.getUserIDFromContext(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	// Verify token ownership
	token, err := h.brokerTokenService.GetBrokerTokenByID(uint(id))
	if err != nil {
		http.Error(w, "Broker token not found", http.StatusNotFound)
		return
	}

	if token.UserID != userID {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if err := h.brokerTokenService.DeleteBrokerToken(uint(id)); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Broker token deleted successfully"})
}

// FAMILY MEMBER BROKER TOKEN ENDPOINTS

// GetFamilyMemberBrokerTokens returns all broker tokens for a family member
func (h *BrokerTokenHandler) GetFamilyMemberBrokerTokens(w http.ResponseWriter, r *http.Request) {
	userID, err := h.getUserIDFromContext(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	familyID, err := strconv.ParseUint(vars["familyId"], 10, 32)
	if err != nil {
		http.Error(w, "Invalid family member ID", http.StatusBadRequest)
		return
	}

	// Verify family member ownership
	member, err := h.familyService.GetFamilyMemberByID(uint(familyID))
	if err != nil {
		http.Error(w, "Family member not found", http.StatusNotFound)
		return
	}

	if member.UserID != userID {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	tokens, err := h.brokerTokenService.GetBrokerTokensByFamilyMemberID(uint(familyID))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tokens)
}

// GetFamilyMemberBrokerToken returns a specific broker token for a family member
func (h *BrokerTokenHandler) GetFamilyMemberBrokerToken(w http.ResponseWriter, r *http.Request) {
	userID, err := h.getUserIDFromContext(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	familyID, err := strconv.ParseUint(vars["familyId"], 10, 32)
	if err != nil {
		http.Error(w, "Invalid family member ID", http.StatusBadRequest)
		return
	}

	// Verify family member ownership
	member, err := h.familyService.GetFamilyMemberByID(uint(familyID))
	if err != nil {
		http.Error(w, "Family member not found", http.StatusNotFound)
		return
	}

	if member.UserID != userID {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	id, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		http.Error(w, "Invalid token ID", http.StatusBadRequest)
		return
	}

	token, err := h.brokerTokenService.GetBrokerTokenByID(uint(id))
	if err != nil {
		http.Error(w, "Broker token not found", http.StatusNotFound)
		return
	}

	// Verify the token belongs to this family member
	if token.FamilyMemberID == nil || *token.FamilyMemberID != uint(familyID) {
		http.Error(w, "Token not found for this family member", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(token)
}

// CreateFamilyMemberBrokerToken creates a new broker token for a family member
func (h *BrokerTokenHandler) CreateFamilyMemberBrokerToken(w http.ResponseWriter, r *http.Request) {
	userID, err := h.getUserIDFromContext(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	familyID, err := strconv.ParseUint(vars["familyId"], 10, 32)
	if err != nil {
		http.Error(w, "Invalid family member ID", http.StatusBadRequest)
		return
	}

	// Verify family member ownership
	member, err := h.familyService.GetFamilyMemberByID(uint(familyID))
	if err != nil {
		http.Error(w, "Family member not found", http.StatusNotFound)
		return
	}

	if member.UserID != userID {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req models.BrokerTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	familyMemberID := uint(familyID)
	token := models.BrokerToken{
		UserID:         userID,
		FamilyMemberID: &familyMemberID,
		BrokerID:       req.BrokerID,
		ClientID:       req.ClientID,
		AccessToken:    req.AccessToken,
		RefreshToken:   req.RefreshToken,
		ExpiresAt:      req.ExpiresAt,
		IsActive:       true,
	}

	createdToken, err := h.brokerTokenService.CreateBrokerToken(token)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(createdToken)
}

// UpdateFamilyMemberBrokerToken updates an existing broker token for a family member
func (h *BrokerTokenHandler) UpdateFamilyMemberBrokerToken(w http.ResponseWriter, r *http.Request) {
	userID, err := h.getUserIDFromContext(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	familyID, err := strconv.ParseUint(vars["familyId"], 10, 32)
	if err != nil {
		http.Error(w, "Invalid family member ID", http.StatusBadRequest)
		return
	}

	// Verify family member ownership
	member, err := h.familyService.GetFamilyMemberByID(uint(familyID))
	if err != nil {
		http.Error(w, "Family member not found", http.StatusNotFound)
		return
	}

	if member.UserID != userID {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	id, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		http.Error(w, "Invalid token ID", http.StatusBadRequest)
		return
	}

	// Verify token exists and belongs to this family member
	token, err := h.brokerTokenService.GetBrokerTokenByID(uint(id))
	if err != nil {
		http.Error(w, "Broker token not found", http.StatusNotFound)
		return
	}

	if token.FamilyMemberID == nil || *token.FamilyMemberID != uint(familyID) {
		http.Error(w, "Token not found for this family member", http.StatusNotFound)
		return
	}

	var req models.BrokerTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	token.BrokerID = req.BrokerID
	token.ClientID = req.ClientID
	token.AccessToken = req.AccessToken
	token.RefreshToken = req.RefreshToken

	if !req.ExpiresAt.IsZero() {
		token.ExpiresAt = req.ExpiresAt
	}

	updatedToken, err := h.brokerTokenService.UpdateBrokerToken(*token)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updatedToken)
}

// DeleteFamilyMemberBrokerToken deletes a broker token for a family member
func (h *BrokerTokenHandler) DeleteFamilyMemberBrokerToken(w http.ResponseWriter, r *http.Request) {
	userID, err := h.getUserIDFromContext(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	familyID, err := strconv.ParseUint(vars["familyId"], 10, 32)
	if err != nil {
		http.Error(w, "Invalid family member ID", http.StatusBadRequest)
		return
	}

	// Verify family member ownership
	member, err := h.familyService.GetFamilyMemberByID(uint(familyID))
	if err != nil {
		http.Error(w, "Family member not found", http.StatusNotFound)
		return
	}

	if member.UserID != userID {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	id, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		http.Error(w, "Invalid token ID", http.StatusBadRequest)
		return
	}

	// Verify token exists and belongs to this family member
	token, err := h.brokerTokenService.GetBrokerTokenByID(uint(id))
	if err != nil {
		http.Error(w, "Broker token not found", http.StatusNotFound)
		return
	}

	if token.FamilyMemberID == nil || *token.FamilyMemberID != uint(familyID) {
		http.Error(w, "Token not found for this family member", http.StatusNotFound)
		return
	}

	if err := h.brokerTokenService.DeleteBrokerToken(uint(id)); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Broker token deleted successfully"})
}
