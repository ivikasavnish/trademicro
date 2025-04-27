package scripts

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// TaskStatus represents the status of a task
type TaskStatus string

const (
	TaskPending   TaskStatus = "pending"
	TaskRunning   TaskStatus = "running"
	TaskCompleted TaskStatus = "completed"
	TaskFailed    TaskStatus = "failed"
)

// Task represents a long-running task
type Task struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Command     string                 `json:"command"`
	Args        []string               `json:"args"`
	Status      TaskStatus             `json:"status"`
	StartTime   time.Time              `json:"start_time"`
	EndTime     time.Time              `json:"end_time"`
	Output      string                 `json:"output"`
	Error       string                 `json:"error"`
	ExitCode    int                    `json:"exit_code"`
	Params      map[string]interface{} `json:"params"`
	RunOnWorker bool                   `json:"run_on_worker"`
}

// TaskManager manages long-running tasks
type TaskManager struct {
	tasks       map[string]*Task
	mu          sync.RWMutex
	workerHost  string
	workerUser  string
	privateKey  string
	logDir      string
	taskTimeout time.Duration
}

// NewTaskManager creates a new task manager
func NewTaskManager(workerHost, workerUser, privateKey, logDir string) *TaskManager {
	// Create log directory if it doesn't exist
	if _, err := os.Stat(logDir); os.IsNotExist(err) {
		log.Printf("Creating task log directory: %s", logDir)
		os.MkdirAll(logDir, 0755)
	}

	return &TaskManager{
		tasks:       make(map[string]*Task),
		workerHost:  workerHost,
		workerUser:  workerUser,
		privateKey:  privateKey,
		logDir:      logDir,
		taskTimeout: 24 * time.Hour, // Default timeout of 24 hours
	}
}

// CreateTask creates a new task
func (tm *TaskManager) CreateTask(name, command string, args []string, params map[string]interface{}, runOnWorker bool) (*Task, error) {
	taskID := fmt.Sprintf("%s-%d", name, time.Now().Unix())
	
	task := &Task{
		ID:          taskID,
		Name:        name,
		Command:     command,
		Args:        args,
		Status:      TaskPending,
		StartTime:   time.Time{},
		EndTime:     time.Time{},
		Output:      "",
		Error:       "",
		ExitCode:    -1,
		Params:      params,
		RunOnWorker: runOnWorker,
	}
	
	tm.mu.Lock()
	tm.tasks[taskID] = task
	tm.mu.Unlock()
	
	return task, nil
}

// RunTask runs a task either locally or on the worker instance
func (tm *TaskManager) RunTask(taskID string) error {
	tm.mu.Lock()
	task, exists := tm.tasks[taskID]
	if !exists {
		tm.mu.Unlock()
		return fmt.Errorf("task %s not found", taskID)
	}
	
	task.Status = TaskRunning
	task.StartTime = time.Now()
	tm.mu.Unlock()
	
	// Create log files
	logFile := filepath.Join(tm.logDir, taskID+".log")
	errFile := filepath.Join(tm.logDir, taskID+".err")
	
	outputFile, err := os.Create(logFile)
	if err != nil {
		return fmt.Errorf("failed to create output log file: %v", err)
	}
	defer outputFile.Close()
	
	errorFile, err := os.Create(errFile)
	if err != nil {
		return fmt.Errorf("failed to create error log file: %v", err)
	}
	defer errorFile.Close()
	
	// Run the task asynchronously
	go func() {
		var cmd *exec.Cmd
		var stdout, stderr bytes.Buffer
		
		if task.RunOnWorker {
			// Run on worker instance via SSH
			sshArgs := []string{
				"-i", tm.privateKey,
				"-o", "StrictHostKeyChecking=no",
				fmt.Sprintf("%s@%s", tm.workerUser, tm.workerHost),
			}
			
			// Construct the command to run on the remote server
			remoteCmd := task.Command
			if len(task.Args) > 0 {
				remoteCmd += " " + strings.Join(task.Args, " ")
			}
			
			// Add the remote command to SSH arguments
			sshArgs = append(sshArgs, remoteCmd)
			
			// Create SSH command
			cmd = exec.Command("ssh", sshArgs...)
		} else {
			// Run locally
			cmd = exec.Command(task.Command, task.Args...)
		}
		
		// Set up pipes for stdout and stderr
		cmd.Stdout = io.MultiWriter(outputFile, &stdout)
		cmd.Stderr = io.MultiWriter(errorFile, &stderr)
		
		// Run the command
		err := cmd.Run()
		
		// Update task status
		tm.mu.Lock()
		defer tm.mu.Unlock()
		
		task.EndTime = time.Now()
		task.Output = stdout.String()
		task.Error = stderr.String()
		
		if err != nil {
			task.Status = TaskFailed
			if exitErr, ok := err.(*exec.ExitError); ok {
				task.ExitCode = exitErr.ExitCode()
			} else {
				task.ExitCode = -1
			}
		} else {
			task.Status = TaskCompleted
			task.ExitCode = 0
		}
		
		// Save task result to JSON file
		resultFile := filepath.Join(tm.logDir, taskID+".json")
		resultData, _ := json.MarshalIndent(task, "", "  ")
		os.WriteFile(resultFile, resultData, 0644)
	}()
	
	return nil
}

// GetTask returns a task by ID
func (tm *TaskManager) GetTask(taskID string) (*Task, error) {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	
	task, exists := tm.tasks[taskID]
	if !exists {
		return nil, fmt.Errorf("task %s not found", taskID)
	}
	
	return task, nil
}

// GetAllTasks returns all tasks
func (tm *TaskManager) GetAllTasks() []*Task {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	
	tasks := make([]*Task, 0, len(tm.tasks))
	for _, task := range tm.tasks {
		tasks = append(tasks, task)
	}
	
	return tasks
}

// CancelTask cancels a running task
func (tm *TaskManager) CancelTask(taskID string) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	
	task, exists := tm.tasks[taskID]
	if !exists {
		return fmt.Errorf("task %s not found", taskID)
	}
	
	if task.Status != TaskRunning {
		return fmt.Errorf("task %s is not running", taskID)
	}
	
	// If running on worker, use SSH to kill the process
	if task.RunOnWorker {
		// Find and kill the process on the worker
		findCmd := fmt.Sprintf("ps aux | grep '%s' | grep -v grep | awk '{print $2}'", task.Command)
		killCmd := fmt.Sprintf("kill -9 $(%s)", findCmd)
		
		sshArgs := []string{
			"-i", tm.privateKey,
			"-o", "StrictHostKeyChecking=no",
			fmt.Sprintf("%s@%s", tm.workerUser, tm.workerHost),
			killCmd,
		}
		
		exec.Command("ssh", sshArgs...).Run()
	}
	
	task.Status = TaskFailed
	task.EndTime = time.Now()
	task.Error = "Task was cancelled"
	
	return nil
}

// CleanupOldTasks removes completed tasks older than the specified duration
func (tm *TaskManager) CleanupOldTasks(age time.Duration) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	
	now := time.Now()
	for id, task := range tm.tasks {
		if (task.Status == TaskCompleted || task.Status == TaskFailed) &&
			!task.EndTime.IsZero() &&
			now.Sub(task.EndTime) > age {
			delete(tm.tasks, id)
		}
	}
}
