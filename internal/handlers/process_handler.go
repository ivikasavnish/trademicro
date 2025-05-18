package handlers

import (
	"github.com/gorilla/mux"
	"github.com/vikasavnish/trademicro/internal/services"
)

type ProcessHandler struct {
	processService services.ProcessService
}

func NewProcessHandler(processService services.ProcessService) *ProcessHandler {
	return &ProcessHandler{
		processService: processService,
	}
}
func (h *ProcessHandler) RegisterRoutes(router *mux.Router) {
	// router.HandleFunc("/users", h.GetUsers).Methods("GET")
	// router.HandleFunc("/users/{id}", h.GetUser).Methods("GET")
	// router.HandleFunc("/users", h.CreateUser).Methods("POST")
	// router.HandleFunc("/users/{id}", h.UpdateUser).Methods("PUT")
	// router.HandleFunc("/users/{id}", h.DeleteUser).Methods("DELETE")
}
