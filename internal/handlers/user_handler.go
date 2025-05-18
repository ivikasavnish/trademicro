package handlers

import (
	"github.com/gorilla/mux"
	"github.com/vikasavnish/trademicro/internal/services"
)

// UserHandler handles user-related requests
type UserHandler struct {
	userService services.UserService
}

func NewUserHandler(userService services.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

func (h *UserHandler) RegisterRoutes(router *mux.Router) {
	// router.HandleFunc("/users", h.GetUsers).Methods("GET")
	// router.HandleFunc("/users/{id}", h.GetUser).Methods("GET")
	// router.HandleFunc("/users", h.CreateUser).Methods("POST")
	// router.HandleFunc("/users/{id}", h.UpdateUser).Methods("PUT")
	// router.HandleFunc("/users/{id}", h.DeleteUser).Methods("DELETE")
}
