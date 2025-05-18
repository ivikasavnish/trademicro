package tasks

import (
	"log"
	"time"

	"gorm.io/gorm"

	"github.com/vikasavnish/trademicro/internal/websocket"
)

// Manager handles the execution of scheduled tasks
type Manager struct {
	db    *gorm.DB
	wsHub *websocket.Hub
	tasks []Task
}

// Task represents a scheduled task that needs to be executed
type Task interface {
	Start()
	Stop()
}

// NewManager creates a new task manager
func NewManager(db *gorm.DB, wsHub *websocket.Hub) *Manager {
	return &Manager{
		db:    db,
		wsHub: wsHub,
		tasks: make([]Task, 0),
	}
}

// RegisterTask registers a task with the manager
func (m *Manager) RegisterTask(task Task) {
	m.tasks = append(m.tasks, task)
}

// StartScheduledTasks starts all registered tasks
func (m *Manager) StartScheduledTasks() {
	// Register symbol update task
	symbolUpdateTask := NewSymbolUpdateTask(m.db, m.wsHub)
	m.RegisterTask(symbolUpdateTask)

	// Start all registered tasks
	for _, task := range m.tasks {
		go task.Start()
	}

	log.Println("Started all scheduled tasks")
}

// StopAllTasks stops all running tasks
func (m *Manager) StopAllTasks() {
	for _, task := range m.tasks {
		task.Stop()
	}
	log.Println("Stopped all scheduled tasks")
}

// SymbolUpdateTask handles updating symbols on a schedule
type SymbolUpdateTask struct {
	db        *gorm.DB
	wsHub     *websocket.Hub
	stopChan  chan struct{}
	isRunning bool
}

// NewSymbolUpdateTask creates a new symbol update task
func NewSymbolUpdateTask(db *gorm.DB, wsHub *websocket.Hub) *SymbolUpdateTask {
	return &SymbolUpdateTask{
		db:        db,
		wsHub:     wsHub,
		stopChan:  make(chan struct{}),
		isRunning: false,
	}
}

// Start begins the symbol update task
func (t *SymbolUpdateTask) Start() {
	if t.isRunning {
		return
	}

	t.isRunning = true
	ticker := time.NewTicker(24 * time.Hour) // Run once per day

	// Run immediately on start
	go t.updateSymbols()

	go func() {
		for {
			select {
			case <-ticker.C:
				t.updateSymbols()
			case <-t.stopChan:
				ticker.Stop()
				t.isRunning = false
				return
			}
		}
	}()

	log.Println("Symbol update task started")
}

// Stop terminates the symbol update task
func (t *SymbolUpdateTask) Stop() {
	if !t.isRunning {
		return
	}

	close(t.stopChan)
	log.Println("Symbol update task stopped")
}

// updateSymbols fetches and updates symbols in the database
func (t *SymbolUpdateTask) updateSymbols() {
	log.Println("Running scheduled symbol update")

	// Here you would add the actual implementation for updating symbols
	// This might involve calling an external API, processing CSV data, etc.

	log.Println("Symbol update completed")
}
