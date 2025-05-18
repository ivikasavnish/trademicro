package handlers

import (
	"github.com/gorilla/mux"
	"github.com/vikasavnish/trademicro/internal/config"
)

type TaskHandler struct {
	config config.ServerConfig
}

func NewTaskHandler(config config.ServerConfig) *TaskHandler {
	return &TaskHandler{
		config: config,
	}
}

func (h *TaskHandler) RegisterRoutes(router *mux.Router) {
}
