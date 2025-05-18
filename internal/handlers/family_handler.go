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

type FamilyMemberHandler struct {
	familyService services.FamilyMemberService
	userService   services.UserService
}

func NewFamilyMemberHandler(familyService services.FamilyMemberService, userService services.UserService) *FamilyMemberHandler {
	return &FamilyMemberHandler{
		familyService: familyService,
		userService:   userService,
	}
}

func (h *FamilyMemberHandler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/family", h.GetFamilyMembers).Methods("GET")
	router.HandleFunc("/family/{id:[0-9]+}", h.GetFamilyMember).Methods("GET")
	router.HandleFunc("/family", h.CreateFamilyMember).Methods("POST")
	router.HandleFunc("/family/{id:[0-9]+}", h.UpdateFamilyMember).Methods("PUT")
	router.HandleFunc("/family/{id:[0-9]+}", h.DeleteFamilyMember).Methods("DELETE")
	router.HandleFunc("/family/{id:[0-9]+}/status", h.ToggleFamilyMemberStatus).Methods("PATCH")
	router.HandleFunc("/family/{id:[0-9]+}/broker", h.UpdateBrokerInfo).Methods("PATCH")
}

// GetFamilyMembers retrieves all family members for the authenticated user
func (h *FamilyMemberHandler) GetFamilyMembers(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context
	userID, err := h.getUserIDFromContext(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get family members
	members, err := h.familyService.GetFamilyMembersByUserID(userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(members)
}

// GetFamilyMember retrieves a single family member by ID
func (h *FamilyMemberHandler) GetFamilyMember(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context
	userID, err := h.getUserIDFromContext(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get member ID from URL
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	// Get family member
	member, err := h.familyService.GetFamilyMemberByID(uint(id))
	if err != nil {
		http.Error(w, "Family member not found", http.StatusNotFound)
		return
	}

	// Check if member belongs to the user
	if member.UserID != userID {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(member)
}

// CreateFamilyMember creates a new family member
func (h *FamilyMemberHandler) CreateFamilyMember(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context
	userID, err := h.getUserIDFromContext(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse request body
	var member models.FamilyMember
	if err := json.NewDecoder(r.Body).Decode(&member); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Set user ID
	member.UserID = userID

	// Create family member
	createdMember, err := h.familyService.CreateFamilyMember(member)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(createdMember)
}

// UpdateFamilyMember updates an existing family member
func (h *FamilyMemberHandler) UpdateFamilyMember(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context
	userID, err := h.getUserIDFromContext(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get member ID from URL
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	// Parse request body
	var member models.FamilyMember
	if err := json.NewDecoder(r.Body).Decode(&member); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Update family member
	updatedMember, err := h.familyService.UpdateFamilyMember(uint(id), userID, member)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updatedMember)
}

// DeleteFamilyMember deletes a family member
func (h *FamilyMemberHandler) DeleteFamilyMember(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context
	userID, err := h.getUserIDFromContext(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get member ID from URL
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	// Delete family member
	if err := h.familyService.DeleteFamilyMember(uint(id), userID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Family member deleted successfully"})
}

// ToggleFamilyMemberStatus activates or deactivates a family member
func (h *FamilyMemberHandler) ToggleFamilyMemberStatus(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context
	userID, err := h.getUserIDFromContext(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get member ID from URL
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	// Parse request body
	var req struct {
		IsActive bool `json:"isActive"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Get existing member to verify ownership
	member, err := h.familyService.GetFamilyMemberByID(uint(id))
	if err != nil {
		http.Error(w, "Family member not found", http.StatusNotFound)
		return
	}

	// Check if member belongs to the user
	if member.UserID != userID {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Update active status
	member.IsActive = req.IsActive
	updatedMember, err := h.familyService.UpdateFamilyMember(uint(id), userID, member)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updatedMember)
}

// UpdateBrokerInfo updates the broker information for a family member
func (h *FamilyMemberHandler) UpdateBrokerInfo(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context
	userID, err := h.getUserIDFromContext(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get member ID from URL
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	// Parse request body
	var req struct {
		ClientID    string `json:"clientId"`
		ClientToken string `json:"clientToken"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Get existing member to verify ownership
	member, err := h.familyService.GetFamilyMemberByID(uint(id))
	if err != nil {
		http.Error(w, "Family member not found", http.StatusNotFound)
		return
	}

	// Check if member belongs to the user
	if member.UserID != userID {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Update broker info
	member.ClientID = req.ClientID
	member.ClientToken = req.ClientToken
	updatedMember, err := h.familyService.UpdateFamilyMember(uint(id), userID, member)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updatedMember)
}

// Helper to get user ID from context
func (h *FamilyMemberHandler) getUserIDFromContext(r *http.Request) (uint, error) {
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
