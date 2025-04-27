package api

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/vikasavnish/trademicro/scripts"
	"github.com/gorilla/mux"
)

// TaskRequest represents a request to create a new task
type TaskRequest struct {
	Name        string                 `json:"name"`
	Command     string                 `json:"command"`
	Args        []string               `json:"args"`
	Params      map[string]interface{} `json:"params"`
	RunOnWorker bool                   `json:"run_on_worker"`
}

// TaskHandler handles task-related API endpoints
type TaskHandler struct {
	taskManager *scripts.TaskManager
}

// NewTaskHandler creates a new task handler
func NewTaskHandler() *TaskHandler {
	// Get worker configuration from environment variables
	workerHost := os.Getenv("WORKER_HOST")
	workerUser := os.Getenv("WORKER_USER")
	privateKeyPath := os.Getenv("WORKER_SSH_KEY")
	logDir := os.Getenv("TASK_LOG_DIR")

	if workerHost == "" {
		workerHost = "instance-20250416-112838" // Default to the n1-highcpu-4 instance
	}

	if workerUser == "" {
		workerUser = "root" // Default user
	}

	if privateKeyPath == "" {
		// Default to SSH key in the app directory
		privateKeyPath = "/opt/trademicro/.ssh/worker_key"
	}

	if logDir == "" {
		logDir = "/opt/trademicro/logs/tasks"
	}

	// Ensure the log directory exists
	os.MkdirAll(logDir, 0755)

	return &TaskHandler{
		taskManager: scripts.NewTaskManager(workerHost, workerUser, privateKeyPath, logDir),
	}
}

// RegisterRoutes registers the task handler routes
func (h *TaskHandler) RegisterRoutes(r *mux.Router) {
	r.HandleFunc("/api/tasks", h.CreateTask).Methods("POST")
	r.HandleFunc("/api/tasks", h.ListTasks).Methods("GET")
	r.HandleFunc("/api/tasks/{id}", h.GetTask).Methods("GET")
	r.HandleFunc("/api/tasks/{id}/start", h.StartTask).Methods("POST")
	r.HandleFunc("/api/tasks/{id}/cancel", h.CancelTask).Methods("POST")
	r.HandleFunc("/api/tasks/{id}/logs", h.GetTaskLogs).Methods("GET")
}

// CreateTask handles the creation of a new task
func (h *TaskHandler) CreateTask(w http.ResponseWriter, r *http.Request) {
	var req TaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Name == "" || req.Command == "" {
		http.Error(w, "Name and command are required", http.StatusBadRequest)
		return
	}

	task, err := h.taskManager.CreateTask(req.Name, req.Command, req.Args, req.Params, req.RunOnWorker)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(task)
}

// ListTasks handles listing all tasks
func (h *TaskHandler) ListTasks(w http.ResponseWriter, r *http.Request) {
	tasks := h.taskManager.GetAllTasks()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tasks)
}

// GetTask handles retrieving a task by ID
func (h *TaskHandler) GetTask(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	taskID := vars["id"]

	task, err := h.taskManager.GetTask(taskID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(task)
}

// StartTask handles starting a task
func (h *TaskHandler) StartTask(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	taskID := vars["id"]

	if err := h.taskManager.RunTask(taskID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	task, _ := h.taskManager.GetTask(taskID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(task)
}

// CancelTask handles cancelling a running task
func (h *TaskHandler) CancelTask(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	taskID := vars["id"]

	if err := h.taskManager.CancelTask(taskID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	task, _ := h.taskManager.GetTask(taskID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(task)
}

// GetTaskLogs handles retrieving logs for a task
func (h *TaskHandler) GetTaskLogs(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	taskID := vars["id"]

	task, err := h.taskManager.GetTask(taskID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	// Get log file paths
	logDir := os.Getenv("TASK_LOG_DIR")
	if logDir == "" {
		logDir = "/opt/trademicro/logs/tasks"
	}

	logFile := filepath.Join(logDir, taskID+".log")
	errFile := filepath.Join(logDir, taskID+".err")

	// Check if log files exist
	if _, err := os.Stat(logFile); os.IsNotExist(err) {
		http.Error(w, "Log file not found", http.StatusNotFound)
		return
	}

	// Read log files
	stdout, _ := os.ReadFile(logFile)
	stderr, _ := os.ReadFile(errFile)

	// Create response
	response := struct {
		Task   *scripts.Task `json:"task"`
		Stdout string        `json:"stdout"`
		Stderr string        `json:"stderr"`
	}{
		Task:   task,
		Stdout: string(stdout),
		Stderr: string(stderr),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// InitTaskCleanup starts a goroutine to periodically clean up old tasks
func (h *TaskHandler) InitTaskCleanup() {
	go func() {
		ticker := time.NewTicker(6 * time.Hour)
		defer ticker.Stop()

		for range ticker.C {
			// Clean up tasks older than 7 days
			h.taskManager.CleanupOldTasks(7 * 24 * time.Hour)
		}
	}()
}
